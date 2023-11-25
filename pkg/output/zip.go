package output

import (
	"bytes"
	"fmt"
	"net/http"
	"time"
)

// WriteZip sends a zip file with the specified filename using the
// supplied Buffer
func (service *Service) WriteZip(w http.ResponseWriter, r *http.Request, filename string, zipData []byte) {
	// log output
	service.logger.Debugf("writing zip %s to client", filename)

	// convert data to Reader
	contentReader := bytes.NewReader(zipData)

	// Set Content-Type and Content-Disposition headers explicitly
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	// calculate sha1 of the PemContent and set as a simplistic ETag
	w.Header().Set("ETag", etagValue(zipData))

	// do not write HTTP Status, ServeContent will handle this

	// ServeContent (technically fielname is not needed here since Content-Type is set explicitly above)
	http.ServeContent(w, r, filename, time.Time{}, contentReader)
}
