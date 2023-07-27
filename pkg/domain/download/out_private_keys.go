package download

import (
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
	key, err := service.getKey(keyName, apiKey, false)
	if err != nil {
		return err
	}

	// return pem file to client
	_, err = service.output.WritePem(w, r, key)
	if err != nil {
		return err
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
	key, err := service.getKey(keyName, apiKey, true)
	if err != nil {
		return err
	}

	// return pem file to client
	_, err = service.output.WritePem(w, r, key)
	if err != nil {
		return err
	}

	return nil
}
