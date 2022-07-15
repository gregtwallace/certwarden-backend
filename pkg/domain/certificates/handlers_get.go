package certificates

import (
	"legocerthub-backend/pkg/output"
	"net/http"
)

// GetAllCertificates fetches all certs from storage and outputs them as JSON
func (service *Service) GetAllCertificates(w http.ResponseWriter, r *http.Request) (err error) {
	// get keys from storage
	keys, err := service.storage.GetAllCertificates()
	if err != nil {
		service.logger.Error(err)
		return output.ErrStorageGeneric
	}

	// return response to client
	_, err = service.output.WriteJSON(w, http.StatusOK, keys, "certificates")
	if err != nil {
		service.logger.Error(err)
		return output.ErrWriteJsonFailed
	}

	return nil
}
