package backup

import (
	"legocerthub-backend/pkg/output"
	"net/http"
	"os"
)

type backupFileListResponse struct {
	output.JsonResponse
	BackupFiles []backupFileDetails `json:"backup_files"`
}

// ListDiskBackupsHandler returns a list of the backups currently on the disk
// as well as some information about them
func (service *Service) ListDiskBackupsHandler(w http.ResponseWriter, r *http.Request) *output.Error {
	// read file list from backup dir
	files, err := os.ReadDir(service.cleanDataStorageBackupPath)
	if err != nil {
		service.logger.Errorf("failed to list backup directory contents (%s)", err)
		return output.ErrInternal
	}

	// backup file list for json response
	backupFiles := []backupFileDetails{}

	// for each file, add it to json response list if it is a backup file
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
				backupFiles = append(backupFiles, bakFile)
			}
		}
	}

	// write response
	response := &backupFileListResponse{}
	response.StatusCode = http.StatusOK
	response.Message = "ok"
	response.BackupFiles = backupFiles

	err = service.output.WriteJSON(w, response)
	if err != nil {
		service.logger.Errorf("failed to write json (%s)", err)
		return output.ErrWriteJsonError
	}

	return nil
}
