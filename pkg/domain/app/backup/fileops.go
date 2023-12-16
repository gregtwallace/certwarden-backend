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
const backupFileMode = 0600

const internalBackupFile = "backup.zip"
const internalBackupHashFile = internalBackupFile + ".sha1"

// createDataBackup creates a backup of the parent app's data root with optional
// exclusion for the backup subdirectories. It returns a zip file of the backup
// or an error if it failed.
func (service *Service) createDataBackup(withOnDiskBackups bool) (zipFileBytes []byte, err error) {
	// make buffer, hasher, and writer for internal backup zip
	internalZipBuffer := new(bytes.Buffer)
	internalZipHasher := sha1.New()
	internalZipWriter := zip.NewWriter(io.MultiWriter(internalZipBuffer, internalZipHasher))

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
			return fmt.Errorf("failed to open file %s for data backup (%s)", path, err)
		}
		defer f.Close()

		// make hasher
		// hasher := sha1.New()
		// fileDataWithHasher := io.TeeReader(f, hasher)

		// create file in zip (trim root prefix off so path in zip matches data root)
		zipFileInternalName := strings.TrimPrefix(path, service.cleanDataStorageRootPath+string(filepath.Separator))
		zipFile, err := internalZipWriter.Create(zipFileInternalName)
		if err != nil {
			return fmt.Errorf("failed to create file %s for data backup (%s)", path, err)
		}

		// copy file to zip file
		// _, err = io.Copy(zipFile, fileDataWithHasher)
		_, err = io.Copy(zipFile, f)
		if err != nil {
			return fmt.Errorf("failed to copy file %s into data backup (%s)", path, err)
		}

		// only write hashes for .yaml and .db files
		// fExt := filepath.Ext(path)
		// if fExt == ".yaml" || fExt == ".db" {
		// 	// create hash file and write it as well
		// 	zipFileHashFile, err := zipWriter.Create(zipFileInternalName + ".sha1")
		// 	if err != nil {
		// 		return fmt.Errorf("failed to make file %s hash file for data backup (%s)", path, err)
		// 	}

		// 	// copy hash (as hex string) to hash file in zip file
		// 	_, err = io.WriteString(zipFileHashFile, fmt.Sprintf("%x", hasher.Sum(nil)))
		// 	if err != nil {
		// 		return fmt.Errorf("failed to copy file %s into data backup (%s)", path, err)
		// 	}
		// }

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
		return nil, fmt.Errorf("failed to close zip.writer (%s)", err)
	}

	// create wrapper zip that contains the hashed backup and the hash
	// file itself
	wrapperZipBuffer := new(bytes.Buffer)
	wrapperZipWriter := zip.NewWriter(wrapperZipBuffer)

	// write internal backup zip
	zipFile, err := wrapperZipWriter.Create(internalBackupFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create internal backup zip in wrapper zip (%s)", err)
	}

	// copy internal backup zip into wrapper
	_, err = io.Copy(zipFile, internalZipBuffer)
	if err != nil {
		return nil, fmt.Errorf("failed to copy internal backup zip into wrapper zip (%s)", err)
	}

	// create hash file in wrapper zip
	zipFileHashFile, err := wrapperZipWriter.Create(internalBackupHashFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create internal backup zip hash in wrapper zip (%s)", err)
	}

	// write hash (as hex string) file in wrapper zip
	_, err = io.WriteString(zipFileHashFile, fmt.Sprintf("%x", internalZipHasher.Sum(nil)))
	if err != nil {
		return nil, fmt.Errorf("failed to copy internal backup hash into wrapper zip (%s)", err)
	}

	// close wrapper zip writer (note: Close() writes the gzip footer and cannot be deferred)
	err = wrapperZipWriter.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close wrapper zip.writer (%s)", err)
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
		return backupFileDetails{}, fmt.Errorf("could not write backup file to disk (%s)", err)
	}

	service.logger.Infof("backup saved to disk (%s)", fileName)

	// return info about new file
	return backupFileDetails{
		Name:      fileName,
		Size:      len(zipFileData),
		ModTime:   createdAt, // not always 100% exact, but close enough
		CreatedAt: &createdAt,
	}, nil
}
