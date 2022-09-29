package private_keys

import (
	"encoding/json"
	"legocerthub-backend/pkg/domain/private_keys/key_crypto"
	"legocerthub-backend/pkg/output"
	"legocerthub-backend/pkg/randomness"
	"legocerthub-backend/pkg/validation"
	"net/http"
	"time"
)

// PostPayload is a struct for posting a new key
type NewPayload struct {
	Name           *string `json:"name"`
	Description    *string `json:"description"`
	AlgorithmValue *string `json:"algorithm_value"`
	PemContent     *string `json:"pem"`
	ApiKey         string  `json:"-"`
	ApiKeyViaUrl   bool    `json:"-"`
	CreatedAt      int     `json:"-"`
	UpdatedAt      int     `json:"-"`
}

// PostNewKey creates a new private key and saves it to storage
func (service *Service) PostNewKey(w http.ResponseWriter, r *http.Request) (err error) {
	var payload NewPayload

	// decode body into payload
	err = json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	/// do validation
	// name (missing or invalid)
	if payload.Name == nil || !service.nameValid(*payload.Name, nil) {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}
	// description (if none, set to blank)
	if payload.Description == nil {
		payload.Description = new(string)
	}
	// key add method
	// error if no method specified
	if (payload.AlgorithmValue == nil || *payload.AlgorithmValue == "") && (payload.PemContent == nil || *payload.PemContent == "") {
		service.logger.Debug(validation.ErrKeyBadOption)
		return output.ErrValidationFailed
	}
	// error if more than one method specified
	if (payload.AlgorithmValue != nil && *payload.AlgorithmValue != "") && (payload.PemContent != nil && *payload.PemContent != "") {
		service.logger.Debug(validation.ErrKeyBadOption)
		return output.ErrValidationFailed
	}
	// generate or verify the key
	// generate with algorithm, error if fails
	if payload.AlgorithmValue != nil && *payload.AlgorithmValue != "" {
		// must initialize to avoid invalid address
		payload.PemContent = new(string)
		*payload.PemContent, err = key_crypto.AlgorithmByValue(*payload.AlgorithmValue).GeneratePrivateKeyPem()
		if err != nil {
			service.logger.Debug(err)
			return output.ErrValidationFailed
		}
	} else if payload.PemContent != nil && *payload.PemContent != "" {
		// pem inputted - verify pem and determine algorithm
		// must initialize to avoid invalid address
		payload.AlgorithmValue = new(string)
		var alg key_crypto.Algorithm
		*payload.PemContent, alg, err = key_crypto.ValidateKeyPem(*payload.PemContent)
		*payload.AlgorithmValue = alg.Value()
		if err != nil {
			service.logger.Debug(err)
			return output.ErrValidationFailed
		}
	}
	// end key add method

	// add additional details to the payload before saving
	payload.ApiKey, err = randomness.GenerateApiKey()
	if err != nil {
		service.logger.Error(err)
		return output.ErrInternal
	}
	payload.ApiKeyViaUrl = false
	payload.CreatedAt = int(time.Now().Unix())
	payload.UpdatedAt = payload.CreatedAt

	// save new key to storage, which also returns the new key id
	id, err := service.storage.PostNewKey(payload)
	if err != nil {
		service.logger.Error(err)
		return output.ErrStorageGeneric
	}

	// return response to client
	response := output.JsonResponse{
		Status:  http.StatusCreated,
		Message: "created",
		ID:      id,
	}

	_, err = service.output.WriteJSON(w, response.Status, response, "response")
	if err != nil {
		service.logger.Error(err)
		return output.ErrWriteJsonFailed
	}

	return nil
}
