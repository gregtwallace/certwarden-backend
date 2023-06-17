package acme_servers

import (
	"errors"
	"legocerthub-backend/pkg/output"
	"legocerthub-backend/pkg/storage"
	"legocerthub-backend/pkg/validation"
)

var (
	ErrIdBad         = errors.New("key id is invalid")
	ErrNoAcmeService = errors.New("acme server id exists but doesn't have a service (how did this happen?!)")
)

// getAcmeServer returns the Server for the specified id or an error.
func (service *Service) getServer(acmeServerId int) (Server, error) {
	// basic check
	if !validation.IsIdExistingValidRange(acmeServerId) {
		service.logger.Debug(ErrIdBad)
		return Server{}, output.ErrValidationFailed
	}

	// verify specified id has an acme service
	// this should never trigger, but just in case
	if service.acmeServers[acmeServerId] == nil {
		return Server{}, ErrNoAcmeService
	}

	// get the Server from storage
	server, err := service.storage.GetOneServerById(acmeServerId)
	if err != nil {
		// special error case for no record found
		if err == storage.ErrNoRecord {
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
