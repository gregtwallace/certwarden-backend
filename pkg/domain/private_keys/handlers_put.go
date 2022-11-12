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
	ID           int     `json:"-"`
	Name         *string `json:"name"`
	Description  *string `json:"description"`
	ApiKeyViaUrl *bool   `json:"api_key_via_url"`
	UpdatedAt    int     `json:"-"`
}

// PutKeyUpdate updates a Key that already exists in storage.
// Only fields received in the payload (non-nil) are updated.
func (service *Service) PutKeyUpdate(w http.ResponseWriter, r *http.Request) (err error) {
	// parse payload
	var payload UpdatePayload
	err = json.NewDecoder(r.Body).Decode(&payload)
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
	_, err = service.getKey(payload.ID)
	if err != nil {
		return err
	}
	// name (optional - check if not nil)
	if payload.Name != nil && !service.nameValid(*payload.Name, &payload.ID) {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}
	// Description and ApiKeyViaUrl do not need validation
	// end validation

	// add additional details to the payload before saving
	payload.UpdatedAt = int(time.Now().Unix())

	// save updated key info to storage
	err = service.storage.PutKeyUpdate(payload)
	if err != nil {
		service.logger.Error(err)
		return output.ErrStorageGeneric
	}

	// return response to client
	response := output.JsonResponse{
		Status:  http.StatusOK,
		Message: "updated",
		ID:      payload.ID,
	}

	_, err = service.output.WriteJSON(w, response.Status, response, "response")
	if err != nil {
		service.logger.Error(err)
		return output.ErrWriteJsonFailed
	}

	return nil
}
