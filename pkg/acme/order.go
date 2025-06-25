package acme

import (
	"crypto/x509"
	"encoding/json"
	"net/http"

	"go.uber.org/zap/zapcore"
)

// NewOrderPayload is the payload to post to ACME newOrder
type NewOrderPayload struct {
	// notBefore and notAfter are optional and not implemented
	Identifiers IdentifierSlice `json:"identifiers"`

	// ACME Profiles Extension
	Profile *string `json:"profile,omitempty"`

	// ACME ARI Extension
	Replaces *string `json:"replaces,omitempty"`
}

// LE response with order information
type Order struct {
	Status         string          `json:"status"`
	Expires        *timeString     `json:"expires"`
	Identifiers    IdentifierSlice `json:"identifiers"`
	Error          *Error          `json:"error,omitempty"`
	Authorizations []string        `json:"authorizations"`
	Finalize       string          `json:"finalize"`
	Certificate    *string         `json:"certificate,omitempty"`
	NotBefore      *timeString     `json:"notBefore,omitempty"`
	NotAfter       *timeString     `json:"notAfter,omitempty"`
	Location       string          `json:"-"` // omit because it is in the header

	// ACME Profiles Extension
	Profile *string `json:"profile,omitempty"`

	// ACME ARI Extension
	Replaces *string `json:"replaces,omitempty"`
}

// Account response decoder
func unmarshalOrder(jsonResp []byte, headers http.Header) (order Order, err error) {
	err = json.Unmarshal(jsonResp, &order)
	if err != nil {
		return Order{}, err
	}

	// order location (url) isn't part of the JSON response, add it from the header.
	order.Location = headers.Get("Location")

	return order, nil
}

// NewOrder posts a secure message to the NewOrder URL of the directory
func (service *Service) NewOrder(payload NewOrderPayload, accountKey AccountKey) (order Order, err error) {
	// Strip `replaces` if server doesn't indicate support, per recommendation in ARI s 5
	if payload.Replaces != nil && !service.SupportsARIExtension() {
		payload.Replaces = nil
	}

	// post new-order
	jsonResp, headers, err := service.postToUrlSigned(payload, service.dir.NewOrder, accountKey)
	// if there is an acme.Error of type `malformed` or `alreadyReplaced` AND the `replaces` field is
	// set, strip the replaces filed and do exactly 1 retry
	acmeErr, isAcmeErr := err.(*Error)
	if isAcmeErr && payload.Replaces != nil && acmeErr != nil &&
		(acmeErr.Type == "urn:ietf:params:acme:error:malformed" || acmeErr.Type == "urn:ietf:params:acme:error:alreadyReplaced") {

		payload.Replaces = nil
		jsonResp, headers, err = service.postToUrlSigned(payload, service.dir.NewOrder, accountKey)
	}
	if err != nil {
		return Order{}, err
	}

	// unmarshal response
	order, err = unmarshalOrder(jsonResp, headers)
	if err != nil {
		return Order{}, err
	}

	return order, nil
}

// GetOrder does a POST-as-GET to fetch the current state of the given order URL
func (service *Service) GetOrder(orderUrl string, accountKey AccountKey) (order Order, err error) {
	// POST-as-GET
	jsonResp, headers, err := service.PostAsGet(orderUrl, accountKey)
	if err != nil {
		return Order{}, err
	}

	// unmarshal response
	order, err = unmarshalOrder(jsonResp, headers)
	if err != nil {
		return Order{}, err
	}

	return order, nil
}

// FinalizeOrder posts the specified CSR to the specified finalize URL
func (service *Service) FinalizeOrder(finalizeUrl string, derCsr []byte, accountKey AccountKey) (order Order, err error) {
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
	jsonResp, headers, err := service.postToUrlSigned(payload, finalizeUrl, accountKey)
	if err != nil {
		return Order{}, err
	}

	// unmarshal response
	order, err = unmarshalOrder(jsonResp, headers)
	if err != nil {
		return Order{}, err
	}

	return order, nil
}
