package download

import (
	"encoding/pem"
	"fmt"
	"legocerthub-backend/pkg/output"
	"legocerthub-backend/pkg/storage"
	"net/http"
	"time"

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

	// fetch the cert using the apiKey
	certPem, _, modtime, err := service.getCertPem(certName, apiKey, true, false)
	if err != nil {
		return err
	}

	// return pem file to client
	service.output.WritePemWithCondition(w, r, fmt.Sprintf("%s.cert.pem", certName), certPem, modtime)

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

	// fetch the cert using the apiKey
	certPem, _, modtime, err := service.getCertPem(certName, apiKey, true, true)
	if err != nil {
		return err
	}

	// return pem file to client
	service.output.WritePemWithCondition(w, r, fmt.Sprintf("%s.cert.pem", certName), certPem, modtime)

	return nil
}

// getCertPem returns the cert pem and private key name if the apiKey matches the
// requested key. It also checks the apiKeyViaUrl property if the client is making
// a request with the apiKey in the Url. The pem is from the most recent valid
// order for the specified cert. The keyName is the name of the key that corresponds
// to that order.
func (service *Service) getCertPem(certName string, apiKey string, fullChain bool, apiKeyViaUrl bool) (certPem string, keyName string, modtime time.Time, err error) {
	// if not running https, error
	if !service.https && !service.devMode {
		return "", "", modtime, output.ErrUnavailableHttp
	}

	// if apiKey is blank, definitely unauthorized
	if apiKey == "" {
		service.logger.Debug(errBlankApiKey)
		return "", "", modtime, output.ErrUnauthorized
	}

	// get the cert from storage
	cert, err := service.storage.GetOneCertByName(certName)
	if err != nil {
		// special error case for no record found
		if err == storage.ErrNoRecord {
			service.logger.Debug(err)
			return "", "", modtime, output.ErrNotFound
		} else {
			service.logger.Error(err)
			return "", "", modtime, output.ErrStorageGeneric
		}
	}

	// if apiKey came from URL, and cert does not support this, error
	if apiKeyViaUrl && !cert.ApiKeyViaUrl {
		service.logger.Debug(errApiKeyFromUrlDisallowed)
		return "", "", modtime, output.ErrUnauthorized
	}

	// verify apikey matches cert apikey (new or old)
	if (apiKey != cert.ApiKey) && (apiKey != cert.ApiKeyNew) {
		service.logger.Debug(errWrongApiKey)
		return "", "", modtime, output.ErrUnauthorized
	}

	// get pem of the most recent valid order for the cert
	_, certPem, err = service.storage.GetCertPemById(cert.ID)
	if err != nil {
		// special error case for no record found
		// of note, this indicates the cert exists but there is no
		// valid order (cert pem) for the cert
		// log warn instead of debug since this is indicative
		// there may be an issue for the user to investigate
		if err == storage.ErrNoRecord {
			service.logger.Warn(err)
			return "", "", modtime, output.ErrNotFound
		} else {
			service.logger.Error(err)
			return "", "", modtime, output.ErrStorageGeneric
		}
	}

	// pem cant be blank
	if certPem == "" {
		service.logger.Debug(errNoPem)
		return "", "", modtime, output.ErrStorageGeneric
	}

	// if not fullchain, discard rest of chain
	if !fullChain {
		certBlock, _ := pem.Decode([]byte(certPem))
		certPem = string(pem.EncodeToMemory(certBlock))
	}

	modtime = time.Unix(int64(cert.CreatedAt), 0)

	// return pem content, key name and modification time
	return certPem, cert.CertificateKey.Name, modtime, nil
}
