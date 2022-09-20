package download

import (
	"legocerthub-backend/pkg/output"
	"legocerthub-backend/pkg/storage"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// DownloadKeyViaHeader is the handler to write a private key to the client
// if the proper apiKey is provided via header (standard method)
func (service *Service) DownloadKeyViaHeader(w http.ResponseWriter, r *http.Request) (err error) {
	// get key name
	params := httprouter.ParamsFromContext(r.Context())
	keyName := params.ByName("name")

	// get apiKey from header
	apiKey := r.Header.Get("X-API-Key")
	// try to get from apikey header if X-API-Key was empty
	if apiKey == "" {
		apiKey = r.Header.Get("apikey")
	}

	// fetch the key using the apiKey
	keyPem, err := service.getKeyPem(keyName, apiKey, false)
	if err != nil {
		return err
	}

	// return pem file to client
	_, err = service.output.WritePem(w, keyPem)
	if err != nil {
		service.logger.Error(err)
		return output.ErrWritePemFailed
	}

	return nil
}

// DownloadKeyViaUrl is the handler to write a private key to the client
// if the proper apiKey is provided via URL (NOT recommended - only implemented
// to support clients that can't specify the apiKey header)
func (service *Service) DownloadKeyViaUrl(w http.ResponseWriter, r *http.Request) (err error) {
	// get key name & apiKey
	params := httprouter.ParamsFromContext(r.Context())
	keyName := params.ByName("name")

	apiKey := getApiKeyFromParams(params)

	// fetch the key using the apiKey
	keyPem, err := service.getKeyPem(keyName, apiKey, true)
	if err != nil {
		return err
	}

	// return pem file to client
	_, err = service.output.WritePem(w, keyPem)
	if err != nil {
		service.logger.Error(err)
		return output.ErrWritePemFailed
	}

	return nil
}

// getKeyPemFile returns the private key pem if the apiKey matches
// the requested key. It also checks the apiKeyViaUrl property if
// the client is making a request with the apiKey in the Url.
func (service *Service) getKeyPem(keyName string, apiKey string, apiKeyViaUrl bool) (keyPem string, err error) {
	// if not running https, error
	if !service.https && !service.devMode {
		return "", output.ErrUnavailableHttp
	}

	// if apiKey is blank, definitely unauthorized
	if apiKey == "" {
		service.logger.Debug(errBlankApiKey)
		return "", output.ErrUnauthorized
	}

	// get the key from storage
	key, err := service.storage.GetOneKeyByName(keyName, true)
	if err != nil {
		// special error case for no record found
		if err == storage.ErrNoRecord {
			service.logger.Debug(err)
			return "", output.ErrNotFound
		} else {
			service.logger.Error(err)
			return "", output.ErrStorageGeneric
		}
	}

	// if apiKey came from URL, and key does not support this, error
	if apiKeyViaUrl && !key.ApiKeyViaUrl {
		service.logger.Debug(errApiKeyFromUrlDisallowed)
		return "", output.ErrUnauthorized
	}

	// verify apikey matches private key's apiKey
	if apiKey != key.ApiKey {
		service.logger.Debug(errWrongApiKey)
		return "", output.ErrUnauthorized
	}

	// return pem content
	return key.Pem, nil
}
