package download

import (
	"legocerthub-backend/pkg/output"
	"legocerthub-backend/pkg/storage"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// ServeKeyPem returns the private key to the client
func (service *Service) GetKeyPemFile(w http.ResponseWriter, r *http.Request) (err error) {
	// if not running https, error
	if !service.https && !service.devMode {
		return output.ErrUnavailableHttp
	}

	// track if apiKey came from URL
	apiKeyInUrl := false

	// get key name
	params := httprouter.ParamsFromContext(r.Context())
	keyName := params.ByName("name")

	// get api key from header
	apiKey := r.Header.Get("X-API-Key")
	// try to get from apikey header if X-API-Key was empty
	if apiKey == "" {
		apiKey = r.Header.Get("apikey")
	}
	// if apiKey is still blank (i.e. not in a proper header), check if apiKey is in the URL
	if apiKey == "" {
		apiKey = params.ByName("apiKey")
		apiKeyInUrl = true
	}
	// if apiKey is still blank, definitely unauthorized
	if apiKey == "" {
		service.logger.Debug(errBlankApiKey)
		return output.ErrUnauthorized
	}

	service.logger.Debugf(keyName)

	// get the key from storage
	key, err := service.storage.GetOneKeyByName(keyName, true)
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

	// if apiKey came from URL, and key does not support this, error
	if apiKeyInUrl && !key.ApiKeyViaUrl {
		service.logger.Debug(errApiKeyFromUrlDisallowed)
		return output.ErrUnauthorized
	}

	// verify apikey matches private key's apiKey
	if apiKey != *key.ApiKey {
		service.logger.Debug(errWrongApiKey)
		return output.ErrUnauthorized
	}

	// return pem file to client
	_, err = service.output.WritePem(w, *key.Pem)
	if err != nil {
		service.logger.Error(err)
		return output.ErrWritePemFailed
	}

	return nil
}
