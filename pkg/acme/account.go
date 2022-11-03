package acme

import (
	"crypto"
	"encoding/json"
	"net/http"
	"strings"
)

// NewAccountPayload is the payload used to post to ACME newAccount
type NewAccountPayload struct {
	TosAgreed bool     `json:"termsOfServiceAgreed"`
	Contact   []string `json:"contact"`
}

// LE response to account data post/update
type Account struct {
	Status    string     `json:"status"`
	Contact   []string   `json:"contact"`
	CreatedAt timeString `json:"createdAt"`
	Location  *string    `json:"-"` // omit because it is in the header
	// -- also available but not in use
	// JsonWebKey jsonWebKey `json:"key"`
	// Orders     string     `json:"orders"`
	// InitialIP  string     `json:"initialIp"`
}

// Account response decoder
func unmarshalAccount(bodyBytes []byte, headers http.Header) (response Account, err error) {
	err = json.Unmarshal(bodyBytes, &response)
	if err != nil {
		return Account{}, err
	}

	// kid isn't part of the JSON response, add it from the header.
	// ACME only returns this if not posting with kid, so have some logic
	// to set it to null if it isn't returned from the server
	if headers.Get("Location") != "" {
		response.Location = new(string)
		*response.Location = headers.Get("Location")
	} else {
		response.Location = nil
	}

	return response, nil
}

// Email() returns an email address from the first string in the Contact slice.
// Any other contacts are discarded.
func (response *Account) Email() string {
	// if contacts are empty, email is blank
	if len(response.Contact) == 0 {
		return ""
	}

	return strings.TrimPrefix(response.Contact[0], "mailto:")
}

// NewAccount posts a secure message to the NewAccount URL of the directory
func (service *Service) NewAccount(payload NewAccountPayload, privateKey crypto.PrivateKey) (response Account, err error) {
	// Create ACME accountKey
	// Register account should never use kid, it must always use JWK
	accountKey := AccountKey{
		Key: privateKey,
	}

	// post new-account
	bodyBytes, headers, err := service.postToUrlSigned(payload, service.dir.NewAccount, accountKey)
	if err != nil {
		return Account{}, err
	}

	// unmarshal response
	response, err = unmarshalAccount(bodyBytes, headers)
	if err != nil {
		return Account{}, err
	}

	return response, nil
}

// UpdateAccountPayload is the payload used to update ACME accounts
type UpdateAccountPayload struct {
	Contact []string `json:"contact,omitempty"`
	Status  string   `json:"status,omitempty"`
}

// UpdateAccount posts a secure message to the kid of the account
// initially support only exists to update the email address
// TODO: key rotation and account deactivation
func (service *Service) UpdateAccount(payload UpdateAccountPayload, accountKey AccountKey) (response Account, err error) {

	// post account update
	bodyBytes, headers, err := service.postToUrlSigned(payload, accountKey.Kid, accountKey)
	if err != nil {
		return Account{}, err
	}

	// unmarshal response
	response, err = unmarshalAccount(bodyBytes, headers)
	if err != nil {
		return Account{}, err
	}

	return response, nil
}

// DeactivateAccount posts deactivated status to the ACME account
// Once deactivated, accounts cannot be re-enabled. This action is DANGEROUS
// and should only be done when there is a complete understanding of the repurcussions.
func (service *Service) DeactivateAccount(accountKey AccountKey) (response Account, err error) {
	// deactivate payload is always the same
	payload := UpdateAccountPayload{
		Status: "deactivated",
	}

	// post account update
	bodyBytes, headers, err := service.postToUrlSigned(payload, accountKey.Kid, accountKey)
	if err != nil {
		return Account{}, err
	}

	// unmarshal response
	response, err = unmarshalAccount(bodyBytes, headers)
	if err != nil {
		return Account{}, err
	}

	return response, nil
}

// RolloverAccountKey rolls over the specified account's key to the newKey. This essentially
// retires the old key from the account and substitutes the new key in its place.
func (service *Service) RolloverAccountKey(newKey crypto.PrivateKey, oldAccountKey AccountKey) (response Account, err error) {
	// build payload
	payload := acmeSignedMessage{}

	// Create ACME accountKey for new key, never use kid always use JWK
	newKeyAccountKey := AccountKey{
		Key: newKey,
	}

	// inner (payload's) header
	var innerHeader protectedHeader

	// new key's alg
	innerHeader.Algorithm, err = newKeyAccountKey.signingAlg()
	if err != nil {
		return Account{}, err
	}

	// new key's jwk
	innerHeader.JsonWebKey, err = newKeyAccountKey.jwk()
	if err != nil {
		return Account{}, err
	}

	// omit kid
	// omit nonce

	// url
	innerHeader.Url = service.dir.KeyChange

	// encode and add to payload
	payload.ProtectedHeader, err = encodeJson(innerHeader)
	if err != nil {
		return Account{}, err
	}
	// end inner (payload's) header

	// inner (payload's) payload
	// old key's jwk
	oldJwk, err := oldAccountKey.jwk()
	if err != nil {
		return Account{}, err
	}

	// build inner payload
	innerPayload := struct {
		AccountUrl string     `json:"account"`
		OldKeyJwk  jsonWebKey `json:"oldKey"`
	}{
		AccountUrl: oldAccountKey.Kid,
		OldKeyJwk:  *oldJwk,
	}

	// encode and add to payload
	payload.Payload, err = encodeJson(innerPayload)
	if err != nil {
		return Account{}, err
	}
	// end inner (payload's) payload

	// inner (payload's) signature
	payload.Signature, err = newKeyAccountKey.Sign(payload)
	if err != nil {
		return Account{}, err
	}
	// end inner (payload's) signature

	// post key change
	bodyBytes, headers, err := service.postToUrlSigned(payload, service.dir.KeyChange, oldAccountKey)
	if err != nil {
		return Account{}, err
	}

	// unmarshal response
	response, err = unmarshalAccount(bodyBytes, headers)
	if err != nil {
		return Account{}, err
	}

	return response, nil
}
