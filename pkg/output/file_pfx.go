package output

import (
	"crypto/sha1"
	"fmt"
	"net/http"
)

// PfxObject is an interface for objects that can be written to the client as
// PFX data. It contains all methods needed to do this.
type PfxObject interface {
	outFile
	PfxContent(legacy3DES bool) ([]byte, error)
}

// WritePfx sends an object supporting PFX output to the client as the appropriate application type
func (service *Service) WritePfx(w http.ResponseWriter, r *http.Request, obj PfxObject, legacy3DES bool) error {
	pfxContent, err := obj.PfxContent(legacy3DES)
	if err != nil {
		service.logger.Errorf("error generating pfx (%s)", err)
		return err
	}

	file := outFileObj{
		filename:        obj.FilenameNoExt() + ".pfx",
		content:         pfxContent,
		httpContentType: "application/x-pkcs12",
		modTime:         obj.Modtime(),
		eTag:            fmt.Sprintf("\"%x\"", sha1.Sum(pfxContent)),
	}

	service.writeFile(w, r, file)

	return nil
}
