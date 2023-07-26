package output

import (
	"fmt"
	"net/http"
)

// PemObject is an interface for objects that can be written to the client as
// PEM data. It contains all methods needed to do this.
type PemObject interface {
	PemFilename() string
	PemContent() string
}

// WritePem sends an object supporting PEM output to the client as the appropriate
// application type
func (service *Service) WritePem(w http.ResponseWriter, r *http.Request, obj PemObject) (bytesWritten int, err error) {
	// log for auditing
	filename := obj.PemFilename()
	service.logger.Debugf("writing pem %s to client", filename)

	// for cert chain: application/pem-certificate-chain
	w.Header().Set("Content-Type", "application/x-pem-file")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.WriteHeader(http.StatusOK)

	bytesWritten, err = w.Write([]byte(obj.PemContent()))
	if err != nil {
		return -1, err
	}

	return bytesWritten, nil
}
