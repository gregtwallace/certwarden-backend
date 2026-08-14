package certificates

import (
	"certwarden-backend/pkg/domain/private_keys"
	"certwarden-backend/pkg/domain/private_keys/key_crypto"
	"certwarden-backend/pkg/output"
	"certwarden-backend/pkg/randomness"
	"certwarden-backend/pkg/validation"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
)

// NewPayload is the struct for creating a new certificate
type NewPayload struct {
	Name                        *string         `json:"name"`
	Description                 *string         `json:"description"`
	PrivateKeyID                *int            `json:"private_key_id"`
	NewKeyAlgorithmValue        *string         `json:"algorithm_value"`
	AcmeAccountID               *int            `json:"acme_account_id"`
	Subject                     *string         `json:"subject"`
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
	ApiKey                      string          `json:"-"`
	ApiKeyViaUrl                bool            `json:"-"`
	CreatedAt                   time.Time       `json:"-"`
	UpdatedAt                   time.Time       `json:"-"`
}

// PostNewCert creates a new certificate object in storage. No actual encryption certificate
// is generated, this only stores the needed information to later contact ACME and acquire
// the cert.
func (service *Service) PostNewCert(w http.ResponseWriter, r *http.Request) *output.JsonError {
	var payload NewPayload

	// decode body into payload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrorJsonErrValidationFailed(err)
	}

	// do validation
	// name
	if payload.Name == nil || !service.nameValid(*payload.Name, nil) {
		service.logger.Debug(ErrNameBad)
		return output.ErrorJsonErrValidationFailed(ErrNameBad)
	}
	// description (if none, set to blank)
	if payload.Description == nil {
		payload.Description = new(string)
	}
	// private key
	// if key id not specified
	if payload.PrivateKeyID == nil {
		service.logger.Debug(ErrKeyIdBad)
		return output.ErrorJsonErrValidationFailed(ErrKeyIdBad)
	}
	// keep track if new key will be generated and saved
	generatedKeyPem := ""
	// if new key id specified
	if validation.IsIdNew(*payload.PrivateKeyID) {
		// confirm algorithm is specified
		if payload.NewKeyAlgorithmValue == nil || *payload.NewKeyAlgorithmValue == "" {
			service.logger.Debug(ErrKeyAlgorithmNone)
			return output.ErrorJsonErrValidationFailed(ErrKeyAlgorithmNone)
		}
		// confirm name is valid for a new key
		if payload.Name == nil || !service.keys.NameValid(*payload.Name, nil) {
			service.logger.Debug(ErrKeyNameBad)
			return output.ErrorJsonErrValidationFailed(ErrKeyNameBad)
		}
		// generate new key pem
		generatedKeyPem, err = key_crypto.AlgorithmByStorageValue(*payload.NewKeyAlgorithmValue).GeneratePrivateKeyPem()
		if err != nil {
			service.logger.Debug(err)
			return output.ErrorJsonErrValidationFailed(err)
		}
	} else {
		// not new key id
		// error if algorithm value was specified
		if payload.NewKeyAlgorithmValue != nil && *payload.NewKeyAlgorithmValue != "" {
			service.logger.Debug(ErrKeyIdAndAlgorithm)
			return output.ErrorJsonErrValidationFailed(ErrKeyIdAndAlgorithm)
		}
		// error if key id is not valid
		if !service.privateKeyIdValid(*payload.PrivateKeyID, nil) {
			service.logger.Debug(ErrKeyIdBad)
			return output.ErrorJsonErrValidationFailed(ErrKeyIdBad)
		}
	}
	// acme account
	if payload.AcmeAccountID == nil {
		err = errors.New("acme account id is unspecified")
		service.logger.Debug(err)
		return output.ErrorJsonErrValidationFailed(err)
	}
	acctUsable, acct := service.accounts.AccountUsable(*payload.AcmeAccountID)
	if !acctUsable {
		err = errors.New("acme account id does not exist or is not usable")
		service.logger.Debug(err)
		return output.ErrorJsonErrValidationFailed(err)
	}
	// subject
	if payload.Subject == nil || !subjectValid(*payload.Subject) {
		service.logger.Debug(ErrDomainBad)
		return output.ErrorJsonErrValidationFailed(ErrDomainBad)
	}
	// subject alts
	// blank is okay, skip validation if not specified
	if payload.SubjectAltNames != nil && !subjectAltsValid(payload.SubjectAltNames) {
		service.logger.Debug(ErrDomainBad)
		return output.ErrorJsonErrValidationFailed(ErrDomainBad)
	}
	// profile Extension -- validate if specified, else blank
	if payload.Profile == nil {
		payload.Profile = new(string)
	} else if *payload.Profile != "" {
		// specified, validate against acme service
		acmeService, err := service.acmeServerService.AcmeService(acct.AcmeServer.ID)
		if err != nil {
			err = fmt.Errorf("failed to retrieve acme service (%w)", err)
			service.logger.Error(err)
			return output.ErrorJsonErrInternal(err)
		}
		if !acmeService.ProfileValidate(*payload.Profile) {
			err = fmt.Errorf("acme service for specified account does not advertise profile `%s`", *payload.Profile)
			service.logger.Debug(err)
			return output.ErrorJsonErrValidationFailed(err)
		}
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

	// CSR Extra Extensions - error checking is now in the custom unmarshal function

	if payload.PreferredRootCN == nil {
		payload.PreferredRootCN = new(string)
	}

	// post processing command / env (don't check valid path, just let errors log if its bad)
	if payload.PostProcessingCommand == nil {
		payload.PostProcessingCommand = new(string)
	}
	if payload.PostProcessingEnvironment == nil {
		payload.PostProcessingEnvironment = []string{}
	}
	// post processing address
	if payload.PostProcessingClientAddress == nil {
		payload.PostProcessingClientAddress = new(string)
	} else if *payload.PostProcessingClientAddress != "" {
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

	// if new private key was generated, save it to storage
	createdAtAndUpdatedAt := time.Now()
	if generatedKeyPem != "" {
		apiKey, err := randomness.GenerateApiKey()
		if err != nil {
			service.logger.Error(err)
			return output.ErrorJsonErrInternal(err)
		}

		// create new key payload
		newKeyPayload := private_keys.NewPayload{
			Name:           payload.Name,
			Description:    payload.Description,
			AlgorithmValue: payload.NewKeyAlgorithmValue,
			PemContent:     &generatedKeyPem,
			ApiKeyDisabled: new(false),
			ApiKeyViaUrl:   payload.ApiKeyViaUrl,
			ApiKey:         apiKey,
			CreatedAt:      createdAtAndUpdatedAt,
			UpdatedAt:      createdAtAndUpdatedAt,
		}

		// save new key to storage, and set the cert key id based on returned key's id
		newKey, err := service.storage.PostNewKey(newKeyPayload)
		if err != nil {
			service.logger.Error(err)
			return output.ErrorJsonErrStorageGeneric(err)
		}
		*payload.PrivateKeyID = newKey.ID
	}

	// add additional details to the payload before saving
	payload.ApiKey, err = randomness.GenerateApiKey()
	if err != nil {
		service.logger.Error(err)
		return output.ErrorJsonErrInternal(err)
	}
	payload.ApiKeyViaUrl = false
	payload.CreatedAt = createdAtAndUpdatedAt
	payload.UpdatedAt = createdAtAndUpdatedAt

	// if client address specified but no aes key, generate key to save (b64 raw url encoded)
	if payload.PostProcessingClientKeyB64 == nil {
		// empty if no processing address
		payload.PostProcessingClientKeyB64 = new("")

		// processing address & user didnt specify an aes key -- generate one
		if payload.PostProcessingClientAddress != nil && *payload.PostProcessingClientAddress != "" {
			key, err := randomness.GenerateAES256KeyAsBase64RawUrl()
			if err != nil {
				err = fmt.Errorf("failed to generate client key for certificate (%w)", err)
				service.logger.Error(err)
				return output.ErrorJsonErrInternal(err)
			}
			payload.PostProcessingClientKeyB64 = &key
		}
	}

	// save new cert
	newCert, err := service.storage.PostNewCert(payload)
	if err != nil {
		service.logger.Error(err)
		return output.ErrorJsonErrStorageGeneric(err)
	}

	// write response
	response := &certificateResponse{}
	response.StatusCode = http.StatusCreated
	response.Message = "created certificate"
	response.Certificate = newCert.detailedResponse()

	err = service.output.WriteJSON(w, response)
	if err != nil {
		service.logger.Errorf("failed to write json (%s)", err)
		return output.ErrorJsonErrWriteJsonError(err)
	}

	return nil
}

// StageNewApiKey generates a new API key and places it in the cert api_key_new
func (service *Service) StageNewApiKey(w http.ResponseWriter, r *http.Request) *output.JsonError {
	// get id param
	idParam := httprouter.ParamsFromContext(r.Context()).ByName("certid")
	certId, err := strconv.Atoi(idParam)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrorJsonErrValidationFailed(err)
	}

	// validation
	// get cert (validate exists)
	cert, outErr := service.GetCertificate(certId)
	if outErr != nil {
		return outErr
	}

	// verify new api key is empty
	if cert.ApiKeyNew != "" {
		err = errors.New("new api key already exists")
		service.logger.Debug(err)
		return output.ErrorJsonErrValidationFailed(err)
	}
	// validation -- end

	// generate new api key
	newApiKey, err := randomness.GenerateApiKey()
	if err != nil {
		service.logger.Error(err)
		return output.ErrorJsonErrInternal(err)
	}

	// update storage
	err = service.storage.PutCertApiKeyNew(certId, newApiKey, time.Now())
	if err != nil {
		service.logger.Error(err)
		return output.ErrorJsonErrStorageGeneric(err)
	}
	cert.ApiKeyNew = newApiKey

	// write response
	response := &certificateResponse{}
	response.StatusCode = http.StatusCreated
	response.Message = "certificate new api key created"
	response.Certificate = cert.detailedResponse()

	err = service.output.WriteJSON(w, response)
	if err != nil {
		service.logger.Errorf("failed to write json (%s)", err)
		return output.ErrorJsonErrWriteJsonError(err)
	}

	return nil
}

