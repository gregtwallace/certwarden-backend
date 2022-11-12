package certificates

import (
	"encoding/json"
	"legocerthub-backend/pkg/challenges"
	"legocerthub-backend/pkg/output"
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
)

// DetailsUpdatePayload is the struct for editing an existing cert. A number of
// fields can be updated by the client on the fly (without ACME interaction).
type DetailsUpdatePayload struct {
	ID                   int      `json:"-"`
	Name                 *string  `json:"name"`
	Description          *string  `json:"description"`
	PrivateKeyId         *int     `json:"private_key_id"`
	ChallengeMethodValue *string  `json:"challenge_method_value"`
	SubjectAltNames      []string `json:"subject_alts"`
	Organization         *string  `json:"organization"`
	OrganizationalUnit   *string  `json:"organizational_unit"`
	Country              *string  `json:"country"`
	State                *string  `json:"state"`
	City                 *string  `json:"city"`
	ApiKeyViaUrl         *bool    `json:"api_key_via_url"`
	UpdatedAt            int      `json:"-"`
}

// PutDetailsCert is a handler that sets various details about a cert and saves
// them to storage. These are all details that should be editable any time.
func (service *Service) PutDetailsCert(w http.ResponseWriter, r *http.Request) (err error) {
	// payload decoding
	var payload DetailsUpdatePayload
	err = json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// get id from param
	idParam := httprouter.ParamsFromContext(r.Context()).ByName("certid")
	payload.ID, err = strconv.Atoi(idParam)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// validation
	// id
	_, err = service.GetCertificate(payload.ID)
	if err != nil {
		return err
	}
	// name (optional)
	if payload.Name != nil && !service.nameValid(*payload.Name, &payload.ID) {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}
	// description - no validation
	// private key (optional)
	if payload.PrivateKeyId != nil && !service.privateKeyIdValid(*payload.PrivateKeyId, &payload.ID) {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}
	// challenge method (optional)
	if payload.ChallengeMethodValue != nil && challenges.MethodByStorageValue(*payload.ChallengeMethodValue) == challenges.UnknownMethod {
		service.logger.Debug("unknown challenge method")
		return output.ErrValidationFailed
	}
	// subject alts (optional)
	if payload.SubjectAltNames != nil && !subjectAltsValid(payload.SubjectAltNames) {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}
	// TODO: Do any validation of CSR components?
	// end validation

	// add additional details to the payload before saving
	payload.UpdatedAt = int(time.Now().Unix())

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
		ID:      payload.ID,
	}

	_, err = service.output.WriteJSON(w, response.Status, response, "response")
	if err != nil {
		service.logger.Error(err)
		return output.ErrWriteJsonFailed
	}

	return nil
}
