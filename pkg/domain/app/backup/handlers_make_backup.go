package backup

import (
	"fmt"
	"legocerthub-backend/pkg/output"
	"net/http"
)

type backupFileMakeResponse struct {
	output.JsonResponse
	BackupFile backupFileDetails `json:"backup_file"`
}

// makeDiskBackupNowHandler creates a new backup of the application in the backup
// folder location; it does not send the backup to the client
func (service *Service) MakeDiskBackupNowHandler(w http.ResponseWriter, r *http.Request) *output.Error {
	backupFileDetails, err := service.CreateBackupOnDisk()
	if err != nil {
		service.logger.Errorf("failed to make on disk backup (%s)", err)
		return output.ErrInternal
	}

	// write success response
	response := &backupFileMakeResponse{}
	response.StatusCode = http.StatusCreated
	response.Message = fmt.Sprintf("%s written to server storage", backupFileDetails.Name)
	response.BackupFile = backupFileDetails

	err = service.output.WriteJSON(w, response)
	if err != nil {
		service.logger.Errorf("failed to write json (%s)", err)
		return output.ErrWriteJsonError
	}

	return nil
}
