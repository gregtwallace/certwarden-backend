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
type privateCertificateChain orders.Order

// privateCertificateChain Output Methods

func (pcc privateCertificateChain) FilenameNoExt() string {
	return fmt.Sprintf("%s.certchainkey", pcc.Certificate.Name)
}

func (pcc privateCertificateChain) Modtime() time.Time {
	return orders.Order(pcc).Modtime()
}

// PemContent returns the combined key + cert + chain pem content
func (pcc privateCertificateChain) PemContent() string {
	keyPem := pcc.FinalizedKey.PemContent()

	// all cert pem data
	certPem := orders.Order(pcc).PemContent()

	// append key + LF + cert
	return keyPem + string([]byte{10}) + certPem
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
	order, err := service.getCertNewestValidOrder(certName, apiKeysCombined, false, true)
	if err != nil {
		return err
	}

	// return pem file to client
	privCertChain := privateCertificateChain(order)
	service.output.WritePem(w, r, privCertChain)

	return nil
}

// DownloadPrivateCertChainViaUrl
func (service *Service) DownloadPrivateCertChainViaUrl(w http.ResponseWriter, r *http.Request) *output.Error {
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
	privCertChain := privateCertificateChain(order)
	service.output.WritePem(w, r, privCertChain)

	return nil
}
