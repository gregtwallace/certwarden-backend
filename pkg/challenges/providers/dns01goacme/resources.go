package dns01goacme

import "certwarden-backend/pkg/acme"

// Provision adds the corresponding DNS record. It essentially just calls go-acme's
// provider "Present" function
func (service *Service) Provision(domain string, token string, keyAuth acme.KeyAuth) error {
	return service.goacmeProvider.Present(domain, token, string(keyAuth))
}

// Provision adds the corresponding DNS record. It essentially just calls go-acme's
// provider "Cleanup" function
func (service *Service) Deprovision(domain string, token string, keyAuth acme.KeyAuth) error {
	return service.goacmeProvider.CleanUp(domain, token, string(keyAuth))
}
