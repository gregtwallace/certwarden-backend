package dnspersist01manual

import (
	"certwarden-backend/pkg/acme"
)

// var errServiceComponent = errors.New("necessary dns-persist-01 manual component is missing")

// needed components from app - none
type App interface {
}

// provider Service struct - doesn't need to do anything
type Service struct {
}

// ChallengeType returns the ACME Challenge Type this provider uses, which is dns-01-persist
func (service *Service) AcmeChallengeType() acme.ChallengeType {
	return acme.ChallengeTypeDnsPersist01
}

// Stop - no-op.
func (service *Service) Stop() error { return nil }

// Configuration options - none
type Config struct {
}

// NewService creates a new service
func NewService(app App, cfg *Config) (*Service, error) {
	// if no config, error
	// if cfg == nil {
	// 	return nil, errServiceComponent
	// }
	service := new(Service)

	return service, nil
}

// Update Service updates the Service to use the new config
func (service *Service) UpdateService(app App, cfg *Config) error {
	// if no config, error
	// if cfg == nil {
	// 	return errServiceComponent
	// }

	return nil
}
