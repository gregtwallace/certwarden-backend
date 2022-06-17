package private_keys

import (
	"encoding/json"
	"io/ioutil"
	"legocerthub-backend/pkg/utils/payload_utils"
	"net/http"
)

// key payload from PUT/POST
type KeyPayload struct {
	ID             *int    `json:"id"`
	Name           *string `json:"name"`
	Description    *string `json:"description"`
	AlgorithmValue *string `json:"algorithm_value"`
	PemContent     *string `json:"pem"`
}

// key payload from an html form (which only supports strings)
type keyHtmlPayload struct {
	ID             *string `json:"id"`
	Name           *string `json:"name"`
	Description    *string `json:"description"`
	AlgorithmValue *string `json:"algorithm_value"`
	PemContent     *string `json:"pem"`
}

// toPayload converts an html (all string) payload to a properly
// typed payload
func (htmlPayload *keyHtmlPayload) toPayload() (KeyPayload, error) {
	var payload KeyPayload
	var err error

	payload.ID, err = payload_utils.StringToInt(htmlPayload.ID)
	if err != nil {
		return KeyPayload{}, err
	}

	payload.Name = htmlPayload.Name

	payload.Description = htmlPayload.Description

	payload.AlgorithmValue = htmlPayload.AlgorithmValue

	payload.PemContent = htmlPayload.PemContent

	return payload, nil
}

// decodePayload reads a request body and then attempts to
// unmarshal it into a payload. If that fails, it attempts to
// unmarshal into an html (all string) payload and then convert
// to a properly typed payload.
func decodePayload(r *http.Request) (KeyPayload, error) {
	var payload KeyPayload

	// read body to allow for multiple unmarshal attempts
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return KeyPayload{}, err
	}

	// decode into properly typed payload
	err = json.Unmarshal(body, &payload)
	if err != nil {
		// if that failed, decode into string payload
		var htmlPayload keyHtmlPayload
		err = json.Unmarshal(body, &htmlPayload)
		if err != nil {
			return KeyPayload{}, err
		}
		// and then convert it to properly typed payload
		payload, err = htmlPayload.toPayload()
		if err != nil {
			return KeyPayload{}, err
		}
	}

	return payload, nil
}
