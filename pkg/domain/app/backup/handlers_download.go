package backup

import (
	"bytes"
	"certwarden-backend/pkg/output"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/julienschmidt/httprouter"
)

// DownloadBackupNowHandler sends the client a backup of the application at this
// exact moment; nothing is saved locally on the server
func (service *Service) DownloadBackupNowHandler(w http.ResponseWriter, r *http.Request) *output.JsonError {
	// get query for backup options
	query := r.URL.Query()
	withOnDiskBackupsParam := query.Get("withondiskbackups")

	// set bools (use default if not explicitly opposite)
	withOnDiskBackups := false
	if strings.EqualFold(withOnDiskBackupsParam, "true") {
		withOnDiskBackups = true
	}

	// make zip file
	zipBytes, err := service.createDataBackup(withOnDiskBackups)
	if err != nil {
		service.logger.Debug(err)
		return output.JsonErrInternal(err)
	}

	// filename & remove extension
	zipFileName, _ := makeBackupZipFileName()
	extension := filepath.Ext(zipFileName)
	zipFilenameNoExt := zipFileName[0 : len(zipFileName)-len(extension)]

	// output
	service.output.WriteZip(w, r, zipFilenameNoExt, zipBytes)

	return nil
}

// DownloadDiskBackupHandler sends an existing backup file from the server to
// the client
func (service *Service) DownloadDiskBackupHandler(w http.ResponseWriter, r *http.Request) *output.JsonError {
	// params
	filenameParam := httprouter.ParamsFromContext(r.Context()).ByName("filename")

	// validate filename is in the form of a backup file (prevent unauthorized file download)
	if !isBackupFileName(filenameParam) {
		return output.JsonErrValidationFailed(errors.New("invalid filename"))
	}

	// open file for reading
	f, err := os.Open(service.cleanDataStorageBackupPath + string(filepath.Separator) + filenameParam)
	if err != nil {
		// 404 for file doesn't exist
		if errors.Is(err, os.ErrNotExist) {
			return output.JsonErrNotFound(errors.New(service.cleanDataStorageBackupPath + string(filepath.Separator) + filenameParam))
		}
		// internal for any other issue
		err = fmt.Errorf("failed to open disk backup for download (%s)", err)
		service.logger.Error(err)
		return output.JsonErrInternal(err)
	}
	defer f.Close()

	// read entire file
	zipBuffer := bytes.NewBuffer(nil)
	_, err = io.Copy(zipBuffer, f)
	if err != nil {
		err = fmt.Errorf("failed to copy disk backup to buffer for download (%s)", err)
		service.logger.Error(err)
		return output.JsonErrInternal(err)
	}

	// remove extension
	extension := filepath.Ext(filenameParam)
	zipFilenameNoExt := filenameParam[0 : len(filenameParam)-len(extension)]

	// send file to client
	service.output.WriteZip(w, r, zipFilenameNoExt, zipBuffer.Bytes())

	return nil
}
