package output

import (
	"bytes"
	"fmt"
	"net/http"
)

// WriteZip sends a zip file with the specified filename using the
// supplied Buffer
func (service *Service) WriteZip(w http.ResponseWriter, filename string, zipBuffer *bytes.Buffer) (bytesWritten int, err error) {
	service.logger.Debugf("writing zip %s to client", filename)

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.WriteHeader(http.StatusOK)

	bytesWritten, err = w.Write(zipBuffer.Bytes())
	if err != nil {
		service.logger.Errorf("error writing zip (%s)", err)
		return -1, errWriteZipError
	}

	return bytesWritten, nil
}
