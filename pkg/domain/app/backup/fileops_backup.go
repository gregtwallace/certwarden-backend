package backup

import (
	"archive/zip"
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const dataStorageBackupDirName = "backup"
const backupFileMode = 0o600

const internalBackupFile = "backup.zip"
const internalBackupHashFile = internalBackupFile + ".sha1"

// createDataBackup creates a backup of the parent app's data root with optional
// exclusion for the backup subdirectories. It returns a zip file of the backup
// or an error if it failed.
func (service *Service) createDataBackup(withOnDiskBackups bool) (zipFileBytes []byte, err error) {
	// lock sql and defer unlock
	unlock, err := service.lockSQLForBackup()
	if err != nil {
		return nil, err
	}
	defer unlock()

	// make buffer, hasher, and writer for internal backup zip
	var internalZipBuffer bytes.Buffer
	internalZipHasher := sha1.New()
	internalZipWriter := zip.NewWriter(io.MultiWriter(&internalZipBuffer, internalZipHasher))

	// walker function to and add to zip, preserving file path
	zipWalker := func(path string, info fs.FileInfo, err error) error {
		// ensure err is passed to the top
		if err != nil {
			return err
		}

		// if folder, return err if skipping folder, else return nil
		// and walker will get to it in a different iteration
		if info.IsDir() {
			if !withOnDiskBackups && path == service.cleanDataStorageBackupPath {
				return filepath.SkipDir
			}

			return nil
		}

		// this is a file, zip it and a hash of it
		f, err := os.Open(path)
		if err != nil {
			return errorFileError(path, "failed to open file")
		}
		defer f.Close()

		// create file in zip (trim root prefix off so path in zip matches data root)
		zipFileInternalName := strings.TrimPrefix(path, service.cleanDataStorageRootPath+string(filepath.Separator))
		zipFile, err := internalZipWriter.Create(zipFileInternalName)
		if err != nil {
			return errorFileError(path, "failed to create file")
		}

		// copy file to zip file
		// _, err = io.Copy(zipFile, fileDataWithHasher)
		_, err = io.Copy(zipFile, f)
		if err != nil {
			return errorFileError(path, "failed to copy file")
		}

		// unlock

		return nil
	}

	// walk root dir
	err = filepath.Walk(service.cleanDataStorageRootPath, zipWalker)
	if err != nil {
		service.logger.Errorf("failed to make backup (%s)", err)
		return nil, err
	}

	// close zip writer (note: Close() writes the gzip footer and cannot be deferred)
	err = internalZipWriter.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close zip.writer (%w)", err)
	}

	// create wrapper zip that contains the hashed backup and the hash
	// file itself
	var wrapperZipBuffer bytes.Buffer
	wrapperZipWriter := zip.NewWriter(&wrapperZipBuffer)

	// write internal backup zip
	zipFile, err := wrapperZipWriter.Create(internalBackupFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create internal backup zip in wrapper zip (%w)", err)
	}

	// copy internal backup zip into wrapper
	_, err = io.Copy(zipFile, &internalZipBuffer)
	if err != nil {
		return nil, fmt.Errorf("failed to copy internal backup zip into wrapper zip (%w)", err)
	}

	// create hash file in wrapper zip
	zipFileHashFile, err := wrapperZipWriter.Create(internalBackupHashFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create internal backup zip hash in wrapper zip (%w)", err)
	}

	// write hash (as hex string) file in wrapper zip
	_, err = fmt.Fprintf(zipFileHashFile, "%x", internalZipHasher.Sum(nil))
	if err != nil {
		return nil, fmt.Errorf("failed to copy internal backup hash into wrapper zip (%w)", err)
	}

	// close wrapper zip writer (note: Close() writes the gzip footer and cannot be deferred)
	err = wrapperZipWriter.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close wrapper zip.writer (%w)", err)
	}

	return wrapperZipBuffer.Bytes(), nil
}

// CreateBackupOnDisk backs up the app state and saves it to the local backup folder. It
// optionally includes log files but never includes on disk backups.
func (service *Service) CreateBackupOnDisk() (backupFileDetails, error) {
	// make backup
	zipFileData, err := service.createDataBackup(false)
	if err != nil {
		return backupFileDetails{}, err
	}

	// save locally
	fileName, createdAt := makeBackupZipFileName()
	fileNameWithPath := service.cleanDataStorageBackupPath + "/" + fileName
	err = os.WriteFile(fileNameWithPath, zipFileData, backupFileMode)
	if err != nil {
		return backupFileDetails{}, fmt.Errorf("could not write backup file to disk (%w)", err)
	}

	service.logger.Infof("backup saved to disk (%s)", fileName)

	// only try to delete if retention config is set
	if service.config != nil && service.config.Retention.MaxCount != nil {
		err = service.deleteCountGreaterThan(*service.config.Retention.MaxCount)
		if err != nil {
			service.logger.Errorf("failed to delete backups over retention count (%s)", err)
		}
	}

	// return info about new file
	return backupFileDetails{
		Name:      fileName,
		Size:      len(zipFileData),
		ModTime:   createdAt, // not always 100% exact, but close enough
		CreatedAt: &createdAt,
	}, nil
}
