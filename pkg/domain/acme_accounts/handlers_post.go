package acme_accounts

import (
	"certwarden-backend/pkg/output"
	"certwarden-backend/pkg/validation"
	"encoding/json"
	"net/http"
	"time"
)

// NewPayload is the payload struct for creating a new account
type NewPayload struct {
	Name         *string `json:"name"`
	Description  *string `json:"description"`
	AcmeServerID *int    `json:"acme_server_id"`
	PrivateKeyID *int    `json:"private_key_id"`
	Status       string  `json:"-"`
	Email        *string `json:"email"`
	AcceptedTos  *bool   `json:"accepted_tos"`
	CreatedAt    int     `json:"-"`
	UpdatedAt    int     `json:"-"`
	Kid          string  `json:"-"`
}

// PostNewAccount is the handler to save a new account to storage. No ACME
// actions (e.g. registration) are taken.
func (service *Service) PostNewAccount(w http.ResponseWriter, r *http.Request) *output.Error {
	var payload NewPayload

	// decode body into payload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// validation
	// name
	if payload.Name == nil || !service.nameValid(*payload.Name, nil) {
		service.logger.Debug(ErrNameBad)
		return output.ErrValidationFailed
	}

	// description (blank if not specified)
	if payload.Description == nil {
		payload.Description = new(string)
		*payload.Description = ""
	}

	// email (make blank if not specified)
	if payload.Email == nil {
		payload.Email = new(string)
		*payload.Email = ""
	} else if !validation.EmailValidOrBlank(*payload.Email) {
		service.logger.Debug(ErrEmailBad)
		return output.ErrValidationFailed
	}

	// TOS must be accepted
	if payload.AcceptedTos == nil || !*payload.AcceptedTos {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// ACME Server
	if payload.AcmeServerID == nil || !service.acmeServerService.AcmeServerValid(*payload.AcmeServerID) {
		service.logger.Debug("acme_server_id not specified or invalid")
		return output.ErrValidationFailed
	}

	// private key (make last since most intense op)
	if payload.PrivateKeyID == nil || !service.keys.KeyAvailable(*payload.PrivateKeyID) {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}
	// end validation

	// add additional details to the payload before saving
	payload.Status = "unknown"
	payload.CreatedAt = int(time.Now().Unix())
	payload.UpdatedAt = payload.CreatedAt
	payload.Kid = ""

	// Save new account details to storage.
	// No ACME actions are performed.
	newAcct, err := service.storage.PostNewAccount(payload)
	if err != nil {
		service.logger.Error(err)
		return output.ErrStorageGeneric
	}

	detailedResp, err := newAcct.detailedResponse(service)
	if err != nil {
		service.logger.Errorf("failed to generate account summary response (%s)", err)
		return output.ErrInternal
	}

	// write response
	response := &accountResponse{}
	response.StatusCode = http.StatusCreated
	response.Message = "created account"
	response.Account = detailedResp

	err = service.output.WriteJSON(w, response)
	if err != nil {
		service.logger.Errorf("failed to write json (%s)", err)
		return output.ErrWriteJsonError
	}

	return nil
}
