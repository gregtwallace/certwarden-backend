package certificates

import (
	"encoding/json"
	"errors"
	"legocerthub-backend/pkg/domain/private_keys"
	"legocerthub-backend/pkg/domain/private_keys/key_crypto"
	"legocerthub-backend/pkg/output"
	"legocerthub-backend/pkg/randomness"
	"legocerthub-backend/pkg/validation"
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
)

// NewPayload is the struct for creating a new certificate
type NewPayload struct {
	Name                      *string  `json:"name"`
	Description               *string  `json:"description"`
	PrivateKeyID              *int     `json:"private_key_id"`
	NewKeyAlgorithmValue      *string  `json:"algorithm_value"`
	AcmeAccountID             *int     `json:"acme_account_id"`
	Subject                   *string  `json:"subject"`
	SubjectAltNames           []string `json:"subject_alts"`
	Organization              *string  `json:"organization"`
	OrganizationalUnit        *string  `json:"organizational_unit"`
	Country                   *string  `json:"country"`
	State                     *string  `json:"state"`
	City                      *string  `json:"city"`
	PostProcessingCommand     *string  `json:"post_processing_command"`
	PostProcessingEnvironment []string `json:"post_processing_environment"`
	// for post processing client, user submits enable or not, if enable key is generated and stored
	// bool is not stored anywhere (disabled == blank key value)
	PostProcessingClientEnable *bool  `json:"post_processing_client_enable"`
	PostProcessingClientKeyB64 string `json:"-"`
	ApiKey                     string `json:"-"`
	ApiKeyViaUrl               bool   `json:"-"`
	CreatedAt                  int    `json:"-"`
	UpdatedAt                  int    `json:"-"`
}

