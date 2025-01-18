package acme_accounts

import (
	"bytes"
	"certwarden-backend/pkg/acme"
	"certwarden-backend/pkg/output"
	"certwarden-backend/pkg/validation"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
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
func (service *Service) PostNewAccount(w http.ResponseWriter, r *http.Request) *output.JsonError {
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
		return output.JsonErrValidationFailed(ErrEmailBad)
	}

	// TOS must be accepted
	if payload.AcceptedTos == nil || !*payload.AcceptedTos {
		service.logger.Debug(err)
		return output.JsonErrValidationFailed(err)
	}

	// ACME Server
	if payload.AcmeServerID == nil || !service.acmeServerService.AcmeServerValid(*payload.AcmeServerID) {
		err = errors.New("acme_server_id not specified or invalid")
		service.logger.Debug(err)
		return output.JsonErrValidationFailed(err)
	}

	// private key (make last since most intense op)
	if payload.PrivateKeyID == nil || !service.keys.KeyAvailable(*payload.PrivateKeyID) {
		err = errors.New("private_key_id not available")
		service.logger.Debug(err)
		return output.JsonErrValidationFailed(err)
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
		return output.JsonErrStorageGeneric(err)
	}

	detailedResp, err := newAcct.detailedResponse(service)
	if err != nil {
		err = fmt.Errorf("failed to generate account summary response (%s)", err)
		service.logger.Error(err)
		return output.JsonErrInternal(err)
	}

	// write response
	response := &accountResponse{}
	response.StatusCode = http.StatusCreated
	response.Message = "created account"
	response.Account = detailedResp

	err = service.output.WriteJSON(w, response)
	if err != nil {
		service.logger.Errorf("failed to write json (%s)", err)
		return output.JsonErrWriteJsonError(err)
	}

	return nil
}

// PostAsGetPayload is used to send a PaG to the specified endpoint
type postAsGetPayload struct {
	URL string `json:"url"`
}

type postAsGetResponse struct {
	output.JsonResponse
	URL    string      `json:"url"`
	Body   string      `json:"body"`
	Header http.Header `json:"headers"`
}

// PostAsGet is a handler to allow a client to send a PaG request to a specified endpoint,
// signing the request using the specified account. This is not used in normal operation
// but is available for testing purposes.
func (service *Service) PostAsGet(w http.ResponseWriter, r *http.Request) *output.JsonError {
	// get id from param
	idParam := httprouter.ParamsFromContext(r.Context()).ByName("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		service.logger.Debug(err)
		return output.JsonErrValidationFailed(err)
	}

	// decode body into payload
	var payload postAsGetPayload
	err = json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		service.logger.Debug(err)
		return output.JsonErrValidationFailed(err)
	}

	// validation
	// id
	account, outErr := service.getAccount(id)
	if outErr != nil {
		return outErr
	}

	// URL
	u, err := url.Parse(payload.URL)
	if err != nil || u.Scheme != "https" || u.Host == "" {
		return output.JsonErrValidationFailed(fmt.Errorf("url (%s) is insecure or not valid", payload.URL))
	}

	// end validation

	// get AccountKey
	acmeAccountKey, err := account.AcmeAccountKey()
	if err != nil {
		service.logger.Error(err)
		return output.JsonErrInternal(err)
	}

	// send the PaG to ACME
	acmeService, err := service.acmeServerService.AcmeService(account.AcmeServer.ID)
	if err != nil {
		service.logger.Error(err)
		return output.JsonErrInternal(err)
	}

	resp, header, err := acmeService.PostAsGet(payload.URL, acmeAccountKey)
	if err != nil {
		// if ACME error, don't actually error this API call, just return the error
		// information to the client
		_, isAcmeErr := err.(*acme.Error)
		if !isAcmeErr {
			service.logger.Error(err)
			return output.JsonErrInternal(err)
		}
	}

	// if response body is json, ensure it is pretty using Indent
	if strings.EqualFold(header.Get("Content-Type"), "application/json") {
		var prettyJson bytes.Buffer
		err = json.Indent(&prettyJson, resp, "", "\t")
		if err != nil {
			service.logger.Errorf("accounts: debug pag returned json content type but failed to indent for output (%s)", err)
		} else {
			resp = prettyJson.Bytes()
		}
	}

	// write response
	response := &postAsGetResponse{}
	response.StatusCode = http.StatusOK
	response.Message = "ok"
	response.URL = payload.URL
	response.Body = string(resp)
	response.Header = header

	err = service.output.WriteJSON(w, response)
	if err != nil {
		service.logger.Errorf("failed to write json (%s)", err)
		return output.JsonErrWriteJsonError(err)
	}

	return nil
}
