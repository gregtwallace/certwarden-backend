package download

import (
	"legocerthub-backend/pkg/output"
	"legocerthub-backend/pkg/storage"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// GetCertPemFile returns the pem file for the most recent valid order to the client
// TODO: implement additional options e.g. specify chain vs. just cert
func (service *Service) GetCertPemFile(w http.ResponseWriter, r *http.Request) (err error) {
	// if not running https, error
	if !service.https && !service.devMode {
		return output.ErrUnavailableHttp
	}

	// get cert name
	certName := httprouter.ParamsFromContext(r.Context()).ByName("name")

	// get api key from header
	apiKey := r.Header.Get("X-API-Key")
	// try to get from apikey header if X-API-Key was empty
	if apiKey == "" {
		apiKey = r.Header.Get("apikey")
	}
	// if apiKey is blank, definitely not authorized
	if apiKey == "" {
		service.logger.Debug(err)
		return output.ErrUnauthorized
	}

	// get the cert from storage
	cert, err := service.storage.GetOneCertByName(certName, false)
	if err != nil {
		// special error case for no record found
		if err == storage.ErrNoRecord {
			service.logger.Debug(err)
			return output.ErrNotFound
		} else {
			service.logger.Error(err)
			return output.ErrStorageGeneric
		}
	}

	// verify apikey matches private key
	if apiKey != *cert.ApiKey {
		service.logger.Debug(err)
		return output.ErrUnauthorized
	}

	// get pem of the most recent valid order for the cert
	pem, err := service.storage.GetCertPemById(*cert.ID)
	if err != nil {
		// special error case for no record found
		// of note, this indicates the cert exists but there is no
		// valid order (cert pem) for the cert
		// log warn instead of debug since this is indicative
		// there may be an issue for the user to investigate
		if err == storage.ErrNoRecord {
			service.logger.Warn(err)
			return output.ErrNotFound
		} else {
			service.logger.Error(err)
			return output.ErrStorageGeneric
		}
	}

	// pem cant be blank
	if pem == "" {
		service.logger.Debug(err)
		return output.ErrStorageGeneric
	}

	// return pem file to client
	_, err = service.output.WritePem(w, pem)
	if err != nil {
		service.logger.Error(err)
		return output.ErrWritePemFailed
	}

	return nil
}