// PostNewCert creates a new certificate object in storage. No actual encryption certificate
// is generated, this only stores the needed information to later contact ACME and acquire
// the cert.
func (service *Service) PostNewCert(w http.ResponseWriter, r *http.Request) *output.Error {
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
	// description (if none, set to blank)
	if payload.Description == nil {
		payload.Description = new(string)
	}
	// private key
	// if key id not specified
	if payload.PrivateKeyID == nil {
		service.logger.Debug(ErrKeyIdBad)
		return output.ErrValidationFailed
	}
	// keep track if new key will be generated and saved
	generatedKeyPem := ""
	// if new key id specified
	if validation.IsIdNew(*payload.PrivateKeyID) {
		// confirm algorithm is specified
		if payload.NewKeyAlgorithmValue == nil || *payload.NewKeyAlgorithmValue == "" {
			service.logger.Debug(ErrKeyAlgorithmNone)
			return output.ErrValidationFailed
		}
		// confirm name is valid for a new key
		if payload.Name == nil || !service.keys.NameValid(*payload.Name, nil) {
			service.logger.Debug(ErrKeyNameBad)
			return output.ErrValidationFailed
		}
		// generate new key pem
		generatedKeyPem, err = key_crypto.AlgorithmByStorageValue(*payload.NewKeyAlgorithmValue).GeneratePrivateKeyPem()
		if err != nil {
			service.logger.Debug(err)
			return output.ErrValidationFailed
		}
	} else {
		// not new key id
		// error if algorithm value was specified
		if payload.NewKeyAlgorithmValue != nil && *payload.NewKeyAlgorithmValue != "" {
			service.logger.Debug(ErrKeyIdAndAlgorithm)
			return output.ErrValidationFailed
		}
		// error if key id is not valid
		if !service.privateKeyIdValid(*payload.PrivateKeyID, nil) {
			service.logger.Debug(err)
			return output.ErrValidationFailed
		}
	}
	// acme account
	if payload.AcmeAccountID == nil || !service.accounts.AccountUsable(*payload.AcmeAccountID) {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}
	// subject
	if payload.Subject == nil || !subjectValid(*payload.Subject) {
		service.logger.Debug(ErrDomainBad)
		return output.ErrValidationFailed
	}
	// subject alts
	// blank is okay, skip validation if not specified
	if payload.SubjectAltNames != nil && !subjectAltsValid(payload.SubjectAltNames) {
		service.logger.Debug(ErrDomainBad)
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
	// post processing command / env (don't check valid path, just let errors log if its bad)
	if payload.PostProcessingCommand == nil {
		payload.PostProcessingCommand = new(string)
	}
	if payload.PostProcessingEnvironment == nil {
		payload.PostProcessingEnvironment = []string{}
	}
	// post processing client enable
	if payload.PostProcessingClientEnable == nil {
		payload.PostProcessingClientEnable = new(bool)
	}
	// end validation

	// if new key was generated, save it to storage
	if generatedKeyPem != "" {
		// create new key payload
		newKeyPayload := private_keys.NewPayload{
			Name:           payload.Name,
			Description:    payload.Description,
			AlgorithmValue: payload.NewKeyAlgorithmValue,
			PemContent:     &generatedKeyPem,
			ApiKeyDisabled: new(bool),
			ApiKeyViaUrl:   payload.ApiKeyViaUrl,
		}
		// set additional new key payload fields
		newKeyPayload.ApiKey, err = randomness.GenerateApiKey()
		if err != nil {
			service.logger.Error(err)
			return output.ErrInternal
		}
		*newKeyPayload.ApiKeyDisabled = false
		newKeyPayload.CreatedAt = int(time.Now().Unix())
		newKeyPayload.UpdatedAt = payload.CreatedAt

		// save new key to storage, and set the cert key id based on returned key's id
		newKey, err := service.storage.PostNewKey(newKeyPayload)
		if err != nil {
			service.logger.Error(err)
			return output.ErrStorageGeneric
		}
		*payload.PrivateKeyID = newKey.ID
	}

	// add additional details to the payload before saving
	payload.ApiKey, err = randomness.GenerateApiKey()
	if err != nil {
		service.logger.Error(err)
		return output.ErrInternal
	}
	payload.ApiKeyViaUrl = false
	payload.CreatedAt = int(time.Now().Unix())
	payload.UpdatedAt = payload.CreatedAt
	// if client enabled, generate key to save (b64 raw url encoded)
	if payload.PostProcessingClientEnable != nil && *payload.PostProcessingClientEnable {
		payload.PostProcessingClientKeyB64, err = randomness.GenerateAES256KeyAsBase64RawUrl()
		if err != nil {
			service.logger.Errorf("failed to generate client key for certificate (%s)", err)
			return output.ErrInternal
		}
	}

	// save new cert
	newCert, err := service.storage.PostNewCert(payload)
	if err != nil {
		service.logger.Error(err)
		return output.ErrStorageGeneric
	}

	// write response
	response := &certificateResponse{}
	response.StatusCode = http.StatusCreated
	response.Message = "created certificate"
	response.Certificate = newCert.detailedResponse(service)

	err = service.output.WriteJSON(w, response)
	if err != nil {
		service.logger.Errorf("failed to write json (%s)", err)
		return output.ErrWriteJsonError
	}

	return nil
}

// StageNewApiKey generates a new API key and places it in the cert api_key_new
func (service *Service) StageNewApiKey(w http.ResponseWriter, r *http.Request) *output.Error {
	// get id param
	idParam := httprouter.ParamsFromContext(r.Context()).ByName("certid")
	certId, err := strconv.Atoi(idParam)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// validation
	// get cert (validate exists)
	cert, outErr := service.GetCertificate(certId)
	if outErr != nil {
		return outErr
	}

	// verify new api key is empty
	if cert.ApiKeyNew != "" {
		service.logger.Debug(errors.New("new api key already exists"))
		return output.ErrValidationFailed
	}
	// validation -- end

	// generate new api key
	newApiKey, err := randomness.GenerateApiKey()
	if err != nil {
		service.logger.Error(err)
		return output.ErrInternal
	}

	// update storage
	err = service.storage.PutCertNewApiKey(certId, newApiKey, int(time.Now().Unix()))
	if err != nil {
		service.logger.Error(err)
		return output.ErrStorageGeneric
	}
	cert.ApiKeyNew = newApiKey

	// write response
	response := &certificateResponse{}
	response.StatusCode = http.StatusCreated
	response.Message = "certificate new api key created"
	response.Certificate = cert.detailedResponse(service)

	err = service.output.WriteJSON(w, response)
	if err != nil {
		service.logger.Errorf("failed to write json (%s)", err)
		return output.ErrWriteJsonError
	}

	return nil
}

// MakeNewClientKey generates a new AES 256 encryption key and saves it to the specified
// certificate
func (service *Service) MakeNewClientKey(w http.ResponseWriter, r *http.Request) *output.Error {
	// get id param
	idParam := httprouter.ParamsFromContext(r.Context()).ByName("certid")
	certId, err := strconv.Atoi(idParam)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
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
		service.logger.Errorf("failed to generate client key (%s)", err)
		return output.ErrInternal
	}

	// update storage
	err = service.storage.PutCertClientKey(certId, clientKey, int(time.Now().Unix()))
	if err != nil {
		service.logger.Error(err)
		return output.ErrStorageGeneric
	}
	cert.PostProcessingClientKeyB64 = clientKey

	// write response
	response := &certificateResponse{}
	response.StatusCode = http.StatusCreated
	response.Message = "certificate new client key created"
	response.Certificate = cert.detailedResponse(service)

	err = service.output.WriteJSON(w, response)
	if err != nil {
		service.logger.Errorf("failed to write json (%s)", err)
		return output.ErrWriteJsonError
	}

	return nil
}
