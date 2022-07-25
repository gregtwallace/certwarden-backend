package acme

import (
	"encoding/json"
	"net/http"
)

// ACME authorization response
type AuthResponse struct {
	Identifier Identifier     `json:"identifier"` // see orders
	Status     string         `json:"status"`
	Expires    acmeTimeString `json:"expires"`
	Challenges []Challenge    `json:"challenges"`
	Wildcard   bool           `json:"wildcard,omitempty"`
}

// Account response decoder
func unmarshalGetAuthResponse(bodyBytes []byte, headers http.Header) (response AuthResponse, err error) {
	err = json.Unmarshal(bodyBytes, &response)
	if err != nil {
		return AuthResponse{}, err
	}

	return response, nil
}

// GetAuth does a POST-as-GET to feth an authorization object
func (service *Service) GetAuth(authUrl string, accountKey AccountKey) (response AuthResponse, err error) {

	// POST-as-GET
	bodyBytes, headers, err := service.postAsGet(authUrl, accountKey)

	// unmarshal response
	response, err = unmarshalGetAuthResponse(bodyBytes, headers)
	if err != nil {
		return AuthResponse{}, err
	}

	return response, nil
}
