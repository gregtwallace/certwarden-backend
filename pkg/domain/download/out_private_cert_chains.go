package download

import (
	"certwarden-backend/pkg/domain/orders"
	"certwarden-backend/pkg/output"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
)

// modified Order to allow implementation of custom out functions
// to properly output the desired content
type privateCertificateChain orders.Order

// privateCertificateChain Output Methods

func (pcc privateCertificateChain) FilenameNoExt() string {
	return fmt.Sprintf("%s.certchainkey.pem", pcc.Certificate.Name)
}

// PemContent returns the combined key + cert + chain pem content
func (pcc privateCertificateChain) PemContent() string {
	keyPem := pcc.FinalizedKey.PemContent()

	// all cert pem data
	certPem := orders.Order(pcc).PemContent()

	// append key + LF + cert
	return keyPem + string([]byte{10}) + certPem
}

func (pcc privateCertificateChain) Modtime() time.Time {
	// if key is nil, return 0 time since Pem output of thie type will fail anyway without a key
	if pcc.FinalizedKey == nil {
		return time.Time{}
	}

	// key time
	keyModtime := pcc.FinalizedKey.Modtime()

	// order (cert) time
	certModtime := orders.Order(pcc).Modtime()

	// return more recent of the two
	if keyModtime.After(certModtime) {
		return keyModtime
	}
	return certModtime
}

// end privateCertificateChain Output Methods

// DownloadPrivateCertChainViaHeader
func (service *Service) DownloadPrivateCertChainViaHeader(w http.ResponseWriter, r *http.Request) *output.Error {
	// get cert name
	params := httprouter.ParamsFromContext(r.Context())
	certName := params.ByName("name")

	// get apiKey from header
	apiKeysCombined := getApiKeyFromHeader(w, r)

	// fetch the private cert
	privCert, err := service.getCertNewestValidPrivateCertChain(certName, apiKeysCombined, false)
	if err != nil {
		return err
	}

	// return pem file to client
	service.output.WritePem(w, r, privCert)

	return nil
}

// DownloadPrivateCertChainViaUrl
func (service *Service) DownloadPrivateCertChainViaUrl(w http.ResponseWriter, r *http.Request) *output.Error {
	// get cert name & apiKey
	params := httprouter.ParamsFromContext(r.Context())
	certName := params.ByName("name")

	apiKeysCombined := getApiKeyFromParams(params)

	// fetch the private cert
	privCert, err := service.getCertNewestValidPrivateCertChain(certName, apiKeysCombined, true)
	if err != nil {
		return err
	}

	// return pem file to client
	service.output.WritePem(w, r, privCert)

	return nil
}

// getCertNewestValidPrivateCertChain gets the appropriate order for the requested Cert and sets its type to
// privateCertificateChain so the proper data is outputted.
// To avoid unauthorized output of a key, both the certificate and key apiKeys must be provided. The format
// for this is the certificate apikey appended to the private key's apikey using a '.' as a separator.
// It also checks the apiKeyViaUrl property if the client is making a request with the apiKey in the Url.
func (service *Service) getCertNewestValidPrivateCertChain(certName string, apiKeysCombined string, apiKeyViaUrl bool) (privateCertificateChain, *output.Error) {
	// separate the apiKeys
	apiKeys := strings.Split(apiKeysCombined, ".")

	// error if not exactly 2 apiKeys
	if len(apiKeys) != 2 {
		return privateCertificateChain{}, output.ErrUnauthorized
	}

	certApiKey := apiKeys[0]
	keyApiKey := apiKeys[1]

	// fetch the cert's newest valid order
	order, err := service.getCertNewestValidOrder(certName, certApiKey, apiKeyViaUrl)
	if err != nil {
		return privateCertificateChain{}, err
	}

	// confirm the private key is valid
	if order.FinalizedKey == nil {
		service.logger.Debug(errFinalizedKeyMissing)
		return privateCertificateChain{}, output.ErrNotFound
	}

	// validate the apiKey for the private key is correct
	if order.FinalizedKey.ApiKey != keyApiKey {
		service.logger.Debug(errWrongApiKey)
		return privateCertificateChain{}, output.ErrUnauthorized
	}

	// return pem content
	return privateCertificateChain(order), nil
}
