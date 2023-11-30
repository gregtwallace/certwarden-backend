package acme

import (
	"crypto/x509"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"go.uber.org/zap/zapcore"
)

var errBadOrderPem = errors.New("pem returned from ACME server failed safety check (see: rfc8555 s 11.4)")

// NewOrderPayload is the payload to post to ACME newOrder
type NewOrderPayload struct {
	// notBefore and notAfter are optional and not implemented
	Identifiers IdentifierSlice `json:"identifiers"`
}

// LE response with order information
type Order struct {
	Status         string          `json:"status"`
	Expires        timeString      `json:"expires"`
	Identifiers    IdentifierSlice `json:"identifiers"`
	Error          *Error          `json:"error,omitempty"`
	Authorizations []string        `json:"authorizations"`
	Finalize       string          `json:"finalize"`
	Certificate    *string         `json:"certificate,omitempty"`
	NotBefore      *timeString     `json:"notBefore,omitempty"`
	NotAfter       *timeString     `json:"notAfter,omitempty"`
	Location       string          `json:"-"` // omit because it is in the header
}

// Account response decoder
func unmarshalOrder(bodyBytes []byte, headers http.Header) (response Order, err error) {
	err = json.Unmarshal(bodyBytes, &response)
	if err != nil {
		return Order{}, err
	}

	// order location (url) isn't part of the JSON response, add it from the header.
	response.Location = headers.Get("Location")

	return response, nil
}

// NewOrder posts a secure message to the NewOrder URL of the directory
func (service *Service) NewOrder(payload NewOrderPayload, accountKey AccountKey) (response Order, err error) {
	// post new-order
	bodyBytes, headers, err := service.postToUrlSigned(payload, service.dir.NewOrder, accountKey)
	if err != nil {
		return Order{}, err
	}

	// unmarshal response
	response, err = unmarshalOrder(bodyBytes, headers)
	if err != nil {
		return Order{}, err
	}

	return response, nil
}

// GetOrder does a POST-as-GET to fetch the current state of the given order URL
func (service *Service) GetOrder(orderUrl string, accountKey AccountKey) (response Order, err error) {
	// POST-as-GET
	bodyBytes, headers, err := service.postAsGet(orderUrl, accountKey)
	if err != nil {
		return Order{}, err
	}

	// unmarshal response
	response, err = unmarshalOrder(bodyBytes, headers)
	if err != nil {
		return Order{}, err
	}

	return response, nil
}

// FinalizeOrder posts the specified CSR to the specified finalize URL
func (service *Service) FinalizeOrder(finalizeUrl string, derCsr []byte, accountKey AccountKey) (response Order, err error) {
	// pretty log CSR names if in debug
	if service.logger.Level() == zapcore.DebugLevel {
		csr, prettyErr := x509.ParseCertificateRequest(derCsr)
		if prettyErr == nil {
			// log CN and DNS names
			service.logger.Debugf("attempting finalize using csr with common name: %s ; and dns name(s): %s", csr.Subject.CommonName, csr.DNSNames)

			// Log full CSR
			// prettyBytes, prettyErr := json.MarshalIndent(csr, "", "\t")
			// if prettyErr == nil {
			// 	service.logger.Debugf("%s", prettyBytes)
			// }
		}
	}

	// insert csr into expected json payload format
	payload := struct {
		Csr string `json:"csr"`
	}{
		Csr: encodeString(derCsr),
	}

	// post csr to finalize URL
	bodyBytes, headers, err := service.postToUrlSigned(payload, finalizeUrl, accountKey)
	if err != nil {
		return Order{}, err
	}

	// unmarshal response
	response, err = unmarshalOrder(bodyBytes, headers)
	if err != nil {
		return Order{}, err
	}

	return response, nil
}

// DownloadCertificate uses POST-as-GET to download a valid certificate from the specified
// url.
func (service *Service) DownloadCertificate(certificateUrl string, accountKey AccountKey) (pemChain string, err error) {
	// POST-as-GET
	bodyBytes, headers, err := service.postAsGet(certificateUrl, accountKey)
	if err != nil {
		return "", err
	}

	// this server only supports pem (application/pem-certificate-chain)
	contentType := headers.Get("Content-type")
	if contentType != "application/pem-certificate-chain" {
		return "", errBadOrderPem
	}

	// validate ACME server didn't return malicious pem (see: RFC8555 s 11.4)
	pemCheck := string(bodyBytes)
	beginString := "-----BEGIN"
	mustBeFollowedBy := " CERTIFICATE"

	// if there is never a begin, invalid pem
	i := strings.Index(pemCheck, beginString)
	if i == -1 {
		return "", errBadOrderPem
	}

	// check every begin to ensure it is followed by CERTIFICATE
	for ; i != -1; i = strings.Index(pemCheck, beginString) {
		pemCheck = pemCheck[(i + len(beginString)):]

		if !strings.HasPrefix(pemCheck, mustBeFollowedBy) {
			return "", errBadOrderPem
		}
	}
	// end - validate (RFC8555 s 11.4)

	// log alternate links in debug
	// TODO: maybe care about these at some point
	// altLinks := []string{}
	// for _, altLink := range headers.Values("Link") {
	// 	if strings.HasSuffix(altLink, "rel=\"alternate\"") {
	// 		altLinks = append(altLinks, altLink)
	// 	}
	// }
	// service.logger.Debugf("alternate download links: %s", altLinks)

	return string(bodyBytes), nil
}
