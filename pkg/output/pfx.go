package output

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"net/http"
)

// PfxObject is an interface for objects that can be written to the client as
// PFX data. It contains all methods needed to do this.
type PfxObject interface {
	OutFile
	PfxContent() []byte
}

// WritePem sends an object supporting PEM output to the client as the appropriate application type
// Note: currently error is not possible
func (service *Service) WritePfx(w http.ResponseWriter, r *http.Request, obj PfxObject) {
	// get filename and log for auditing
	filename := obj.FilenameNoExt() + ".pfx"
	service.logger.Debugf("writing pfx %s to client %s", filename, r.RemoteAddr)

	// get pem content and convert to Reader
	pfxContent := obj.PfxContent()
	contentReader := bytes.NewReader(pfxContent)

	// Set Content-Type and Content-Disposition headers explicitly
	w.Header().Set("Content-Type", "application/x-pkcs12")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	// calculate sha1 of the PfxContent and set as a simplistic ETag
	w.Header().Set("ETag", fmt.Sprintf("\"%x\"", sha1.Sum(pfxContent)))

	// do not write HTTP Status, ServeContent will handle this

	// ServeContent (technically fielname is not needed here since Content-Type is set explicitly above)
	http.ServeContent(w, r, filename, obj.Modtime(), contentReader)
}
