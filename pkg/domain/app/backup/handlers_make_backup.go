package backup

import (
	"certwarden-backend/pkg/output"
	"fmt"
	"net/http"
)

type backupFileMakeResponse struct {
	output.JsonResponse
	BackupFile backupFileDetails `json:"backup_file"`
}

// makeDiskBackupNowHandler creates a new backup of the application in the backup
// folder location; it does not send the backup to the client
func (service *Service) MakeDiskBackupNowHandler(w http.ResponseWriter, r *http.Request) *output.JsonError {
	backupFileDetails, err := service.CreateBackupOnDisk()
	if err != nil {
		err = fmt.Errorf("failed to make on disk backup (%s)", err)
		service.logger.Error(err)
		return output.JsonErrInternal(err)
	}

	// write success response
	response := &backupFileMakeResponse{}
	response.StatusCode = http.StatusCreated
	response.Message = fmt.Sprintf("%s written to server storage", backupFileDetails.Name)
	response.BackupFile = backupFileDetails

	err = service.output.WriteJSON(w, response)
	if err != nil {
		service.logger.Errorf("failed to write json (%s)", err)
		return output.JsonErrWriteJsonError(err)
	}

	return nil
}
