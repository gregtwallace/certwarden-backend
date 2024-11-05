package download

import (
	"certwarden-backend/pkg/output"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// DownloadCertViaHeader is the handler to write a cert to the client
// if the proper apiKey is provided via header (standard method)
func (service *Service) DownloadCertViaHeader(w http.ResponseWriter, r *http.Request) *output.Error {
	// get name from request
	params := httprouter.ParamsFromContext(r.Context())
	certName := params.ByName("name")

	// get apiKey from header
	apiKey := getApiKeyFromHeader(w, r)

	// fetch the cert's newest order using the apiKey
	order, err := service.getCertNewestValidOrder(certName, apiKey, false, false)
	if err != nil {
		return err
	}

	// return pem file to client
	service.output.WritePem(w, r, order)

	return nil
}

// DownloadCertViaUrl is the handler to write a cert to the client
// if the proper apiKey is provided via URL (NOT recommended - only implemented
// to support clients that can't specify the apiKey header)
func (service *Service) DownloadCertViaUrl(w http.ResponseWriter, r *http.Request) *output.Error {
	// get cert name & apiKey
	params := httprouter.ParamsFromContext(r.Context())
	certName := params.ByName("name")

	apiKey := getApiKeyFromParams(params)

	// fetch the cert's newest order using the apiKey
	order, err := service.getCertNewestValidOrder(certName, apiKey, true, false)
	if err != nil {
		return err
	}

	// return pem file to client
	service.output.WritePem(w, r, order)

	return nil
}
