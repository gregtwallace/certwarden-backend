package output

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"net/http"
	"time"
)

// PemObject is an interface for objects that can be written to the client as
// PEM data. It contains all methods needed to do this.
type PemObject interface {
	PemFilename() string
	PemContent() string
	PemModtime() time.Time
}

// WritePem sends an object supporting PEM output to the client as the appropriate application type
// Note: currently error is not possible
func (service *Service) WritePem(w http.ResponseWriter, r *http.Request, obj PemObject) error {
	// get filename and log for auditing
	filename := obj.PemFilename()
	service.logger.Debugf("writing pem %s to client", filename)

	// get pem content and convert to Reader
	pemContent := []byte(obj.PemContent())
	contentReader := bytes.NewReader(pemContent)

	// Set Content-Type and Content-Disposition headers explicitly
	w.Header().Set("Content-Type", "application/x-pem-file")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	// calculate sha1 of the PemContent and set as a simplistic ETag
	etagValue := sha1.Sum(pemContent)
	w.Header().Set("ETag", fmt.Sprintf("\"%x\"", etagValue))

	// do not write HTTP Status, ServeContent will handle this

	// ServeContent (technically fielname is not needed here since Content-Type is set explicitly above)
	http.ServeContent(w, r, filename, obj.PemModtime(), contentReader)

	return nil
}
