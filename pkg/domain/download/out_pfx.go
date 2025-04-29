package download

import (
	"certwarden-backend/pkg/domain/orders"
	"certwarden-backend/pkg/output"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	"software.sslmate.com/src/go-pkcs12"
)

// modified Order to allow implementation of custom out functions
// to properly output the desired content
type pfxPrivateCertificateChain orders.Order

// pfxPrivateCertificateChain Output Methods

func (pfxpcc pfxPrivateCertificateChain) FilenameNoExt() string {
	return pfxpcc.Certificate.Name
}

func (pfxpcc pfxPrivateCertificateChain) Modtime() time.Time {
	return orders.Order(pfxpcc).Modtime()
}

// keyPemToKey returns the private key from pemBytes
func keyPemToKey(keyPem []byte) (key any, err error) {
	// decode private key
	keyPemBlock, _ := pem.Decode(keyPem)
	if keyPemBlock == nil {
		return nil, errors.New("key pem block did not decode")
	}

	// parsing depends on block type
	switch keyPemBlock.Type {
	case "RSA PRIVATE KEY": // PKCS1
		var rsaKey *rsa.PrivateKey
		rsaKey, err = x509.ParsePKCS1PrivateKey(keyPemBlock.Bytes)
		if err != nil {
			return nil, err
		}
		return rsaKey, nil

	case "EC PRIVATE KEY": // SEC1, ASN.1
		var ecdKey *ecdsa.PrivateKey
		ecdKey, err = x509.ParseECPrivateKey(keyPemBlock.Bytes)
		if err != nil {
			return nil, err
		}
		return ecdKey, nil

	case "PRIVATE KEY": // PKCS8
		pkcs8Key, err := x509.ParsePKCS8PrivateKey(keyPemBlock.Bytes)
		if err != nil {
			return nil, err
		}
		return pkcs8Key, nil

	default:
		// fallthrough
	}

	return nil, errors.New("key pem block type unsupported")
}

// certPemToCerts returns the certificate from cert pem bytes. if the pem
// bytes contain more than one certificate, the first is returned as the
// certificate and the rest are returned as an array for what is presumably
// the rest of a chain
func certPemToCerts(certPem []byte) (cert *x509.Certificate, certChain []*x509.Certificate, err error) {
	// decode 1st cert
	certPemBlock, rest := pem.Decode(certPem)
	if certPemBlock == nil {
		return nil, nil, errors.New("cert pem block did not decode")
	}

	// parse 1st cert
	cert, err = x509.ParseCertificate(certPemBlock.Bytes)
	if err != nil {
		return nil, nil, err
	}

	// decode cert chain
	certChainPemBlocks := []*pem.Block{}
	for {
		// try to decode next block
		var nextCertBlock *pem.Block
		nextCertBlock, rest = pem.Decode(rest)

		// no next block, done
		if nextCertBlock == nil {
			break
		}

		// success, append
		certChainPemBlocks = append(certChainPemBlocks, nextCertBlock)
	}

	// parse each cert in chain
	certChain = []*x509.Certificate{}
	for i := range certChainPemBlocks {
		certChainMember, err := x509.ParseCertificate(certChainPemBlocks[i].Bytes)
		if err != nil {
			return nil, nil, err
		}

		certChain = append(certChain, certChainMember)
	}

	return cert, certChain, nil
}

// PfxContent returns the combined key + cert + chain pfx content; it accepts a bool
// legacy3DES that when true uses the legacy 3DES encryption algorithm. This is needed
// for compatibility with some older systems.
func (pfxpcc pfxPrivateCertificateChain) PfxContent(legacy3DES bool) (pfxData []byte, err error) {
	// get private key
	key, err := keyPemToKey([]byte(pfxpcc.FinalizedKey.PemContent()))
	if err != nil {
		return nil, err
	}

	// get cert and chain (if there is a chain)
	cert, certChain, err := certPemToCerts([]byte(orders.Order(pfxpcc).PemContent()))
	if err != nil {
		return nil, err
	}

	// encode using legace pkcs12 (3DES)
	if legacy3DES {
		pfxData, err = pkcs12.Legacy.Encode(key, cert, certChain, pfxpcc.FinalizedKey.ApiKey)
		if err != nil {
			return nil, err
		}

		return pfxData, nil
	}

	// encode using modern pkcs12 standard
	pfxData, err = pkcs12.Modern.Encode(key, cert, certChain, pfxpcc.FinalizedKey.ApiKey)
	if err != nil {
		return nil, err
	}

	return pfxData, nil
}

// end privateCertificateChain Output Methods

// DownloadPfxViaHeader
func (service *Service) DownloadPfxViaHeader(w http.ResponseWriter, r *http.Request) *output.JsonError {
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

	// legacy 3DES specified?
	legacy3DES := false
	if r.URL.Query().Has("3des") {
		legacy3DES = true
	}

	// return pfx file to client
	pfxPrivCert := pfxPrivateCertificateChain(order)
	service.output.WritePfx(w, r, pfxPrivCert, legacy3DES)

	return nil
}

// DownloadPfxViaUrl
func (service *Service) DownloadPfxViaUrl(w http.ResponseWriter, r *http.Request) *output.JsonError {
	// get cert name & apiKey
	params := httprouter.ParamsFromContext(r.Context())
	certName := params.ByName("name")

	apiKeysCombined := getApiKeyFromParams(params)

	// fetch the private cert
	order, outErr := service.getCertNewestValidOrder(certName, apiKeysCombined, true, true)
	if outErr != nil {
		return outErr
	}

	// legacy 3DES specified?
	legacy3DES := false
	if r.URL.Query().Has("3des") {
		legacy3DES = true
	}

	// return pfx file to client
	pfxPrivCert := pfxPrivateCertificateChain(order)
	err := service.output.WritePfx(w, r, pfxPrivCert, legacy3DES)
	if err != nil {
		return output.JsonErrInternal(err)
	}

	return nil
}
