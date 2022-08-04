package acme

import (
	"encoding/json"
	"net/http"
)

// NewOrderPayload is the payload to post to ACME newOrder
type NewOrderPayload struct {
	// notBefore and notAfter are optional and not implemented
	Identifiers []Identifier `json:"identifiers"`
}

// ACME identifier object
type Identifier struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// a slice of identifiers
// allows writing a method for an array of them
type IdentifierSlice []Identifier

// LE response with order information
type Order struct {
	Status         string          `json:"status"`
	Expires        timeString      `json:"expires"`
	Identifiers    IdentifierSlice `json:"identifiers"`
	Error          *Error          `json:"error,omitempty"`
	Authorizations []string        `json:"authorizations"`
	Finalize       string          `json:"finalize"`
	Certificate    string          `json:"certificate,omitempty"`
	NotBefore      timeString      `json:"notBefore,omitempty"`
	NotAfter       timeString      `json:"notAfter,omitempty"`
	Location       string          `json:"-"` // omit because it is in the header
}

// dnsIdentifiers returns a slice of the value strings for a response's
// array of identifier objects that are of type 'dns'
func (ids *IdentifierSlice) DnsIdentifiers() []string {
	var s []string

	for _, id := range *ids {
		if id.Type == "dns" {
			s = append(s, id.Value)
		}
	}

	return s
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
