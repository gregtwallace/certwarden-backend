package certificates

import (
	"certwarden-backend/pkg/output"
	"certwarden-backend/pkg/validation"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
)

// DetailsUpdatePayload is the struct for editing an existing cert. A number of
// fields can be updated by the client on the fly (without ACME interaction).
type DetailsUpdatePayload struct {
	ID                          int                 `json:"-"`
	Name                        *string             `json:"name"`
	Description                 *string             `json:"description"`
	PrivateKeyId                *int                `json:"private_key_id"`
	SubjectAltNames             []string            `json:"subject_alts"`
	Organization                *string             `json:"organization"`
	OrganizationalUnit          *string             `json:"organizational_unit"`
	Country                     *string             `json:"country"`
	State                       *string             `json:"state"`
	City                        *string             `json:"city"`
	CSRExtraExtensions          []CertExtensionJSON `json:"csr_extra_extensions"`
	PreferredRootCN             *string             `json:"preferred_root_cn"`
	PostProcessingCommand       *string             `json:"post_processing_command"`
	PostProcessingEnvironment   []string            `json:"post_processing_environment"`
	PostProcessingClientAddress *string             `json:"post_processing_client_address"`
	ApiKey                      *string             `json:"api_key"`
	ApiKeyNew                   *string             `json:"api_key_new"`
	ApiKeyViaUrl                *bool               `json:"api_key_via_url"`
	UpdatedAt                   int                 `json:"-"`
}

// PutDetailsCert is a handler that sets various details about a cert and saves
// them to storage. These are all details that should be editable any time.
func (service *Service) PutDetailsCert(w http.ResponseWriter, r *http.Request) *output.JsonError {
	// payload decoding
	var payload DetailsUpdatePayload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		service.logger.Debug(err)
		return output.JsonErrValidationFailed(err)
	}

	// get id from param
	idParam := httprouter.ParamsFromContext(r.Context()).ByName("certid")
	payload.ID, err = strconv.Atoi(idParam)
	if err != nil {
		service.logger.Debug(err)
		return output.JsonErrValidationFailed(err)
	}

	// validation
	// id
	cert, outErr := service.GetCertificate(payload.ID)
	if outErr != nil {
		service.logger.Debug(ErrIdBad)
		return output.JsonErrValidationFailed(ErrIdBad)
	}
	// name (optional)
	if payload.Name != nil && !service.nameValid(*payload.Name, &payload.ID) {
		service.logger.Debug(ErrNameBad)
		return output.JsonErrValidationFailed(ErrNameBad)
	}
	// description - no validation
	// private key (optional)
	if payload.PrivateKeyId != nil && !service.privateKeyIdValid(*payload.PrivateKeyId, &payload.ID) {
		return output.JsonErrValidationFailed(ErrKeyIdBad)
	}
	// subject alts (optional)
	// if new alts are being specified
	if payload.SubjectAltNames != nil {
		if !subjectAltsValid(payload.SubjectAltNames) {
			service.logger.Debug(ErrDomainBad)
			return output.JsonErrValidationFailed(ErrDomainBad)
		}

	} else if len(cert.SubjectAltNames) > 0 {
		// if keeping old alts and they exist (more than 0)
		// verify against the challenge method
		if !subjectAltsValid(cert.SubjectAltNames) {
			service.logger.Debug(ErrDomainBad)
			return output.JsonErrValidationFailed(ErrDomainBad)
		}
	}
	// api key must be at least 10 characters long
	if payload.ApiKey != nil && len(*payload.ApiKey) < 10 {
		service.logger.Debug(ErrApiKeyBad)
		return output.JsonErrValidationFailed(ErrApiKeyBad)
	}
	// api key new must be at least 10 characters long
	if payload.ApiKeyNew != nil && *payload.ApiKeyNew != "" && len(*payload.ApiKeyNew) < 10 {
		service.logger.Debug(ErrApiKeyNewBad)
		return output.JsonErrValidationFailed(ErrApiKeyNewBad)
	}
	// TODO: Do any validation of CSR components?

	// CSR Extra Extensions - check each extra extension for proper formatting
	for i := range payload.CSRExtraExtensions {
		_, err = payload.CSRExtraExtensions[i].ToCertExtension()
		if err != nil {
			service.logger.Debug(err)
			return output.JsonErrValidationFailed(err)
		}
	}

	// post processing command & env are optional but nothing to validate

	// post processing address
	if payload.PostProcessingClientAddress == nil {
		payload.PostProcessingClientAddress = new(string)
	} else if *payload.PostProcessingClientAddress != "" {
		valid := validation.DomainValid(*payload.PostProcessingClientAddress, false)
		if !valid {
			service.logger.Debug(ErrClientAddressBad)
			return output.JsonErrValidationFailed(ErrClientAddressBad)
		}
	}

	// end validation

	// add additional details to the payload before saving
	payload.UpdatedAt = int(time.Now().Unix())

	// save account name and desc to storage, which also returns the account id with new
	// name and description
	updatedCert, err := service.storage.PutDetailsCert(payload)
	if err != nil {
		service.logger.Error(err)
		return output.JsonErrStorageGeneric(err)
	}

	// write response
	response := &certificateResponse{}
	response.StatusCode = http.StatusOK
	response.Message = "updated certificate"
	response.Certificate = updatedCert.detailedResponse()

	err = service.output.WriteJSON(w, response)
	if err != nil {
		service.logger.Errorf("failed to write json (%s)", err)
		return output.JsonErrWriteJsonError(err)
	}

	return nil
}
