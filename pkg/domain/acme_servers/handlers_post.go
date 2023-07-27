package acme_servers

import (
	"encoding/json"
	"legocerthub-backend/pkg/acme"
	"legocerthub-backend/pkg/output"
	"net/http"
	"time"
)

// NewPayload is used to post a new Server to LeGo
type NewPayload struct {
	Name         *string `json:"name"`
	Description  *string `json:"description"`
	DirectoryURL *string `json:"directory_url"`
	IsStaging    *bool   `json:"is_staging"`
	CreatedAt    int     `json:"-"`
	UpdatedAt    int     `json:"-"`
}

// PostNewServer creates a new server, saves it to storage, and starts an *acme.Service
func (service *Service) PostNewServer(w http.ResponseWriter, r *http.Request) (err error) {
	var payload NewPayload

	// decode body into payload
	err = json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// do validation
	// name (missing or invalid)
	if payload.Name == nil || !service.nameValid(*payload.Name, nil) {
		service.logger.Debug(ErrNameBad)
		return output.ErrValidationFailed
	}
	// description (if none, set to blank)
	if payload.Description == nil {
		payload.Description = new(string)
	}
	// directory url (required - confirm it actually fetches and decodes properly)
	if payload.DirectoryURL == nil {
		service.logger.Debug("cant post: directory url is missing")
		return output.ErrValidationFailed
	} else if !service.directoryUrlValid(*payload.DirectoryURL) {
		// if not nil, and validation fails, return error
		return output.ErrBadDirectoryURL
	}
	// is staging (required)
	if payload.IsStaging == nil {
		service.logger.Debug("cant post: is_staging is missing")
		return output.ErrValidationFailed
	}
	// end validation

	// add additional details to the payload before saving
	payload.CreatedAt = int(time.Now().Unix())
	payload.UpdatedAt = payload.CreatedAt

	// save new key to storage, which also returns the new key id
	serverId, err := service.storage.PostNewServer(payload)
	if err != nil {
		service.logger.Error(err)
		return output.ErrStorageGeneric
	}

	// spin up new acme.Service
	service.mu.Lock()
	defer service.mu.Unlock()

	service.acmeServers[serverId], err = acme.NewService(service, *payload.DirectoryURL)
	if err != nil {
		service.logger.Error(err)
		return output.ErrInternal
	}

	// return response to client
	response := output.JsonResponse{
		Status:  http.StatusCreated,
		Message: "created",
		ID:      serverId,
	}

	_, err = service.output.WriteJSON(w, response.Status, response, "response")
	if err != nil {
		return err
	}

	return nil
}
