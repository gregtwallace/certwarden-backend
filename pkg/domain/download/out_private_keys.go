package download

import (
	"certwarden-backend/pkg/output"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// DownloadKeyViaHeader is the handler to write a private key to the client
// if the proper apiKey is provided via header (standard method)
func (service *Service) DownloadKeyViaHeader(w http.ResponseWriter, r *http.Request) *output.Error {
	// get key name
	params := httprouter.ParamsFromContext(r.Context())
	keyName := params.ByName("name")

	// get apiKey from header
	apiKey := getApiKeyFromHeader(w, r)

	// fetch the key using the apiKey
	key, err := service.getKey(keyName, apiKey, false)
	if err != nil {
		return err
	}

	// return pem file to client
	service.output.WritePem(w, r, key)

	return nil
}

// DownloadKeyViaUrl is the handler to write a private key to the client
// if the proper apiKey is provided via URL (NOT recommended - only implemented
// to support clients that can't specify the apiKey header)
func (service *Service) DownloadKeyViaUrl(w http.ResponseWriter, r *http.Request) *output.Error {
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
	service.output.WritePem(w, r, key)

	return nil
}
