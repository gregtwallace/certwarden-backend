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
	Name                        *string             `json:"name"`
	Description                 *string             `json:"description"`
	PrivateKeyID                *int                `json:"private_key_id"`
	NewKeyAlgorithmValue        *string             `json:"algorithm_value"`
	AcmeAccountID               *int                `json:"acme_account_id"`
	Subject                     *string             `json:"subject"`
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
	PostProcessingClientKeyB64  string              `json:"-"`
	Profile                     *string             `json:"profile"`
	TechPhone                   *string             `json:"tech_phone"`
	TechEmail                   *string             `json:"tech_email"`
	ApiKey                      string              `json:"-"`
	ApiKeyViaUrl                bool                `json:"-"`
	CreatedAt                   int                 `json:"-"`
	UpdatedAt                   int                 `json:"-"`
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
		return output.JsonErrValidationFailed(err)
	}

	// validation
	// name
	if payload.Name == nil || !service.nameValid(*payload.Name, nil) {
		service.logger.Debug(ErrNameBad)
		return output.JsonErrValidationFailed(ErrNameBad)
	}
	// description (if none, set to blank)
	if payload.Description == nil {
		payload.Description = new(string)
	}
	// private key
	// if key id not specified
	if payload.PrivateKeyID == nil {
		service.logger.Debug(ErrKeyIdBad)
		return output.JsonErrValidationFailed(ErrKeyIdBad)
	}
	// keep track if new key will be generated and saved
	generatedKeyPem := ""
	// if new key id specified
	if validation.IsIdNew(*payload.PrivateKeyID) {
		// confirm algorithm is specified
		if payload.NewKeyAlgorithmValue == nil || *payload.NewKeyAlgorithmValue == "" {
			service.logger.Debug(ErrKeyAlgorithmNone)
			return output.JsonErrValidationFailed(ErrKeyAlgorithmNone)
		}
		// confirm name is valid for a new key
		if payload.Name == nil || !service.keys.NameValid(*payload.Name, nil) {
			service.logger.Debug(ErrKeyNameBad)
			return output.JsonErrValidationFailed(ErrKeyNameBad)
		}
		// generate new key pem
		generatedKeyPem, err = key_crypto.AlgorithmByStorageValue(*payload.NewKeyAlgorithmValue).GeneratePrivateKeyPem()
		if err != nil {
			service.logger.Debug(err)
			return output.JsonErrValidationFailed(err)
		}
	} else {
		// not new key id
		// error if algorithm value was specified
		if payload.NewKeyAlgorithmValue != nil && *payload.NewKeyAlgorithmValue != "" {
			service.logger.Debug(ErrKeyIdAndAlgorithm)
			return output.JsonErrValidationFailed(ErrKeyIdAndAlgorithm)
		}
		// error if key id is not valid
		if !service.privateKeyIdValid(*payload.PrivateKeyID, nil) {
			service.logger.Debug(ErrKeyIdBad)
			return output.JsonErrValidationFailed(ErrKeyIdBad)
		}
	}
	// acme account
	if payload.AcmeAccountID == nil {
		err = errors.New("acme account id is unspecified")
		service.logger.Debug(err)
		return output.JsonErrValidationFailed(err)
	}
	acctUsable, acct := service.accounts.AccountUsable(*payload.AcmeAccountID)
	if !acctUsable {
		err = errors.New("acme account id does not exist or is not usable")
		service.logger.Debug(err)
		return output.JsonErrValidationFailed(err)
	}
	// subject
	if payload.Subject == nil || !subjectValid(*payload.Subject) {
		service.logger.Debug(ErrDomainBad)
		return output.JsonErrValidationFailed(ErrDomainBad)
	}
	// subject alts
	// blank is okay, skip validation if not specified
	if payload.SubjectAltNames != nil && !subjectAltsValid(payload.SubjectAltNames) {
		service.logger.Debug(ErrDomainBad)
		return output.JsonErrValidationFailed(ErrDomainBad)
	}
	// profile Extension -- validate if specified, else blank
	if payload.Profile == nil {
		payload.Profile = new(string)
	} else if *payload.Profile != "" {
		// specified, validate against acme service
		acmeService, err := service.acmeServerService.AcmeService(acct.AcmeServer.ID)
		if err != nil {
			err = fmt.Errorf("failed to retrieve acme service (%s)", err)
			service.logger.Error(err)
			return output.JsonErrInternal(err)
		}
		if !acmeService.ProfileValidate(*payload.Profile) {
			err = fmt.Errorf("acme service for specified account does not advertise profile `%s`", *payload.Profile)
			service.logger.Debug(err)
			return output.JsonErrValidationFailed(err)
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

	// CSR Extra Extensions - check each extra extension for proper formatting
	for i := range payload.CSRExtraExtensions {
		_, err = payload.CSRExtraExtensions[i].ToCertExtension()
		if err != nil {
			service.logger.Debug(err)
			return output.JsonErrValidationFailed(err)
		}
	}

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
			return output.JsonErrValidationFailed(ErrClientAddressBad)
		}
	}
	// end validation

	// tech phone number
	if payload.TechPhone == nil {
		payload.TechPhone = new(string)
	} else if *payload.TechPhone != "" {
		valid := validation.PhoneValid(*payload.TechPhone)
		if !valid {
			service.logger.Debug(ErrPhoneBad)
			return output.JsonErrValidationFailed(ErrPhoneBad)
		}
	}
	// end validation

	// tech email address
	if payload.TechEmail == nil {
		payload.TechEmail = new(string)
	} else if *payload.TechEmail != "" {
		valid := validation.EmailValid(*payload.TechEmail)
		if !valid {
			service.logger.Debug(ErrEmailBad)
			return output.JsonErrValidationFailed(ErrEmailBad)
		}
	}
	// end validation

	// if new private key was generated, save it to storage
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
			return output.JsonErrInternal(err)
		}
		*newKeyPayload.ApiKeyDisabled = false
		newKeyPayload.CreatedAt = int(time.Now().Unix())
		newKeyPayload.UpdatedAt = payload.CreatedAt

		// save new key to storage, and set the cert key id based on returned key's id
		newKey, err := service.storage.PostNewKey(newKeyPayload)
		if err != nil {
			service.logger.Error(err)
			return output.JsonErrStorageGeneric(err)
		}
		*payload.PrivateKeyID = newKey.ID
	}

	// add additional details to the payload before saving
	payload.ApiKey, err = randomness.GenerateApiKey()
	if err != nil {
		service.logger.Error(err)
		return output.JsonErrInternal(err)
	}
	payload.ApiKeyViaUrl = false
	payload.CreatedAt = int(time.Now().Unix())
	payload.UpdatedAt = payload.CreatedAt
	// if client address specified, generate key to save (b64 raw url encoded)
	if payload.PostProcessingClientAddress != nil && *payload.PostProcessingClientAddress != "" {
		payload.PostProcessingClientKeyB64, err = randomness.GenerateAES256KeyAsBase64RawUrl()
		if err != nil {
			err = fmt.Errorf("failed to generate client key for certificate (%s)", err)
			service.logger.Error(err)
			return output.JsonErrInternal(err)
		}
	}

	// save new cert
	newCert, err := service.storage.PostNewCert(payload)
	if err != nil {
		service.logger.Error(err)
		return output.JsonErrStorageGeneric(err)
	}

	// write response
	response := &certificateResponse{}
	response.StatusCode = http.StatusCreated
	response.Message = "created certificate"
	response.Certificate = newCert.detailedResponse()

	err = service.output.WriteJSON(w, response)
	if err != nil {
		service.logger.Errorf("failed to write json (%s)", err)
		return output.JsonErrWriteJsonError(err)
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
		return output.JsonErrValidationFailed(err)
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
		return output.JsonErrValidationFailed(err)
	}
	// validation -- end

	// generate new api key
	newApiKey, err := randomness.GenerateApiKey()
	if err != nil {
		service.logger.Error(err)
		return output.JsonErrInternal(err)
	}

	// update storage
	err = service.storage.PutCertNewApiKey(certId, newApiKey, int(time.Now().Unix()))
	if err != nil {
		service.logger.Error(err)
		return output.JsonErrStorageGeneric(err)
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
		return output.JsonErrWriteJsonError(err)
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
		return output.JsonErrValidationFailed(err)
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
		err = fmt.Errorf("failed to generate client key (%s)", err)
		service.logger.Error(err)
		return output.JsonErrInternal(err)
	}

	// update storage
	err = service.storage.PutCertClientKey(certId, clientKey, int(time.Now().Unix()))
	if err != nil {
		service.logger.Error(err)
		return output.JsonErrStorageGeneric(err)
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
		return output.JsonErrWriteJsonError(err)
	}

	return nil
}
