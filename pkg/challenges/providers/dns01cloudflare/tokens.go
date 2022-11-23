package dns01cloudflare

import "errors"

func (service *Service) Provision(resourceName string, resourceContent string) error {
	return errors.New("dns-01 cloudflare provisioning not yet implemented")
}

func (service *Service) Deprovision(resourceName string) error {
	return errors.New("dns-01 cloudflare deprovisioning not yet implemented")
}
