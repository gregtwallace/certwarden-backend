package certificates

import (
	"certwarden-backend/pkg/output"
	"certwarden-backend/pkg/validation"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
)

// UpdatePayload is the struct for editing an existing cert. A number of
// fields can be updated by the client on the fly (without ACME interaction).
type UpdatePayload struct {
	ID                          int             `json:"-"`
	Name                        *string         `json:"name"`
	Description                 *string         `json:"description"`
	PrivateKeyId                *int            `json:"private_key_id"`
	SubjectAltNames             []string        `json:"subject_alts"`
	Organization                *string         `json:"organization"`
	OrganizationalUnit          *string         `json:"organizational_unit"`
	Country                     *string         `json:"country"`
	State                       *string         `json:"state"`
	City                        *string         `json:"city"`
	CSRExtraExtensions          []CertExtension `json:"csr_extra_extensions"`
	PreferredRootCN             *string         `json:"preferred_root_cn"`
	PostProcessingCommand       *string         `json:"post_processing_command"`
	PostProcessingEnvironment   []string        `json:"post_processing_environment"`
	PostProcessingClientAddress *string         `json:"post_processing_client_address"`
	PostProcessingClientKeyB64  *string         `json:"post_processing_client_key"`
	Profile                     *string         `json:"profile"`
	ApiKey                      *string         `json:"api_key"`
	ApiKeyNew                   *string         `json:"api_key_new"`
	ApiKeyViaUrl                *bool           `json:"api_key_via_url"`
	UpdatedAt                   time.Time       `json:"-"`
}

// PutDetailsCert is a handler that sets various details about a cert and saves
// them to storage. These are all details that should be editable any time.
func (service *Service) PutDetailsCert(w http.ResponseWriter, r *http.Request) *output.JsonError {
	// payload decoding
	var payload UpdatePayload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrorJsonErrValidationFailed(err)
	}

	// get id from param
	idParam := httprouter.ParamsFromContext(r.Context()).ByName("certid")
	payload.ID, err = strconv.Atoi(idParam)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrorJsonErrValidationFailed(err)
	}

	// validation
	// id
	cert, outErr := service.GetCertificate(payload.ID)
	if outErr != nil {
		service.logger.Debug(ErrIdBad)
		return output.ErrorJsonErrValidationFailed(ErrIdBad)
	}
	// name (optional)
	if payload.Name != nil && !service.nameValid(*payload.Name, &payload.ID) {
		service.logger.Debug(ErrNameBad)
		return output.ErrorJsonErrValidationFailed(ErrNameBad)
	}
	// description - no validation
	// private key (optional)
	if payload.PrivateKeyId != nil && !service.privateKeyIdValid(*payload.PrivateKeyId, &payload.ID) {
		return output.ErrorJsonErrValidationFailed(ErrKeyIdBad)
	}
	// subject alts (optional)
	// if new alts are being specified
	if payload.SubjectAltNames != nil {
		if !subjectAltsValid(payload.SubjectAltNames) {
			service.logger.Debug(ErrDomainBad)
			return output.ErrorJsonErrValidationFailed(ErrDomainBad)
		}

	} else if len(cert.SubjectAltNames) > 0 {
		// if keeping old alts and they exist (more than 0)
		// verify against the challenge method
		if !subjectAltsValid(cert.SubjectAltNames) {
			service.logger.Debug(ErrDomainBad)
			return output.ErrorJsonErrValidationFailed(ErrDomainBad)
		}
	}
	// profile Extension -- validate if specified
	if payload.Profile != nil && *payload.Profile != "" {
		// specified, validate against acme service
		acmeService, err := service.acmeServerService.AcmeService(cert.Account.AcmeServer.ID)
		if err != nil {
			err = fmt.Errorf("failed to retrieve acme service (%s)", err)
			service.logger.Error(err)
			return output.ErrorJsonErrInternal(err)
		}
		if !acmeService.ProfileValidate(*payload.Profile) {
			err = fmt.Errorf("acme service for specified account does not advertise profile `%s`", *payload.Profile)
			service.logger.Debug(err)
			return output.ErrorJsonErrValidationFailed(err)
		}
	}
	// api key must be at least 10 characters long
	if payload.ApiKey != nil && len(*payload.ApiKey) < 10 {
		service.logger.Debug(ErrApiKeyBad)
		return output.ErrorJsonErrValidationFailed(ErrApiKeyBad)
	}
	// api key new must be at least 10 characters long
	if payload.ApiKeyNew != nil && *payload.ApiKeyNew != "" && len(*payload.ApiKeyNew) < 10 {
		service.logger.Debug(ErrApiKeyNewBad)
		return output.ErrorJsonErrValidationFailed(ErrApiKeyNewBad)
	}

	// CSR Extra Extensions - error checking is now in the custom unmarshal function

	// post processing command & env are optional but nothing to validate

	// post processing address
	if payload.PostProcessingClientAddress != nil && *payload.PostProcessingClientAddress != "" {
		valid := validation.DomainAndPortValid(*payload.PostProcessingClientAddress)
		if !valid {
			service.logger.Debug(ErrClientAddressBad)
			return output.ErrorJsonErrValidationFailed(ErrClientAddressBad)
		}
	}

	// post processing aes key (if specified)
	if payload.PostProcessingClientKeyB64 != nil {
		valid := clientKeyB64Valid(*payload.PostProcessingClientKeyB64)
		if !valid {
			service.logger.Debug(ErrPostProcessingClientKeyB64Bad)
			return output.ErrorJsonErrValidationFailed(ErrPostProcessingClientKeyB64Bad)
		}
	}

	// end validation

	// add additional details to the payload before saving
	payload.UpdatedAt = time.Now()

	// save account name and desc to storage, which also returns the account id with new
	// name and description
	updatedCert, err := service.storage.PutCertUpdate(payload)
	if err != nil {
		service.logger.Error(err)
		return output.ErrorJsonErrStorageGeneric(err)
	}

	// write response
	response := &certificateResponse{}
	response.StatusCode = http.StatusOK
	response.Message = "updated certificate"
	response.Certificate = updatedCert.detailedResponse()

	err = service.output.WriteJSON(w, response)
	if err != nil {
		service.logger.Errorf("failed to write json (%s)", err)
		return output.ErrorJsonErrWriteJsonError(err)
	}

	return nil
}
