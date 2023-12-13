package backup

import (
	"errors"
	"fmt"
	"legocerthub-backend/pkg/output"
	"net/http"
	"os"
	"path/filepath"

	"github.com/julienschmidt/httprouter"
)

// DeleteDiskBackupHandler deletes an existing backup from the server's disk.
func (service *Service) DeleteDiskBackupHandler(w http.ResponseWriter, r *http.Request) *output.Error {
	// params
	filenameParam := httprouter.ParamsFromContext(r.Context()).ByName("filename")

	// validate filename is in the form of a backup file (prevent unauthorized file download)
	if !isBackupFile(filenameParam) {
		return output.ErrValidationFailed
	}

	// stat file to confirm it exists
	_, err := os.Stat(service.cleanDataStorageBackupPath + string(filepath.Separator) + filenameParam)
	if err != nil {
		// 404 for file doesn't exist
		if errors.Is(err, os.ErrNotExist) {
			return output.ErrNotFound
		}
		// internal for any other issue
		service.logger.Errorf("failed to stat disk backup for delete (%s)", err)
		return output.ErrInternal
	}

	// delete file
	err = os.Remove(service.cleanDataStorageBackupPath + string(filepath.Separator) + filenameParam)
	if err != nil {
		service.logger.Errorf("failed to delete disk backup (%s)", err)
		return output.ErrInternal
	}

	service.logger.Infof("backup deleted from disk (%s)", filenameParam)

	// write response
	response := &output.JsonResponse{
		StatusCode: http.StatusOK,
		Message:    fmt.Sprintf("deleted disk backup (%s)", filenameParam),
	}

	err = service.output.WriteJSON(w, response)
	if err != nil {
		service.logger.Errorf("failed to write json (%s)", err)
		return output.ErrWriteJsonError
	}

	return nil
}
