package acme_servers

import (
	"certwarden-backend/pkg/acme"
	"certwarden-backend/pkg/output"
	"certwarden-backend/pkg/storage"
	"certwarden-backend/pkg/validation"
	"errors"
	"strings"
)

var (
	ErrIdBad   = errors.New("server id is invalid")
	ErrNameBad = errors.New("server name is not valid")
)

// getAcmeServer returns the Server for the specified id or an error.
func (service *Service) getServer(acmeServerId int) (Server, *output.Error) {
	// basic check
	if !validation.IsIdExistingValidRange(acmeServerId) {
		service.logger.Debug(ErrIdBad)
		return Server{}, output.ErrValidationFailed
	}

	// verify specified id has an acme service
	// this should never trigger, but just in case
	if service.acmeServers[acmeServerId] == nil {
		service.logger.Error("acme server id exists but doesn't have a service (how did this happen?!)")
		return Server{}, output.ErrInternal
	}

	// get the Server from storage
	server, err := service.storage.GetOneServerById(acmeServerId)
	if err != nil {
		// special error case for no record found
		if errors.Is(err, storage.ErrNoRecord) {
			service.logger.Debug(err)
			return Server{}, output.ErrNotFound
		} else {
			service.logger.Error(err)
			return Server{}, output.ErrStorageGeneric
		}
	}

	return server, nil
}

// AcmeServerValid returns true if the specified acme server id is valid
func (service *Service) AcmeServerValid(acmeServerId int) bool {
	// try to get server
	_, err := service.getServer(acmeServerId)

	// if err == nil, valid, otherwise invalid
	return err == nil
}

// nameValid returns true if the specified server name is acceptable and
// false if it is not. This check includes validating specified
// characters and also confirms the name is not already in use by another
// server. If an id is specified, the name will also be accepted if the name
// is already in use by the specified id.
func (service *Service) nameValid(serverName string, serverId *int) bool {
	// basic character/length check
	if !validation.NameValid(serverName) {
		return false
	}

	// make sure the name isn't already in use in storage
	key, err := service.storage.GetOneServerByName(serverName)
	if errors.Is(err, storage.ErrNoRecord) {
		// no rows means name is not in use
		return true
	} else if err != nil {
		// any other error
		return false
	}

	// if the returned server is the server being edited, name is ok
	if serverId != nil && key.ID == *serverId {
		return true
	}

	return false
}

// directoryUrlValid returns true if the specified acme directory url
// starts with https and actually returns a valid json ACME directory object
func (service *Service) directoryUrlValid(dirUrl string) bool {
	// require directory be specified as https://
	if !strings.HasPrefix(dirUrl, "https://") {
		return false
	}

	// check that dir actually fetches correctly
	_, err := acme.FetchAcmeDirectory(service.httpClient, dirUrl)
	if err != nil {
		service.logger.Debug(err)
		return false
	}

	return true
}
