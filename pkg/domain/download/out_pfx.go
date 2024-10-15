package download

import (
	"certwarden-backend/pkg/domain/orders"
	"certwarden-backend/pkg/output"
	"crypto/x509"
	"encoding/base64"
	"net/http"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
	"software.sslmate.com/src/go-pkcs12"
)

// modified Order to allow implementation of custom out functions
// to properly output the desired content
type pfxPrivateCertificateChain privateCertificateChain

// pfxPrivateCertificateChain Output Methods

func (pcc pfxPrivateCertificateChain) FilenameNoExt() string {
	return pcc.Certificate.Name
}

func TrimIdentifiers(pem string) []byte {
	trimParts := []string{
		"-----BEGIN CERTIFICATE-----\n",
		"-----BEGIN PRIVATE KEY-----\n",
		"-----BEGIN RSA PRIVATE KEY-----\n",
		"-----END CERTIFICATE-----\n",
		"-----END PRIVATE KEY-----\n",
		"-----END RSA PRIVATE KEY-----\n",
		"-----BEGIN EC PRIVATE KEY-----\n",
		"-----END EC PRIVATE KEY-----\n",
	}

	for _, part := range trimParts {
		pem = strings.ReplaceAll(pem, part, "")
	}

	decoded, err := base64.StdEncoding.DecodeString(pem)
	if err != nil {
		return nil
	}

	return decoded
}

func ParsePrivateKey(byteval []byte) (interface{}, error) {
	var res interface{}
	var err error

	res, err = x509.ParsePKCS1PrivateKey(byteval)
	if err != nil {
		res, err = x509.ParsePKCS8PrivateKey(byteval)
		if err != nil {
			res, err = x509.ParseECPrivateKey(byteval)
		}
	}

	return res, err
}

// PfxContent returns the combined key + cert + chain pfx content
func (pcc pfxPrivateCertificateChain) PfxContent() []byte {
	certPem := orders.Order(pcc).PemContentNoChain()
	certChainPem := orders.Order(pcc).PemContentChainOnly()
	keyPem := pcc.FinalizedKey.PemContent()

	certPemTrimmed := TrimIdentifiers(certPem)
	keyPemTrimmed := TrimIdentifiers(keyPem)

	fullChainPem := []*x509.Certificate{}

	splitCertChainPem := strings.Split(certChainPem, "-----END CERTIFICATE-----\n")
	for _, chainCert := range splitCertChainPem {
		if len(chainCert) == 0 {
			continue
		}
		trimmed := TrimIdentifiers(chainCert)
		cert, err := x509.ParseCertificate(trimmed)
		if err != nil {
			return nil
		}
		fullChainPem = append(fullChainPem, cert)
	}

	x509cert, err := x509.ParseCertificate(certPemTrimmed)
	if err != nil {
		return nil
	}

	x509key, err := ParsePrivateKey(keyPemTrimmed)
	if err != nil {
		return nil
	}

	pfx, err := pkcs12.Modern.Encode(x509key, x509cert, fullChainPem, pcc.FinalizedKey.ApiKey)
	if err != nil {
		return nil
	}

	return pfx
}

func (pcc pfxPrivateCertificateChain) Modtime() time.Time {
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

// DownloadPfxViaHeader
func (service *Service) DownloadPfxViaHeader(w http.ResponseWriter, r *http.Request) *output.Error {
	// get cert name
	params := httprouter.ParamsFromContext(r.Context())
	certName := params.ByName("name")

	// get apiKey from header
	apiKeysCombined := getApiKeyFromHeader(w, r)

	// fetch the private cert
	privCert, err := service.getCertNewestPfx(certName, apiKeysCombined, false)
	if err != nil {
		return err
	}

	// return pem file to client
	service.output.WritePfx(w, r, privCert)

	return nil
}

// DownloadPfxViaUrl
func (service *Service) DownloadPfxViaUrl(w http.ResponseWriter, r *http.Request) *output.Error {
	// get cert name & apiKey
	params := httprouter.ParamsFromContext(r.Context())
	certName := params.ByName("name")

	apiKeysCombined := getApiKeyFromParams(params)

	// fetch the private cert
	privCert, err := service.getCertNewestPfx(certName, apiKeysCombined, true)
	if err != nil {
		return err
	}

	// return pem file to client
	service.output.WritePfx(w, r, privCert)

	return nil
}

// getCertNewestPfx gets the appropriate order for the requested Cert and sets its type to
// privateCertificateChain so the proper data is outputted.
// To avoid unauthorized output of a key, both the certificate and key apiKeys must be provided. The format
// for this is the certificate apikey appended to the private key's apikey using a '.' as a separator.
// It also checks the apiKeyViaUrl property if the client is making a request with the apiKey in the Url.
func (service *Service) getCertNewestPfx(certName string, apiKeysCombined string, apiKeyViaUrl bool) (pfxPrivateCertificateChain, *output.Error) {
	// separate the apiKeys
	apiKeys := strings.Split(apiKeysCombined, ".")

	// error if not exactly 2 apiKeys
	if len(apiKeys) != 2 {
		return pfxPrivateCertificateChain{}, output.ErrUnauthorized
	}

	certApiKey := apiKeys[0]
	keyApiKey := apiKeys[1]

	// fetch the cert's newest valid order
	order, err := service.getCertNewestValidOrder(certName, certApiKey, apiKeyViaUrl)
	if err != nil {
		return pfxPrivateCertificateChain{}, err
	}

	// confirm the private key is valid
	if order.FinalizedKey == nil {
		service.logger.Debug(errFinalizedKeyMissing)
		return pfxPrivateCertificateChain{}, output.ErrNotFound
	}

	// validate the apiKey for the private key is correct
	if order.FinalizedKey.ApiKey != keyApiKey {
		service.logger.Debug(errWrongApiKey)
		return pfxPrivateCertificateChain{}, output.ErrUnauthorized
	}

	// return pfx content
	return pfxPrivateCertificateChain(order), nil
}
