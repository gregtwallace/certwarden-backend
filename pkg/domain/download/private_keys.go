package download

import (
	"fmt"
	"legocerthub-backend/pkg/output"
	"legocerthub-backend/pkg/storage"
	"net/http"
	"time"

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
	keyPem, modtime, err := service.getKeyPem(keyName, apiKey, false)
	if err != nil {
		return err
	}

	// return pem file to client
	service.output.WritePemWithCondition(w, r, fmt.Sprintf("%s.key.pem", keyName), keyPem, modtime)

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
	keyPem, modtime, err := service.getKeyPem(keyName, apiKey, true)
	if err != nil {
		return err
	}

	// return pem file to client
	service.output.WritePemWithCondition(w, r, fmt.Sprintf("%s.key.pem", keyName), keyPem, modtime)

	return nil
}

// getKeyPemFile returns the private key pem if the apiKey matches
// the requested key. It also checks the apiKeyViaUrl property if
// the client is making a request with the apiKey in the Url.
func (service *Service) getKeyPem(keyName string, apiKey string, apiKeyViaUrl bool) (keyPem string, modtime time.Time, err error) {
	// if not running https, error
	if !service.https && !service.devMode {
		return "", modtime, output.ErrUnavailableHttp
	}

	// if apiKey is blank, definitely unauthorized
	if apiKey == "" {
		service.logger.Debug(errBlankApiKey)
		return "", modtime, output.ErrUnauthorized
	}

	// get the key from storage
	key, unixtime, err := service.storage.GetOneKeyByName(keyName)
	if err != nil {
		// special error case for no record found
		if err == storage.ErrNoRecord {
			service.logger.Debug(err)
			return "", modtime, output.ErrNotFound
		} else {
			service.logger.Error(err)
			return "", modtime, output.ErrStorageGeneric
		}
	}

	// if key is disabled via API, error
	if key.ApiKeyDisabled {
		service.logger.Debug(errApiDisabled)
		return "", modtime, output.ErrUnauthorized
	}

	// if apiKey came from URL, and key does not support this, error
	if apiKeyViaUrl && !key.ApiKeyViaUrl {
		service.logger.Debug(errApiKeyFromUrlDisallowed)
		return "", modtime, output.ErrUnauthorized
	}

	// verify apikey matches private key's apiKey (new or old)
	if (apiKey != key.ApiKey) && (apiKey != key.ApiKeyNew) {
		service.logger.Debug(errWrongApiKey)
		return "", modtime, output.ErrUnauthorized
	}

	modtime = time.Unix(int64(unixtime), 0)

	// return pem content
	return key.Pem, modtime, nil
}
