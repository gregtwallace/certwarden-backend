package output

import (
	"net/http"
	"time"
)

// WriteZip sends a zip file with the specified filename using the supplied content
func (service *Service) WriteZip(w http.ResponseWriter, r *http.Request, filenameNoExt string, zipContent []byte) {
	file := outFileObj{
		filename:        filenameNoExt + ".zip",
		content:         zipContent,
		httpContentType: "application/zip",
		modTime:         time.Time{},
		// no eTag
	}

	service.writeFile(w, r, file)
}
