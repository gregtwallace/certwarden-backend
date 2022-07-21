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

// LE response with order information
type OrderResponse struct {
	Status         string         `json:"status"`
	Expires        acmeTimeString `json:"expires"`
	Identifiers    []Identifier   `json:"identifiers"`
	Authorizations []string       `json:"authorizations"`
	Finalize       string         `json:"finalize"`
	Certificate    string         `json:"certificate,omitempty"`
	Location       string         `json:"-"` // omit because it is in the header
	// not implemented
	// NotBefore      acmeTimeString `json:"notBefore"`
	// NotAfter       acmeTimeString `json:"notAfter"`
}

// dnsIdentifiers returns a slice of the value strings for a response's
// array of identifier objects that are of type 'dns'
func (resp *OrderResponse) DnsIdentifiers() []string {
	var s []string

	for _, id := range resp.Identifiers {
		if id.Type == "dns" {
			s = append(s, id.Value)
		}
	}

	return s
}

// Account response decoder
func unmarshalOrderResponse(bodyBytes []byte, headers http.Header) (response OrderResponse, err error) {
	err = json.Unmarshal(bodyBytes, &response)
	if err != nil {
		return OrderResponse{}, err
	}

	// order location (url) isn't part of the JSON response, add it from the header.
	response.Location = headers.Get("Location")

	return response, nil
}

// NewOrder posts a secure message to the NewOrder URL of the directory
func (service *Service) NewOrder(payload NewOrderPayload, accountKey AccountKey) (response OrderResponse, err error) {

	// post new-order
	bodyBytes, headers, err := service.postToUrlSigned(payload, service.dir.NewOrder, accountKey)
	if err != nil {
		return OrderResponse{}, err
	}

	// unmarshal response
	response, err = unmarshalOrderResponse(bodyBytes, headers)
	if err != nil {
		return OrderResponse{}, err
	}

	return response, nil
}
