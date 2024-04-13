package backup

import (
	"certwarden-backend/pkg/output"
	"net/http"
)

type backupFileListResponse struct {
	output.JsonResponse
	Config      *Config             `json:"config"`
	BackupFiles []backupFileDetails `json:"backup_files"`
}

// ListDiskBackupsHandler returns a list of the backups currently on the disk
// as well as some information about them
func (service *Service) ListDiskBackupsHandler(w http.ResponseWriter, r *http.Request) *output.Error {
	// get file list
	filesInfo, err := service.listBackupFiles()
	if err != nil {
		return output.ErrInternal
	}

	// write response
	response := &backupFileListResponse{}
	response.StatusCode = http.StatusOK
	response.Message = "ok"
	response.Config = service.config
	response.BackupFiles = filesInfo

	err = service.output.WriteJSON(w, response)
	if err != nil {
		service.logger.Errorf("failed to write json (%s)", err)
		return output.ErrWriteJsonError
	}

	return nil
}
