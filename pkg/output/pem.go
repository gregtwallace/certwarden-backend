package output

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// PemObject is an interface for objects that can be written to the client as
// PEM data. It contains all methods needed to do this.
type PemObject interface {
	PemFilename() string
	PemContent() string
}

// WritePem sends an object supporting PEM output to the client as the appropriate application type
// Note: currently error is not possible
func (service *Service) WritePem(w http.ResponseWriter, r *http.Request, obj PemObject) error {
	// get filename and log for auditing
	filename := obj.PemFilename()
	service.logger.Debugf("writing pem %s to client", filename)

	// get pem content and convert to Reader
	contentReader := strings.NewReader(obj.PemContent())

	// Set Content-Type and Content-Disposition headers explicitly
	w.Header().Set("Content-Type", "application/x-pem-file")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	// do not write HTTP Status, ServeContent will handle this

	// ServeContent (technically fielname is not needed here since Content-Type is set explicitly above)
	http.ServeContent(w, r, filename, time.Time{}, contentReader)

	return nil
}
