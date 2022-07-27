package acme

import (
	"encoding/json"
	"net/http"
)

// ACME authorization response
type Authorization struct {
	Identifier Identifier  `json:"identifier"` // see orders
	Status     string      `json:"status"`
	Expires    timeString  `json:"expires"`
	Challenges []Challenge `json:"challenges"`
	Wildcard   bool        `json:"wildcard,omitempty"`
}

// Account response decoder
func unmarshalAuthorization(bodyBytes []byte, headers http.Header) (response Authorization, err error) {
	err = json.Unmarshal(bodyBytes, &response)
	if err != nil {
		return Authorization{}, err
	}

	return response, nil
}

// GetAuth does a POST-as-GET to feth an authorization object
func (service *Service) GetAuth(authUrl string, accountKey AccountKey) (response Authorization, err error) {

	// POST-as-GET
	bodyBytes, headers, err := service.postAsGet(authUrl, accountKey)
	if err != nil {
		return Authorization{}, err
	}

	// unmarshal response
	response, err = unmarshalAuthorization(bodyBytes, headers)
	if err != nil {
		return Authorization{}, err
	}

	return response, nil
}
