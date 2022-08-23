package certificates

import (
	"legocerthub-backend/pkg/challenges"
	"legocerthub-backend/pkg/output"
	"legocerthub-backend/pkg/storage"
	"legocerthub-backend/pkg/validation"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// GetAllCertificates fetches all certs from storage and outputs them as JSON
func (service *Service) GetAllCerts(w http.ResponseWriter, r *http.Request) (err error) {
	// get keys from storage
	keys, err := service.storage.GetAllCerts()
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

// GetOneCert is an http handler that returns one Certificate based on its unique id in the
// form of JSON written to w
func (service *Service) GetOneCert(w http.ResponseWriter, r *http.Request) (err error) {
	// convert id param to an integer
	idParam := httprouter.ParamsFromContext(r.Context()).ByName("certid")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// if id is new, provide some info
	err = validation.IsIdNew(&id)
	if err == nil {
		return service.GetNewCertOptions(w, r)
	}

	// if id < 0 & not new, it is definitely not valid
	err = validation.IsIdExisting(&id)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// get from storage
	cert, err := service.storage.GetOneCertById(id, false)
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

	// return response to client
	_, err = service.output.WriteJSON(w, http.StatusOK, cert, "certificate")
	if err != nil {
		service.logger.Error(err)
		return output.ErrWriteJsonFailed
	}

	return nil
}

// GetNewCertOptions is an http handler that returns information the client GUI needs to properly
// present options when the user is creating a certificate
func (service *Service) GetNewCertOptions(w http.ResponseWriter, r *http.Request) (err error) {
	// certificate options / info to assist client with new certificate posting
	newCertOptions := newCertOptions{}

	// available accounts
	newCertOptions.AvailableAccounts, err = service.accounts.GetAvailableAccounts()
	if err != nil {
		service.logger.Error(err)
		return output.ErrStorageGeneric
	}

	// available private keys
	newCertOptions.AvailableKeys, err = service.keys.GetAvailableKeys()
	if err != nil {
		service.logger.Error(err)
		return output.ErrStorageGeneric
	}

	// available challenge methods
	newCertOptions.AvailableChallengeMethods = challenges.ListOfMethods()

	// return response to client
	_, err = service.output.WriteJSON(w, http.StatusOK, newCertOptions, "certificate_options")
	if err != nil {
		service.logger.Error(err)
		return output.ErrWriteJsonFailed
	}

	return nil
}

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
		service.logger.Error(err)
		return output.ErrStorageGeneric
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
