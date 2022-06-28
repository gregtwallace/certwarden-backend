package acme

import (
	"crypto"
	"encoding/json"
	"strings"
)

type NewAccountPayload struct {
	TosAgreed bool     `json:"termsOfServiceAgreed"`
	Contact   []string `json:"contact"`
}

// LE response to account data post/update
type AcmeNewAccountResponse struct {
	Contact   []string `json:"contact"`
	CreatedAt string   `json:"createdAt"`
	Status    string   `json:"status"`
	Location  string   `json:"-"` // omit because it is in the header
	// -- also available but not in use
	// JsonWebKey jsonWebKey `json:"key"`
	// Orders     string     `json:"orders"`
	// InitialIP  string     `json:"initialIp"`
}

// CreatedAt() returns the created at time in unix format. If there is an error
// converting, return 0
func (response *AcmeNewAccountResponse) CreatedAtUnix() (int, error) {
	time, err := acmeToUnixTime(response.CreatedAt)
	if err != nil {
		return 0, err
	}

	return time, nil
}

// Email() returns an email address from the first string in the Contact slice.
// Any other contacts are discarded.
func (response *AcmeNewAccountResponse) Email() string {
	// if contacts are empty, email is blank
	if len(response.Contact) == 0 {
		return ""
	}

	return strings.TrimPrefix(response.Contact[0], "mailto:")
}

// RegisterAccount posts a secure message to the NewAccount URL of the directory
func (service *Service) NewAccount(payload NewAccountPayload, privateKey crypto.PrivateKey) (response AcmeNewAccountResponse, err error) {
	// Create ACME accountKey
	// Register account should never use kid, it must always use JWK
	var accountKey AccountKey
	accountKey.Key = privateKey
	accountKey.Kid = "" // no-op, just a reminder

	// post new-account
	bodyBytes, headers, err := service.postToUrlSigned(payload, service.dir.NewAccount, accountKey)
	if err != nil {
		return AcmeNewAccountResponse{}, err
	}

	// try to decode an error
	var errorResponse AcmeErrorResponse
	err = json.Unmarshal(bodyBytes, &errorResponse)
	if err == nil {
		// return error if acme response was an error
		return AcmeNewAccountResponse{}, errorResponse.Error()
	} else {
		// if error didn't decode, decode generally
		err = json.Unmarshal(bodyBytes, &response)
		if err != nil {
			return AcmeNewAccountResponse{}, err
		}
	}

	// kid isn't part of the JSON response, add it from the header
	response.Location = headers.Get("Location")

	return response, nil
}
