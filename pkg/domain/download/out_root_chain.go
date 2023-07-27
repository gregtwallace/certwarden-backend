package download

import (
	"fmt"
	"legocerthub-backend/pkg/domain/orders"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// rootChain is modified Order to allow implementation of custom pem functions
// to properly output the desired content
type rootChain orders.Order

// rootChain Output Pem Methods

// PemFilename is the chain filename
func (rc rootChain) PemFilename() string {
	return fmt.Sprintf("%s.chain.pem", rc.Certificate.Name)
}

// PemContent returns the PemContentChainOnly instead of what order would
// normally return
func (rc rootChain) PemContent() string {
	return orders.Order(rc).PemContentChainOnly()
}

// end rootChain Output Pem Methods

// DownloadCertRootChainViaHeader is the handler to write just a
// cert's chain to the client, if the proper apiKey is provided via
// header (standard method)
func (service *Service) DownloadCertRootChainViaHeader(w http.ResponseWriter, r *http.Request) (err error) {
	// get cert name
	params := httprouter.ParamsFromContext(r.Context())
	certName := params.ByName("name")

	// get apiKey from header
	apiKey := r.Header.Get("X-API-Key")
	// try to get from apikey header if X-API-Key was empty
	if apiKey == "" {
		apiKey = r.Header.Get("apikey")
	}

	// fetch the cert's newest order using the apiKey, as rootChain type
	rootChain, err := service.getCertNewestValidRootChain(certName, apiKey, false)
	if err != nil {
		return err
	}

	// return pem file to client
	_, err = service.output.WritePem(w, r, rootChain)
	if err != nil {
		return err
	}

	return nil
}

// DownloadCertRootChainViaUrl is the handler to write just a
// cert's chain to the client, if the proper apiKey is provided via
// URL (NOT recommended - only implemented to support clients that
// can't specify the apiKey header)
func (service *Service) DownloadCertRootChainViaUrl(w http.ResponseWriter, r *http.Request) (err error) {
	// get cert name & apiKey
	params := httprouter.ParamsFromContext(r.Context())
	certName := params.ByName("name")

	apiKey := getApiKeyFromParams(params)

	// fetch the cert's newest order using the apiKey, as rootChain type
	rootChain, err := service.getCertNewestValidRootChain(certName, apiKey, true)
	if err != nil {
		return err
	}

	// return pem file to client
	_, err = service.output.WritePem(w, r, rootChain)
	if err != nil {
		return err
	}

	return nil
}

// getCertNewestValidRootChain gets the appropriate order for the requested Cert and sets its type to
// rootChain so the proper data is outputted
func (service *Service) getCertNewestValidRootChain(certName string, apiKey string, apiKeyViaUrl bool) (rootChain, error) {
	order, err := service.getCertNewestValidOrder(certName, apiKey, apiKeyViaUrl)
	if err != nil {
		return rootChain{}, err
	}

	return rootChain(order), nil
}
