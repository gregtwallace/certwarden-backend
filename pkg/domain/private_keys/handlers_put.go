package private_keys

import (
	"encoding/json"
	"legocerthub-backend/pkg/output"
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
)

// UpdatePayload is the struct for editing an existing Key's
// information (only certain fields are editable)
type UpdatePayload struct {
	ID             int     `json:"-"`
	Name           *string `json:"name"`
	Description    *string `json:"description"`
	ApiKey         *string `json:"api_key"`
	ApiKeyNew      *string `json:"api_key_new"`
	ApiKeyDisabled *bool   `json:"api_key_disabled"`
	ApiKeyViaUrl   *bool   `json:"api_key_via_url"`
	UpdatedAt      int     `json:"-"`
}

// PutKeyUpdate updates a Key that already exists in storage.
// Only fields received in the payload (non-nil) are updated.
func (service *Service) PutKeyUpdate(w http.ResponseWriter, r *http.Request) *output.Error {
	// parse payload
	var payload UpdatePayload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// get id param
	idParam := httprouter.ParamsFromContext(r.Context()).ByName("id")
	payload.ID, err = strconv.Atoi(idParam)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// validation
	// id
	_, outErr := service.getKey(payload.ID)
	if outErr != nil {
		return outErr
	}
	// name (optional - check if not nil)
	if payload.Name != nil && !service.NameValid(*payload.Name, &payload.ID) {
		service.logger.Debug(ErrNameBad)
		return output.ErrValidationFailed
	}
	// api key must be at least 10 characters long
	if payload.ApiKey != nil && len(*payload.ApiKey) < 10 {
		service.logger.Debug(ErrApiKeyBad)
		return output.ErrValidationFailed
	}
	// api key new must be at least 10 characters long
	if payload.ApiKeyNew != nil && *payload.ApiKeyNew != "" && len(*payload.ApiKeyNew) < 10 {
		service.logger.Debug(ErrApiKeyNewBad)
		return output.ErrValidationFailed
	}
	// Description, ApiKeyDisabled, and ApiKeyViaUrl do not need validation
	// end validation

	// add additional details to the payload before saving
	payload.UpdatedAt = int(time.Now().Unix())

	// save updated key info to storage
	updatedKey, err := service.storage.PutKeyUpdate(payload)
	if err != nil {
		service.logger.Error(err)
		return output.ErrStorageGeneric
	}

	// write response
	response := &privateKeyResponse{}
	response.StatusCode = http.StatusOK
	response.Message = "updated private key"
	response.PrivateKey = updatedKey.detailedResponse()

	// return response to client
	err = service.output.WriteJSON(w, response)
	if err != nil {
		service.logger.Errorf("failed to write json (%s)", err)
		return output.ErrWriteJsonError
	}

	return nil
}
