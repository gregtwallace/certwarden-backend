package output

import (
	"bytes"
	"fmt"
	"net/http"
	"time"
)

// writeZip sends a zip file with the specified filename using the supplied data
func (service *Service) writeZip(w http.ResponseWriter, r *http.Request, filename string, zipData []byte) {
	// log output
	service.logger.Debugf("writing zip %s to client", filename)

	// convert data to Reader
	contentReader := bytes.NewReader(zipData)

	// Set Content-Type and Content-Disposition headers explicitly
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	// do not write HTTP Status, ServeContent will handle this

	// ServeContent (technically fielname is not needed here since Content-Type is set explicitly above)
	http.ServeContent(w, r, filename, time.Time{}, contentReader)
}

// WriteZipNoStoreCache sends a zip file with the specified filename using the supplied data
// including a no-store header indicating the file should not be stored in cache. This is useful
// for sensitive files or files that always change.
func (service *Service) WriteZipNoStoreCache(w http.ResponseWriter, r *http.Request, filename string, zipData []byte) {
	// write no-store Cache header
	w.Header().Set("Cache-Control", "no-store")

	// call common writer
	service.writeZip(w, r, filename, zipData)
}

// add zip w/ ETag if ever needed
