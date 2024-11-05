package output

import (
	"crypto/sha1"
	"fmt"
	"net/http"
)

// PemObject is an interface for objects that can be written to the client as
// PEM data. It contains all methods needed to do this.
type PemObject interface {
	outFile
	PemContent() string
}

// WritePem sends an object supporting PEM output to the client as the appropriate application type
func (service *Service) WritePem(w http.ResponseWriter, r *http.Request, obj PemObject) {
	pemContent := []byte(obj.PemContent())

	file := outFileObj{
		filename:        obj.FilenameNoExt() + ".pem",
		content:         pemContent,
		httpContentType: "application/x-pem-file",
		modTime:         obj.Modtime(),
		eTag:            fmt.Sprintf("\"%x\"", sha1.Sum(pemContent)),
	}

	service.writeFile(w, r, file)
}
