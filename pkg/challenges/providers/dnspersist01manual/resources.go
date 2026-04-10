package dnspersist01manual

import (
	"certwarden-backend/pkg/acme"
)

// Provision is a no-op (records are expected to be manually created and persistent)
func (service *Service) Provision(_ string, _ string, _ acme.KeyAuth) error {
	return nil
}

// Deprovision is a no-op (records are expected to be manually created and persistent)
func (service *Service) Deprovision(_ string, _ string, _ acme.KeyAuth) error {
	return nil
}
