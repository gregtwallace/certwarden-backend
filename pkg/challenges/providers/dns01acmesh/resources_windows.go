//go:build windows

package dns01acmesh

import (
	"certwarden-backend/pkg/acme"
)

// Provision adds the requested DNS record.
func (service *Service) Provision(_ string, _ string, _ acme.KeyAuth) error {
	return errWindows
}

// Deprovision deletes the corresponding DNS record.
func (service *Service) Deprovision(_ string, _ string, _ acme.KeyAuth) error {
	return errWindows
}
