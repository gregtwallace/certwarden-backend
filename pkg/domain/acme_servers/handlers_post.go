package acme_servers

import (
	"certwarden-backend/pkg/acme"
	"certwarden-backend/pkg/output"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// NewPayload is used to post a new Server
type NewPayload struct {
	Name         *string `json:"name"`
	Description  *string `json:"description"`
	DirectoryURL *string `json:"directory_url"`
	IsStaging    *bool   `json:"is_staging"`
	CreatedAt    int     `json:"-"`
	UpdatedAt    int     `json:"-"`
}

// PostNewServer creates a new server, saves it to storage, and starts an *acme.Service
func (service *Service) PostNewServer(w http.ResponseWriter, r *http.Request) *output.JsonError {
	var payload NewPayload

	// decode body into payload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		service.logger.Debug(err)
		return output.JsonErrValidationFailed(err)
	}

	// do validation
	// name (missing or invalid)
	if payload.Name == nil || !service.nameValid(*payload.Name, nil) {
		service.logger.Debug(ErrNameBad)
		return output.JsonErrValidationFailed(ErrNameBad)
	}
	// description (if none, set to blank)
	if payload.Description == nil {
		payload.Description = new(string)
	}
	// directory url (required - confirm it actually fetches and decodes properly)
	if payload.DirectoryURL == nil {
		err = errors.New("cant post: directory url is missing")
		service.logger.Debug(err)
		return output.JsonErrValidationFailed(err)
	} else if !service.directoryUrlValid(*payload.DirectoryURL) {
		// if not nil, and validation fails, return error
		return output.JsonErrValidationFailed(errBadDirectoryURL)
	}
	// is staging (required)
	if payload.IsStaging == nil {
		err = errors.New("is_staging is not specified")
		service.logger.Debug(err)
		return output.JsonErrValidationFailed(err)
	}
	// end validation

	// add additional details to the payload before saving
	payload.CreatedAt = int(time.Now().Unix())
	payload.UpdatedAt = payload.CreatedAt

	// save new key to storage, which also returns the new server
	newServer, err := service.storage.PostNewServer(payload)
	if err != nil {
		service.logger.Error(err)
		return output.JsonErrStorageGeneric(err)
	}

	// spin up new acme.Service
	service.mu.Lock()
	defer service.mu.Unlock()

	service.acmeServers[newServer.ID], err = acme.NewService(service, *payload.DirectoryURL)
	if err != nil {
		service.logger.Error(err)
		return output.JsonErrInternal(err)
	}

	// make detailed response
	detailedResp, err := newServer.detailedResponse(service)
	if err != nil {
		err = fmt.Errorf("failed to generate server summary response (%s)", err)
		service.logger.Error(err)
		return output.JsonErrInternal(err)
	}

	// write response
	response := &acmeServerResponse{}
	response.StatusCode = http.StatusOK
	response.Message = "created server"
	response.Server = detailedResp

	// return response to client
	err = service.output.WriteJSON(w, response)
	if err != nil {
		service.logger.Errorf("failed to write json (%s)", err)
		return output.JsonErrWriteJsonError(err)
	}

	return nil
}
