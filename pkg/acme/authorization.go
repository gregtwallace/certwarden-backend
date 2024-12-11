package acme

import (
	"encoding/json"
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
func unmarshalAuthorization(jsonResp json.RawMessage) (auth Authorization, err error) {
	err = json.Unmarshal(jsonResp, &auth)
	if err != nil {
		return Authorization{}, err
	}

	return auth, nil
}

// GetAuth does a POST-as-GET to fetch an authorization object
func (service *Service) GetAuth(authUrl string, accountKey AccountKey) (auth Authorization, err error) {

	// POST-as-GET
	jsonResp, _, err := service.PostAsGet(authUrl, accountKey)
	if err != nil {
		return Authorization{}, err
	}

	// unmarshal response
	auth, err = unmarshalAuthorization(jsonResp)
	if err != nil {
		return Authorization{}, err
	}

	return auth, nil
}
