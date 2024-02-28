package certificates

import (
	"encoding/json"
	"legocerthub-backend/pkg/output"
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
)

// DetailsUpdatePayload is the struct for editing an existing cert. A number of
// fields can be updated by the client on the fly (without ACME interaction).
type DetailsUpdatePayload struct {
	ID                        int      `json:"-"`
	Name                      *string  `json:"name"`
	Description               *string  `json:"description"`
	PrivateKeyId              *int     `json:"private_key_id"`
	SubjectAltNames           []string `json:"subject_alts"`
	Organization              *string  `json:"organization"`
	OrganizationalUnit        *string  `json:"organizational_unit"`
	Country                   *string  `json:"country"`
	State                     *string  `json:"state"`
	City                      *string  `json:"city"`
	PostProcessingCommand     *string  `json:"post_processing_command"`
	PostProcessingEnvironment []string `json:"post_processing_environment"`
	ApiKey                    *string  `json:"api_key"`
	ApiKeyNew                 *string  `json:"api_key_new"`
	ApiKeyViaUrl              *bool    `json:"api_key_via_url"`
	UpdatedAt                 int      `json:"-"`
}

// PutDetailsCert is a handler that sets various details about a cert and saves
// them to storage. These are all details that should be editable any time.
func (service *Service) PutDetailsCert(w http.ResponseWriter, r *http.Request) *output.Error {
	// payload decoding
	var payload DetailsUpdatePayload
	err := json.NewDecoder(r.Body).Decode(&payload)
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
	cert, outErr := service.GetCertificate(payload.ID)
	if outErr != nil {
		service.logger.Debug(ErrIdBad)
		return output.ErrValidationFailed
	}
	// name (optional)
	if payload.Name != nil && !service.nameValid(*payload.Name, &payload.ID) {
		service.logger.Debug(ErrNameBad)
		return output.ErrValidationFailed
	}
	// description - no validation
	// private key (optional)
	if payload.PrivateKeyId != nil && !service.privateKeyIdValid(*payload.PrivateKeyId, &payload.ID) {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}
	// subject alts (optional)
	// if new alts are being specified
	if payload.SubjectAltNames != nil {
		if !subjectAltsValid(payload.SubjectAltNames) {
			service.logger.Debug(ErrDomainBad)
			return output.ErrValidationFailed
		}

	} else if len(cert.SubjectAltNames) > 0 {
		// if keeping old alts and they exist (more than 0)
		// verify against the challenge method
		if !subjectAltsValid(cert.SubjectAltNames) {
			service.logger.Debug(ErrDomainBad)
			return output.ErrValidationFailed
		}
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
	// TODO: Do any validation of CSR components?

	// post processing command & env are optional but nothing to validate

	// end validation

	// add additional details to the payload before saving
	payload.UpdatedAt = int(time.Now().Unix())

	// save account name and desc to storage, which also returns the account id with new
	// name and description
	updatedCert, err := service.storage.PutDetailsCert(payload)
	if err != nil {
		service.logger.Error(err)
		return output.ErrStorageGeneric
	}

	// write response
	response := &certificateResponse{}
	response.StatusCode = http.StatusOK
	response.Message = "updated certificate"
	response.Certificate = updatedCert.detailedResponse()

	err = service.output.WriteJSON(w, response)
	if err != nil {
		service.logger.Errorf("failed to write json (%s)", err)
		return output.ErrWriteJsonError
	}

	return nil
}
