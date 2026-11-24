package backup

import (
	"certwarden-backend/pkg/output"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/julienschmidt/httprouter"
)

// DeleteDiskBackupHandler deletes an existing backup from the server's disk.
func (service *Service) DeleteDiskBackupHandler(w http.ResponseWriter, r *http.Request) *output.JsonError {
	// params
	filenameParam := httprouter.ParamsFromContext(r.Context()).ByName("filename")

	// validate filename is in the form of a backup file (prevent unauthorized file download)
	if !isBackupFileName(filenameParam) {
		return output.ErrorJsonErrValidationFailed(errors.New("invalid filename"))
	}

	// stat file to confirm it exists
	_, err := os.Stat(service.cleanDataStorageBackupPath + string(filepath.Separator) + filenameParam)
	if err != nil {
		// 404 for file doesn't exist
		if errors.Is(err, os.ErrNotExist) {
			return output.ErrorJsonErrNotFound(errors.New(service.cleanDataStorageBackupPath + string(filepath.Separator) + filenameParam))
		}
		// internal for any other issue
		err = fmt.Errorf("failed to stat disk backup for delete (%w)", err)
		service.logger.Error(err)
		return output.ErrorJsonErrInternal(err)
	}

	// delete file
	err = os.Remove(service.cleanDataStorageBackupPath + string(filepath.Separator) + filenameParam)
	if err != nil {
		err = fmt.Errorf("failed to delete disk backup (%w)", err)
		service.logger.Error(err)
		return output.ErrorJsonErrInternal(err)
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
		return output.ErrorJsonErrWriteJsonError(err)
	}

	return nil
}
