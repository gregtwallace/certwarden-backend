package download

import (
	"legocerthub-backend/pkg/output"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// DownloadCertViaHeader is the handler to write a cert to the client
// if the proper apiKey is provided via header (standard method)
func (service *Service) DownloadCertViaHeader(w http.ResponseWriter, r *http.Request) (err error) {
	// get name from request
	params := httprouter.ParamsFromContext(r.Context())
	certName := params.ByName("name")

	// get apiKey from header
	apiKey := r.Header.Get("X-API-Key")
	// try to get from apikey header if X-API-Key was empty
	if apiKey == "" {
		apiKey = r.Header.Get("apikey")
	}

	// fetch the cert's newest order using the apiKey
	order, err := service.getCertNewestValidOrder(certName, apiKey, false)
	if err != nil {
		return err
	}

	// return pem file to client
	_, err = service.output.WritePem(w, r, order)
	if err != nil {
		service.logger.Error(err)
		return output.ErrWritePemFailed
	}

	return nil
}

// DownloadCertViaUrl is the handler to write a cert to the client
// if the proper apiKey is provided via URL (NOT recommended - only implemented
// to support clients that can't specify the apiKey header)
func (service *Service) DownloadCertViaUrl(w http.ResponseWriter, r *http.Request) (err error) {
	// get cert name & apiKey
	params := httprouter.ParamsFromContext(r.Context())
	certName := params.ByName("name")

	apiKey := getApiKeyFromParams(params)

	// fetch the cert's newest order using the apiKey
	order, err := service.getCertNewestValidOrder(certName, apiKey, true)
	if err != nil {
		return err
	}

	// return pem file to client
	_, err = service.output.WritePem(w, r, order)
	if err != nil {
		service.logger.Error(err)
		return output.ErrWritePemFailed
	}

	return nil
}
