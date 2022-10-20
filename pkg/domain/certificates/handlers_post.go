package certificates

import (
	"encoding/json"
	"legocerthub-backend/pkg/challenges"
	"legocerthub-backend/pkg/output"
	"legocerthub-backend/pkg/randomness"
	"legocerthub-backend/pkg/validation"
	"net/http"
	"time"
)

// NewPayload is the struct for creating a new certificate
type NewPayload struct {
	Name                 *string  `json:"name"`
	Description          *string  `json:"description"`
	PrivateKeyID         *int     `json:"private_key_id"`
	AcmeAccountID        *int     `json:"acme_account_id"`
	ChallengeMethodValue *string  `json:"challenge_method_value"`
	Subject              *string  `json:"subject"`
	SubjectAltNames      []string `json:"subject_alts"`
	Organization         *string  `json:"organization"`
	OrganizationalUnit   *string  `json:"organizational_unit"`
	Country              *string  `json:"country"`
	State                *string  `json:"state"`
	City                 *string  `json:"city"`
	ApiKey               string   `json:"-"`
	ApiKeyViaUrl         bool     `json:"-"`
	CreatedAt            int      `json:"-"`
	UpdatedAt            int      `json:"-"`
}

// PostNewCert creates a new certificate object in storage. No actual encryption certificate
// is generated, this only stores the needed information to later contact ACME and acquire
// the cert.
func (service *Service) PostNewCert(w http.ResponseWriter, r *http.Request) (err error) {
	var payload NewPayload

	// decode body into payload
	err = json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// validation
	// name
	if payload.Name == nil || !service.nameValid(*payload.Name, nil) {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}
	// description (if none, set to blank)
	if payload.Description == nil {
		payload.Description = new(string)
	}
	// private key
	if payload.PrivateKeyID == nil || !service.privateKeyIdValid(*payload.PrivateKeyID, nil) {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}
	// acme account
	if payload.AcmeAccountID == nil || !service.accounts.AccountUsable(*payload.AcmeAccountID) {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}
	// challenge method
	if payload.ChallengeMethodValue == nil || challenges.MethodByValue(*payload.ChallengeMethodValue) == challenges.UnknownMethod {
		service.logger.Debug("unknown challenge method")
		return output.ErrValidationFailed
	}
	// subject
	if payload.Subject == nil || !validation.DomainValid(*payload.Subject) {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}
	// subject alts
	// blank is okay, skip validation if not specified
	if payload.SubjectAltNames != nil && subjectAltsValid(payload.SubjectAltNames) {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}
	// CSR
	// set to blank if don't exist
	// TODO: Do any validation of CSR components?
	if payload.Organization == nil {
		payload.Organization = new(string)
	}
	if payload.OrganizationalUnit == nil {
		payload.OrganizationalUnit = new(string)
	}
	if payload.Country == nil {
		payload.Country = new(string)
	}
	if payload.State == nil {
		payload.State = new(string)
	}
	if payload.City == nil {
		payload.City = new(string)
	}
	// end validation

	// add additional details to the payload before saving
	payload.ApiKey, err = randomness.GenerateApiKey()
	if err != nil {
		service.logger.Error(err)
		return output.ErrInternal
	}
	payload.ApiKeyViaUrl = false
	payload.CreatedAt = int(time.Now().Unix())
	payload.UpdatedAt = payload.CreatedAt

	// Save new account details to storage
	// No ACME actions are performed.
	id, err := service.storage.PostNewCert(payload)
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
