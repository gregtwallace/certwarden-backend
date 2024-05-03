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
type privateCertificate orders.Order

// privateCertificate Output Methods

func (pc privateCertificate) FilenameNoExt() string {
	return fmt.Sprintf("%s.certkey", pc.Certificate.Name)
}

func (pc privateCertificate) PemContent() string {
	keyPem := pc.FinalizedKey.PemContent()
	// don't include the cert chain
	certPem := orders.Order(pc).PemContentNoChain()

	// append key + LF + cert
	return keyPem + string([]byte{10}) + certPem
}

func (pc privateCertificate) Modtime() time.Time {
	// if key is nil, return 0 time since Pem output of thie type will fail anyway without a key
	if pc.FinalizedKey == nil {
		return time.Time{}
	}

	// key time
	keyModtime := pc.FinalizedKey.Modtime()

	// order (cert) time
	certModtime := orders.Order(pc).Modtime()

	// return more recent of the two
	if keyModtime.After(certModtime) {
		return keyModtime
	}
	return certModtime
}

// end privateCertificate Output Methods

// DownloadPrivateCertViaHeader
func (service *Service) DownloadPrivateCertViaHeader(w http.ResponseWriter, r *http.Request) *output.Error {
	// get cert name
	params := httprouter.ParamsFromContext(r.Context())
	certName := params.ByName("name")

	// get apiKey from header
	apiKeysCombined := getApiKeyFromHeader(w, r)

	// fetch the private cert
	privCert, err := service.getCertNewestValidPrivateCert(certName, apiKeysCombined, false)
	if err != nil {
		return err
	}

	// return pem file to client
	service.output.WritePem(w, r, privCert)

	return nil
}

// DownloadPrivateCertViaUrl
func (service *Service) DownloadPrivateCertViaUrl(w http.ResponseWriter, r *http.Request) *output.Error {
	// get cert name & apiKey
	params := httprouter.ParamsFromContext(r.Context())
	certName := params.ByName("name")

	apiKeysCombined := getApiKeyFromParams(params)

	// fetch the private cert
	privCert, err := service.getCertNewestValidPrivateCert(certName, apiKeysCombined, true)
	if err != nil {
		return err
	}

	// return pem file to client
	service.output.WritePem(w, r, privCert)

	return nil
}

// getCertNewestValidPrivateCert gets the appropriate order for the requested Cert and sets its type to
// privateCertificate so the proper data is outputted.
// To avoid unauthorized output of a key, both the certificate and key apiKeys must be provided. The format
// for this is the certificate apikey appended to the private key's apikey using a '.' as a separator.
// It also checks the apiKeyViaUrl property if the client is making a request with the apiKey in the Url.
func (service *Service) getCertNewestValidPrivateCert(certName string, apiKeysCombined string, apiKeyViaUrl bool) (privateCertificate, *output.Error) {
	// separate the apiKeys
	apiKeys := strings.Split(apiKeysCombined, ".")

	// error if not exactly 2 apiKeys
	if len(apiKeys) != 2 {
		return privateCertificate{}, output.ErrUnauthorized
	}

	certApiKey := apiKeys[0]
	keyApiKey := apiKeys[1]

	// fetch the cert's newest valid order
	order, err := service.getCertNewestValidOrder(certName, certApiKey, apiKeyViaUrl)
	if err != nil {
		return privateCertificate{}, err
	}

	// confirm the private key is valid
	if order.FinalizedKey == nil {
		service.logger.Debug(errFinalizedKeyMissing)
		return privateCertificate{}, output.ErrNotFound
	}

	// validate the apiKey for the private key is correct
	if order.FinalizedKey.ApiKey != keyApiKey {
		service.logger.Debug(errWrongApiKey)
		return privateCertificate{}, output.ErrUnauthorized
	}

	// return pem content
	return privateCertificate(order), nil
}