// MakeNewClientKey generates a new AES 256 encryption key and saves it to the specified
// certificate
func (service *Service) MakeNewClientKey(w http.ResponseWriter, r *http.Request) *output.JsonError {
	// get id param
	idParam := httprouter.ParamsFromContext(r.Context()).ByName("certid")
	certId, err := strconv.Atoi(idParam)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrorJsonErrValidationFailed(err)
	}

	// validation
	// get cert (validate exists)
	cert, outErr := service.GetCertificate(certId)
	if outErr != nil {
		return outErr
	}
	// validation -- end

	// generate AES 256 key
	clientKey, err := randomness.GenerateAES256KeyAsBase64RawUrl()
	if err != nil {
		err = fmt.Errorf("failed to generate client key (%w)", err)
		service.logger.Error(err)
		return output.ErrorJsonErrInternal(err)
	}

	// update storage
	err = service.storage.PutCertClientKey(certId, clientKey, time.Now())
	if err != nil {
		service.logger.Error(err)
		return output.ErrorJsonErrStorageGeneric(err)
	}
	cert.PostProcessingClientKeyB64 = clientKey

	// write response
	response := &certificateResponse{}
	response.StatusCode = http.StatusCreated
	response.Message = "certificate new client key created"
	response.Certificate = cert.detailedResponse()

	err = service.output.WriteJSON(w, response)
	if err != nil {
		service.logger.Errorf("failed to write json (%s)", err)
		return output.ErrorJsonErrWriteJsonError(err)
	}

	return nil
}
