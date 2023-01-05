package download

import (
	"bytes"
	"encoding/pem"
	"fmt"
	"legocerthub-backend/pkg/output"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

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

	// fetch the cert chain using the apiKey
	certChainPem, err := service.getCertRootChainPem(certName, apiKey, false)
	if err != nil {
		return err
	}

	// return pem file to client
	_, err = service.output.WritePem(w, fmt.Sprintf("%s.chain.pem", certName), certChainPem)
	if err != nil {
		service.logger.Error(err)
		return output.ErrWritePemFailed
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

	// fetch the cert chain using the apiKey
	certChainPem, err := service.getCertRootChainPem(certName, apiKey, true)
	if err != nil {
		return err
	}

	// return pem file to client
	_, err = service.output.WritePem(w, fmt.Sprintf("%s.chain.pem", certName), certChainPem)
	if err != nil {
		service.logger.Error(err)
		return output.ErrWritePemFailed
	}

	return nil
}

// getCertRootChainPem returns the cert's root chain pem if the
// apiKey matches the requested key. It also checks the apiKeyViaUrl
// property if the client is making a request with the apiKey in the Url.
// The pem is from the most recent valid order for the specified cert.
// TODO: Allow tweaking of root chain components
func (service *Service) getCertRootChainPem(certName string, apiKey string, apiKeyViaUrl bool) (rootChainPem string, err error) {
	// if not running https, error
	if !service.https && !service.devMode {
		return "", output.ErrUnavailableHttp
	}

	// fetch the full certificate chain
	certPem, _, err := service.getCertPem(certName, apiKey, true, apiKeyViaUrl)
	if err != nil {
		return "", err
	}

	// decode the first cert in the chain and discard it
	// this effectively leaves the root chain as the "rest"
	_, chain := pem.Decode([]byte(certPem))

	// remove any extraneouse chars before the first cert begins
	beginIndex := bytes.Index(chain, []byte{45}) // ascii code for dash character
	chain = chain[beginIndex:]

	// return pem content
	return string(chain), nil
}
