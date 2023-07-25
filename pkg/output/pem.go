package output

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"net/http"
	"time"
)

// WritePem sends the pem string to the client as the appropriate
// application type
func (service *Service) WritePem(w http.ResponseWriter, filename string, pem string) (bytesWritten int, err error) {
	service.logger.Debugf("writing pem %s to client", filename)

	// for cert chain: application/pem-certificate-chain
	w.Header().Set("Content-Type", "application/x-pem-file")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.WriteHeader(http.StatusOK)

	bytesWritten, err = w.Write([]byte(pem))
	if err != nil {
		return -1, err
	}

	return bytesWritten, nil
}

// WritePem sends the pem string to the client as the appropriate
// application type, ETag and Timestamp (if available)
func (service *Service) WritePemWithCondition(w http.ResponseWriter, r *http.Request, filename string, pem string, modtime time.Time) {
	service.logger.Debugf("writing pem %s to client", filename)

	// Get SHA1 for PEM
	hasher := sha1.New()
	hasher.Write([]byte(pem))

	// for cert chain: application/pem-certificate-chain
	w.Header().Set("Content-Type", "application/x-pem-file")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	// Add ETag header
	w.Header().Set("ETag", fmt.Sprintf("\"%x\"", hasher.Sum(nil)))
	http.ServeContent(w, r, filename, modtime, bytes.NewReader([]byte(pem)))
}
