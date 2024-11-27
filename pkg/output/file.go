package output

import (
	"bytes"
	"fmt"
	"net/http"
	"time"
)

// outFile is a common interface that all file outputs share
type outFile interface {
	FilenameNoExt() string
	Modtime() time.Time
}

// outFileObj is the object the public functions should populate to use the common
// writeFile function
type outFileObj struct {
	filename        string
	content         []byte
	httpContentType string
	modTime         time.Time
	eTag            string
}

// writeNoCacheFile is a generic file output function that is used by other public file output
// functions; it also includes a `no-store` cache header
func (service *Service) writeFile(w http.ResponseWriter, r *http.Request, file outFileObj) {
	// get filename and log for auditing
	filename := file.filename
	service.logger.Debugf("writing file %s to client %s", filename, r.RemoteAddr)

	// get pem content and convert to Reader
	contentReader := bytes.NewReader(file.content)

	// Set Content-Type and Content-Disposition headers explicitly
	w.Header().Set("Content-Type", file.httpContentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	// set eTag if there is one
	if file.eTag != "" {
		w.Header().Set("ETag", file.eTag)
	}

	// write no-store Cache header - these files could change at any time (e.g., a renewal)
	w.Header().Set("Cache-Control", "no-store")

	// do not write HTTP Status, ServeContent will handle this

	// ServeContent (technically fielname is not needed here since Content-Type is set explicitly above)
	http.ServeContent(w, r, filename, file.modTime, contentReader)
}
