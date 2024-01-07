package download

import (
	"fmt"
	"legocerthub-backend/pkg/domain/orders"
	"legocerthub-backend/pkg/output"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
)

// modified Order to allow implementation of custom pem functions
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

// PemModtime returns the time the rootChain's certificate resource was last updated at.
func (rc rootChain) PemModtime() time.Time {
	// use Order default
	return orders.Order(rc).PemModtime()
}

// end rootChain Output Pem Methods

// DownloadCertRootChainViaHeader is the handler to write just a
// cert's chain to the client, if the proper apiKey is provided via
// header (standard method)
func (service *Service) DownloadCertRootChainViaHeader(w http.ResponseWriter, r *http.Request) *output.Error {
	// get cert name
	params := httprouter.ParamsFromContext(r.Context())
	certName := params.ByName("name")

	// get apiKey from header
	apiKey := getApiKeyFromHeader(w, r)

	// fetch the cert's newest order using the apiKey, as rootChain type
	rootChain, err := service.getCertNewestValidRootChain(certName, apiKey, false)
	if err != nil {
		return err
	}

	// return pem file to client
	service.output.WritePem(w, r, rootChain)

	return nil
}

// DownloadCertRootChainViaUrl is the handler to write just a
// cert's chain to the client, if the proper apiKey is provided via
// URL (NOT recommended - only implemented to support clients that
// can't specify the apiKey header)
func (service *Service) DownloadCertRootChainViaUrl(w http.ResponseWriter, r *http.Request) *output.Error {
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
	service.output.WritePem(w, r, rootChain)

	return nil
}

// getCertNewestValidRootChain gets the appropriate order for the requested Cert and sets its type to
// rootChain so the proper data is outputted
func (service *Service) getCertNewestValidRootChain(certName string, apiKey string, apiKeyViaUrl bool) (rootChain, *output.Error) {
	order, err := service.getCertNewestValidOrder(certName, apiKey, apiKeyViaUrl)
	if err != nil {
		return rootChain{}, err
	}

	return rootChain(order), nil
}
