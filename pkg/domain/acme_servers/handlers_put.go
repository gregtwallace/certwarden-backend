package acme_servers

import (
	"encoding/json"
	"legocerthub-backend/pkg/acme"
	"legocerthub-backend/pkg/output"
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
)

// UpdatePayload is the struct for editing an existing Server's
// information (only certain fields are editable)
type UpdatePayload struct {
	ID           int     `json:"-"`
	Name         *string `json:"name"`
	Description  *string `json:"description"`
	DirectoryURL *string `json:"directory_url"`
	IsStaging    *bool   `json:"is_staging"`
	UpdatedAt    int     `json:"-"`
}

// PutServerUpdate updates a Server that already exists in storage.
// Only fields received in the payload (non-nil) are updated.
func (service *Service) PutServerUpdate(w http.ResponseWriter, r *http.Request) *output.Error {
	// parse payload
	var payload UpdatePayload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// get id param
	idParam := httprouter.ParamsFromContext(r.Context()).ByName("id")
	payload.ID, err = strconv.Atoi(idParam)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// validation
	// id
	_, outErr := service.getServer(payload.ID)
	if outErr != nil {
		return outErr
	}
	// name (optional - check if not nil)
	if payload.Name != nil && !service.nameValid(*payload.Name, &payload.ID) {
		service.logger.Debug(ErrNameBad)
		return output.ErrValidationFailed
	}
	// directory_url (optional - check if not nil)
	if payload.DirectoryURL != nil && !service.directoryUrlValid(*payload.DirectoryURL) {
		return output.ErrBadDirectoryURL
	}
	// Description, and IsStaging do not need validation
	// end validation

	// add additional details to the payload before saving
	payload.UpdatedAt = int(time.Now().Unix())

	// save updated key info to storage
	updatedServer, err := service.storage.PutServerUpdate(payload)
	if err != nil {
		service.logger.Error(err)
		return output.ErrStorageGeneric
	}

	// if directory url changed, create new acme.Service
	if payload.DirectoryURL != nil {
		service.mu.Lock()
		defer service.mu.Unlock()

		service.acmeServers[payload.ID], err = acme.NewService(service, *payload.DirectoryURL)
		if err != nil {
			service.logger.Error(err)
			return output.ErrInternal
		}
	}

	// make detailed response
	detailedResp, err := updatedServer.detailedResponse(service)
	if err != nil {
		service.logger.Errorf("failed to generate server summary response (%s)", err)
		return output.ErrInternal
	}

	// write response
	response := &acmeServerResponse{}
	response.StatusCode = http.StatusOK
	response.Message = "updated server"
	response.Server = detailedResp

	// return response to client
	err = service.output.WriteJSON(w, response)
	if err != nil {
		service.logger.Errorf("failed to write json (%s)", err)
		return output.ErrWriteJsonError
	}

	return nil
}
