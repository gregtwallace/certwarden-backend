package private_keys

import (
	"encoding/json"
	"legocerthub-backend/pkg/domain/private_keys/key_crypto"
	"legocerthub-backend/pkg/output"
	"legocerthub-backend/pkg/utils"
	"net/http"
)

// PostPayload is a struct for posting a new key
type NewPayload struct {
	ID             *int    `json:"id"`
	Name           *string `json:"name"`
	Description    *string `json:"description"`
	AlgorithmValue *string `json:"algorithm_value"`
	PemContent     *string `json:"pem"`
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
	// id
	err = utils.IsIdNew(payload.ID)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}
	// name
	err = service.isNameValid(payload.ID, payload.Name)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	/// key add method
	// error if no method specified
	if (payload.AlgorithmValue == nil || *payload.AlgorithmValue == "") && (payload.PemContent == nil || *payload.PemContent == "") {
		service.logger.Debug(ErrBadKey)
		return output.ErrValidationFailed
	}
	// error if more than one method specified
	if (payload.AlgorithmValue != nil && *payload.AlgorithmValue != "") && (payload.PemContent != nil && *payload.PemContent != "") {
		service.logger.Debug(ErrBadKey)
		return output.ErrValidationFailed
	}
	// generate or verify the key
	// generate with algorithm, error if fails
	if payload.AlgorithmValue != nil && *payload.AlgorithmValue != "" {
		// must initialize to avoid invalid address
		payload.PemContent = new(string)
		*payload.PemContent, err = key_crypto.GeneratePrivateKeyPem(*payload.AlgorithmValue)
		if err != nil {
			service.logger.Debug(err)
			return output.ErrValidationFailed
		}
	} else if payload.PemContent != nil && *payload.PemContent != "" {
		// pem inputted - verify pem and determine algorithm
		// must initialize to avoid invalid address
		payload.AlgorithmValue = new(string)
		*payload.PemContent, *payload.AlgorithmValue, err = key_crypto.ValidateKeyPem(*payload.PemContent)
		if err != nil {
			service.logger.Debug(err)
			return output.ErrValidationFailed
		}
	}
	///

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

	_, err = output.WriteJSON(w, response.Status, response, "response")
	if err != nil {
		service.logger.Error(err)
		return output.ErrWriteJsonFailed
	}

	return nil
}
