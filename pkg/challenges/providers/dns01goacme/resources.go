package dns01goacme

// Provision adds the corresponding DNS record. It essentially just calls go-acme's
// provider "Present" function
func (service *Service) Provision(domain, token, keyAuth string) error {
	return service.goacmeProvider.Present(domain, token, keyAuth)
}

// Provision adds the corresponding DNS record. It essentially just calls go-acme's
// provider "Cleanup" function
func (service *Service) Deprovision(domain, token, keyAuth string) error {
	return service.goacmeProvider.CleanUp(domain, token, keyAuth)
}
