package certificates

import (
	"encoding/json"
	"legocerthub-backend/pkg/challenges"
	"legocerthub-backend/pkg/output"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// DetailsUpdatePayload is the struct for editing an existing cert. A number of
// fields can be updated by the client on the fly (without ACME interaction).
type DetailsUpdatePayload struct {
	ID                   *int    `json:"id"`
	Name                 *string `json:"name"`
	Description          *string `json:"description"`
	ChallengeMethodValue *string `json:"challenge_method_value"`
	CommonName           *string `json:"common_name"`
	Organization         *string `json:"organization"`
	OrganizationalUnit   *string `json:"organizational_unit"`
	Country              *string `json:"country"`
	State                *string `json:"state"`
	City                 *string `json:"city"`
}

// PutDetailsCert is a handler that sets various details about a cert and saves
// them to storage. These are all details that should be editable any time.
func (service *Service) PutDetailsCert(w http.ResponseWriter, r *http.Request) (err error) {
	idParamStr := httprouter.ParamsFromContext(r.Context()).ByName("certid")

	// convert id param to an integer
	idParam, err := strconv.Atoi(idParamStr)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// payload decoding
	var payload DetailsUpdatePayload
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
	// name (optional)
	if payload.Name != nil {
		err = service.isNameValid(payload.ID, payload.Name)
		if err != nil {
			service.logger.Debug(err)
			return output.ErrValidationFailed
		}
	}
	// challenge method (optional)
	if payload.ChallengeMethodValue != nil {
		_, err = challenges.MethodByValue(*payload.ChallengeMethodValue)
		if err != nil {
			service.logger.Debug(err)
			return output.ErrValidationFailed
		}
	}
	// TODO: CSR detail validation
	///

	// save account name and desc to storage, which also returns the account id with new
	// name and description
	err = service.storage.PutDetailsCert(payload)
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
