package certificates

import (
	"encoding/json"
	"legocerthub-backend/pkg/domain/certificates/challenges"
	"legocerthub-backend/pkg/output"
	"legocerthub-backend/pkg/validation"
	"net/http"
)

// NewPayload is the struct for creating a new certificate
type NewPayload struct {
	ID                   *int     `json:"id"`
	Name                 *string  `json:"name"`
	Description          *string  `json:"description"`
	PrivateKeyID         *int     `json:"private_key_id"`
	AcmeAccountID        *int     `json:"acme_account_id"`
	ChallengeMethodValue *string  `json:"challenge_method_value"`
	Subject              *string  `json:"subject,omitempty"`
	SubjectAltNames      []string `json:"subject_alts,omitempty"`
	CommonName           *string  `json:"common_name,omitempty"`
	Organization         *string  `json:"organization,omitempty"`
	OrganizationalUnit   *string  `json:"organizational_unit,omitempty"`
	Country              *string  `json:"country,omitempty"`
	State                *string  `json:"state,omitempty"`
	City                 *string  `json:"city,omitempty"`
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

	/// do validation
	// id
	err = validation.IsIdNew(payload.ID)
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
	// private key
	err = service.keys.IsPrivateKeyAvailable(payload.PrivateKeyID)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}
	// acme account
	err = service.accounts.IsAcmeAccountValid(payload.AcmeAccountID)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}
	// challenge method
	if payload.ChallengeMethodValue == nil { // check for nil pointer
		service.logger.Debug("missing challenge method value")
		return output.ErrValidationFailed
	}
	_, err = challenges.ChallengeMethodByValue(*payload.ChallengeMethodValue)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}
	// subject
	err = validation.IsDomainValid(payload.Subject)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}
	// subject alts
	// blank is okay, skip validation if not specified
	if payload.SubjectAltNames != nil {
		for _, altName := range payload.SubjectAltNames {
			err = validation.IsDomainValid(&altName)
			if err != nil {
				service.logger.Debug(err)
				return output.ErrValidationFailed
			}
		}
	}
	// TODO: CSR components
	///

	// Save new account details to storage.
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
