package acme_accounts

import (
	"encoding/json"
	"io/ioutil"
	"legocerthub-backend/pkg/utils/payload_utils"
	"net/http"
)

// acme account payload from PUT/POST
type AccountPayload struct {
	ID           *int    `json:"id"`
	Name         *string `json:"name"`
	Description  *string `json:"description"`
	Email        *string `json:"email"`
	PrivateKeyID *int    `json:"private_key_id"`
	AcceptedTos  *bool   `json:"accepted_tos"`
	IsStaging    *bool   `json:"is_staging"`
}

// acme account payload from an html form (which only supports strings)
type accountHtmlPayload struct {
	ID           *string `json:"id"`
	Name         *string `json:"name"`
	Description  *string `json:"description"`
	Email        *string `json:"email"`
	PrivateKeyID *string `json:"private_key_id"`
	AcceptedTos  *string `json:"accepted_tos"`
	IsStaging    *string `json:"is_staging"`
}

// toPayload converts an html (all string) payload to a properly
// typed payload
func (htmlPayload *accountHtmlPayload) toPayload() (AccountPayload, error) {
	var payload AccountPayload
	var err error

	payload.ID, err = payload_utils.StringToInt(htmlPayload.ID)
	if err != nil {
		return AccountPayload{}, err
	}

	payload.Name = htmlPayload.Name

	payload.Description = htmlPayload.Description

	payload.Email = htmlPayload.Email

	payload.PrivateKeyID, err = payload_utils.StringToInt(htmlPayload.PrivateKeyID)
	if err != nil {
		return AccountPayload{}, err
	}

	payload.AcceptedTos = payload_utils.StringToBool(htmlPayload.AcceptedTos)

	payload.IsStaging = payload_utils.StringToBool(htmlPayload.IsStaging)

	return payload, nil
}

// decodePayload reads a request body and then attempts to
// unmarshal it into a payload. If that fails, it attempts to
// unmarshal into an html (all string) payload and then convert
// to a properly typed payload.
func decodePayload(r *http.Request) (AccountPayload, error) {
	var payload AccountPayload

	// read body to allow for multiple unmarshal attempts
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return AccountPayload{}, err
	}

	// decode into properly typed payload
	err = json.Unmarshal(body, &payload)
	if err != nil {
		// if that failed, decode into string payload
		var htmlPayload accountHtmlPayload
		err = json.Unmarshal(body, &htmlPayload)
		if err != nil {
			return AccountPayload{}, err
		}
		// and then convert it to properly typed payload
		payload, err = htmlPayload.toPayload()
		if err != nil {
			return AccountPayload{}, err
		}
	}

	return payload, nil
}
