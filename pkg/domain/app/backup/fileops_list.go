package backup

import (
	"os"
)

// backupFileDetails contains information about an on disk backup file
type backupFileDetails struct {
	Name      string `json:"name"`
	Size      int    `json:"size"`
	ModTime   int    `json:"modtime"`
	CreatedAt *int   `json:"created_at,omitempty"`
}

// time returns created_at if it exists, otherwise it returns modtime
func (backupFile *backupFileDetails) unixTime() int {
	if backupFile.CreatedAt != nil {
		return *backupFile.CreatedAt
	}

	// else use modtime
	return backupFile.ModTime
}

// listBackupFiles returns a list of the backup files on the server
func (service *Service) listBackupFiles() ([]backupFileDetails, error) {
	// read file list from backup dir
	files, err := os.ReadDir(service.cleanDataStorageBackupPath)
	if err != nil {
		service.logger.Errorf("failed to list backup directory contents (%s)", err)
		return nil, err
	}

	// for each file, add it to json response list if it is a backup file
	backupFilesInfo := []backupFileDetails{}
	for i := range files {
		// ignore directories
		if !files[i].IsDir() {

			// get name
			bakFile := backupFileDetails{}
			bakFile.Name = files[i].Name()

			// only list if it is a backup file
			if isBackupFile(bakFile.Name) {
				// stat file
				fStat, err := files[i].Info()
				if err != nil {
					// if cant stat, bad file, skip it
					service.logger.Warnf("backup file %s in backup dir that cant be stat'd (%s)", bakFile.Name, err)
					continue
				}

				// populate file properties
				bakFile.Size = int(fStat.Size())
				bakFile.ModTime = int(fStat.ModTime().Unix())

				// calculate created at from filename, omit if doesn't decode
				nameTime, err := backupZipTime(bakFile.Name)
				if err != nil {
					service.logger.Warnf("backup file with improperly formatted timestamp in backup dir (err decoding time: %s)", err)
				} else {
					bakFile.CreatedAt = new(int)
					*bakFile.CreatedAt = int(nameTime.Unix())
				}

				// add to list
				backupFilesInfo = append(backupFilesInfo, bakFile)
			}
		}
	}

	return backupFilesInfo, nil
}
