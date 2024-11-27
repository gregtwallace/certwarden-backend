package download

import (
	"certwarden-backend/pkg/domain/orders"
	"certwarden-backend/pkg/output"
	"fmt"
	"net/http"
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

func (pc privateCertificate) Modtime() time.Time {
	return orders.Order(pc).Modtime()
}

func (pc privateCertificate) PemContent() string {
	keyPem := pc.FinalizedKey.PemContent()
	// don't include the cert chain
	certPem := orders.Order(pc).PemContentNoChain()

	// append key + LF + cert
	return keyPem + string([]byte{10}) + certPem
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
	order, err := service.getCertNewestValidOrder(certName, apiKeysCombined, false, true)
	if err != nil {
		return err
	}

	// return pem file to client
	privCert := privateCertificate(order)
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
	order, err := service.getCertNewestValidOrder(certName, apiKeysCombined, true, true)
	if err != nil {
		return err
	}

	// return pem file to client
	privCert := privateCertificate(order)
	service.output.WritePem(w, r, privCert)

	return nil
}
