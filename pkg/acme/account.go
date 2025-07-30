package acme

import (
	"crypto"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

// NewAccountPayload is the payload the caller populates to post a new
// ACME account. This is NOT the payload that is ACTUALLY posted to ACME
type NewAccountPayload struct {
	Contact                       []string
	TosAgreed                     bool
	ExternalAccountBindingKid     string
	ExternalAccountBindingHmacKey string
}

// acmeNewAccountPayload is the newAccount payload that is actually posted
// to ACME
type acmeNewAccountPayload struct {
	Contact                []string           `json:"contact"`
	TosAgreed              bool               `json:"termsOfServiceAgreed"`
	ExternalAccountBinding *acmeSignedMessage `json:"externalAccountBinding,omitempty"`
}

// LE response to account data post/update
type Account struct {
	Status    string     `json:"status"`
	Contact   []string   `json:"contact"`
	CreatedAt timeString `json:"createdAt,omitempty"` // non-standard field
	Location  *string    `json:"-"`                   // omit because it is in the header
	// -- also available but not in use
	// JsonWebKey jsonWebKey `json:"key"`
	// Orders     string     `json:"orders"`
	// InitialIP  string     `json:"initialIp"`
}

// Account response decoder
func unmarshalAccount(jsonResp json.RawMessage, headers http.Header) (acct Account, err error) {
	err = json.Unmarshal(jsonResp, &acct)
	if err != nil {
		return Account{}, err
	}

	// kid isn't part of the JSON response, add it from the header.
	// ACME only returns this if not posting with kid, so have some logic
	// to set it to null if it isn't returned from the server
	if headers.Get("Location") != "" {
		acct.Location = new(string)
		*acct.Location = headers.Get("Location")
	} else {
		acct.Location = nil
	}

	return acct, nil
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
func (service *Service) NewAccount(payload NewAccountPayload, privateKey crypto.PrivateKey) (acct Account, err error) {
	// Create ACME accountKey
	// Register account should never use kid, it must always use JWK
	accountKey := AccountKey{
		Key: privateKey,
	}

	// url to post to
	url := service.dir.NewAccount

	// real ACME payload
	acmePayload := &acmeNewAccountPayload{
		Contact:   payload.Contact,
		TosAgreed: payload.TosAgreed,
	}

	// if External Account Binding is present, modify payload accordingly
	if payload.ExternalAccountBindingKid != "" && payload.ExternalAccountBindingHmacKey != "" {
		acmePayload.ExternalAccountBinding, err = accountKey.makeExternalAccountBinding(payload.ExternalAccountBindingKid, payload.ExternalAccountBindingHmacKey, url)
		if err != nil {
			return Account{}, err
		}
	}

	// post new-account
	jsonResp, headers, err := service.postToUrlSigned(acmePayload, url, accountKey)
	if err != nil {
		return Account{}, err
	}

	// unmarshal response
	acct, err = unmarshalAccount(jsonResp, headers)
	if err != nil {
		return Account{}, err
	}

	return acct, nil
}

// GetAccount does a POST-as-GET to fetch the current state of the given accountKey's Account
func (service *Service) GetAccount(accountKey AccountKey) (acct Account, err error) {
	// POST-as-GET
	jsonResp, headers, err := service.PostAsGet(accountKey.Kid, accountKey)
	if err != nil {
		return Account{}, err
	}

	// unmarshal response
	acct, err = unmarshalAccount(jsonResp, headers)
	if err != nil {
		return Account{}, err
	}

	return acct, nil
}

// UpdateAccountPayload is the payload used to update ACME accounts
type UpdateAccountPayload struct {
	Contact []string `json:"contact,omitempty"`
	Status  string   `json:"status,omitempty"`
}

// UpdateAccount posts a secure message to the kid of the account
// initially support only exists to update the email address
func (service *Service) UpdateAccount(payload UpdateAccountPayload, accountKey AccountKey) (acct Account, err error) {
	// post account update
	jsonResp, headers, err := service.postToUrlSigned(payload, accountKey.Kid, accountKey)
	if err != nil {
		return Account{}, err
	}

	// unmarshal response
	acct, err = unmarshalAccount(jsonResp, headers)
	if err != nil {
		return Account{}, err
	}

	return acct, nil
}

// DeactivateAccount posts deactivated status to the ACME account
// Once deactivated, accounts cannot be re-enabled. This action is DANGEROUS
// and should only be done when there is a complete understanding of the repurcussions.
func (service *Service) DeactivateAccount(accountKey AccountKey) (acct Account, err error) {
	// deactivate payload is always the same
	payload := UpdateAccountPayload{
		Status: "deactivated",
	}

	// post account update
	jsonResp, headers, err := service.postToUrlSigned(payload, accountKey.Kid, accountKey)
	if err != nil {
		return Account{}, err
	}

	// unmarshal response
	acct, err = unmarshalAccount(jsonResp, headers)
	if err != nil {
		return Account{}, err
	}

	return acct, nil
}

// RolloverAccountKey rolls over the specified account's key to the newKey. This essentially
// retires the old key from the account and substitutes the new key in its place.
func (service *Service) RolloverAccountKey(newKey crypto.PrivateKey, oldAccountKey AccountKey) (err error) {
	// if directory doesn't contain a keyChange URL, return error
	if service.dir.KeyChange == "" {
		return errors.New("acme: account key rollover failed (acme directory does not contain a keyChange url)")
	}

	// build payload (which is itself an acme signed message with some special properties)
	// Create ACME accountKey for new key, ALWAYS use JWK (NEVER kid)
	newKeyAccountKey := AccountKey{
		Key: newKey,
	}

	// payload of the payload is this special struct
	oldJwk, err := oldAccountKey.jwk()
	if err != nil {
		return err
	}

	innerPayload := struct {
		AccountUrl string     `json:"account"`
		OldKeyJwk  jsonWebKey `json:"oldKey"`
	}{
		AccountUrl: oldAccountKey.Kid,
		OldKeyJwk:  *oldJwk,
	}

	// build payload message with new key
	msg, err := makeAcmeSignedMessage(innerPayload, "", service.dir.KeyChange, newKeyAccountKey)
	if err != nil {
		return err
	}

	// post key change
	// no response/headers expected on key roll (see rfc 8555 s 7.3.5)
	_, _, err = service.postToUrlSigned(msg, service.dir.KeyChange, oldAccountKey)
	if err != nil {
		return err
	}

	return nil
}
