package private_keys

import (
	"encoding/json"
	"legocerthub-backend/pkg/output"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// NameDescPayload is the struct for editing an existing key
type NameDescPayload struct {
	ID          *int    `json:"id"`
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

// PutExistingKey updates a key that already exists in storage.
// only the name and description are allowed to be modified
func (service *Service) PutNameDescKey(w http.ResponseWriter, r *http.Request) (err error) {
	idParamStr := httprouter.ParamsFromContext(r.Context()).ByName("id")

	// convert id param to an integer
	idParam, err := strconv.Atoi(idParamStr)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	var payload NameDescPayload
	err = json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	/// validation
	// id
	err = service.isIdExistingMatch(idParam, payload.ID)
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
	///

	// save key name and desc to storage, which also returns the key id with new
	// name and description
	err = service.storage.PutNameDescKey(payload)
	if err != nil {
		service.logger.Error(err)
		return output.ErrStorageGeneric
	}

	// return response to client
	response := output.JsonResponse{
		Status:  http.StatusOK,
		Message: "updated",
		ID:      idParam,
	}

	_, err = service.output.WriteJSON(w, response.Status, response, "response")
	if err != nil {
		service.logger.Error(err)
		return output.ErrWriteJsonFailed
	}

	return nil
}
