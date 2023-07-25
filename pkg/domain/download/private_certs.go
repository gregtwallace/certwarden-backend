package download

import (
	"fmt"
	"legocerthub-backend/pkg/output"
	"net/http"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
)

// DownloadPrivateCertViaHeader
func (service *Service) DownloadPrivateCertViaHeader(w http.ResponseWriter, r *http.Request) (err error) {
	// get cert name
	params := httprouter.ParamsFromContext(r.Context())
	certName := params.ByName("name")

	// get apiKey from header
	apiKey := r.Header.Get("X-API-Key")
	// try to get from apikey header if X-API-Key was empty
	if apiKey == "" {
		apiKey = r.Header.Get("apikey")
	}

	// fetch the private cert
	certPem, modtime, err := service.getPrivateCertPem(certName, apiKey, false)
	if err != nil {
		return err
	}

	// return pem file to client
	service.output.WritePemWithCondition(w, r, fmt.Sprintf("%s.certkey.pem", certName), certPem, modtime)

	return nil
}

// DownloadPrivateCertViaUrl
func (service *Service) DownloadPrivateCertViaUrl(w http.ResponseWriter, r *http.Request) (err error) {
	// get cert name & apiKey
	params := httprouter.ParamsFromContext(r.Context())
	certName := params.ByName("name")

	apiKey := getApiKeyFromParams(params)

	// fetch the private cert
	certPem, modtime, err := service.getPrivateCertPem(certName, apiKey, true)
	if err != nil {
		return err
	}

	// return pem file to client
	service.output.WritePemWithCondition(w, r, fmt.Sprintf("%s.certkey.pem", certName), certPem, modtime)

	return nil
}

// getPrivateCertPem returns the cert's private key pem appended to the cert's
// public certificate pem. ApiKeys should be the certificate apikey appended
// to the private key's apikey using a '.' as a separator. It also checks
// the apiKeyViaUrl property if the client is making a request with the apiKey
// in the Url. The pem is from the most recent valid order for the specified cert.
// The key is the matching key for the order. An order is returned if the key
// has been deleted.
// TODO: Allow entire cert chain to be provided
func (service *Service) getPrivateCertPem(certName string, apiKeysString string, apiKeyViaUrl bool) (privateCertPem string, modtime time.Time, err error) {
	// if not running https, error
	if !service.https && !service.devMode {
		return "", modtime, output.ErrUnavailableHttp
	}

	// separate the apiKeys
	apiKeys := strings.Split(apiKeysString, ".")

	// error if not exactly 2 apiKeys
	if len(apiKeys) != 2 {
		return "", modtime, output.ErrUnauthorized
	}

	certApiKey := apiKeys[0]
	keyApiKey := apiKeys[1]

	// fetch the full certificate chain
	certPem, keyName, certmodtime, err := service.getCertPem(certName, certApiKey, false, apiKeyViaUrl)
	if err != nil {
		return "", modtime, err
	}

	// fetch the matching private key
	keyPem, keymodtime, err := service.getKeyPem(keyName, keyApiKey, apiKeyViaUrl)
	if err != nil {
		return "", modtime, err
	}

	// append key and cert
	privateCertPem = keyPem + string([]byte{10}) + certPem

	if certmodtime.After(keymodtime) {
		modtime = certmodtime
	} else {
		modtime = keymodtime
	}
	// return pem content
	return privateCertPem, modtime, nil
}
