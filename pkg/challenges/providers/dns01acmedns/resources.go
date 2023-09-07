package dns01acmedns

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"legocerthub-backend/pkg/output"
	"net/http"
	"strings"
)

// AvailableDomains returns all of the domains that this provider instance can
// provision records for.
func (service *Service) AvailableDomains() []string {
	return service.domains
}

// acmeDnsResource contains the needed configuration to update
// and acme-dns record and also the corresponding 'real' domain
// that will have a certificate issued for it.
type acmeDnsResource struct {
	RealDomain string                `yaml:"real_domain" json:"real_domain"`
	FullDomain string                `yaml:"full_domain" json:"full_domain"`
	Username   output.RedactedString `yaml:"username" json:"username"`
	Password   output.RedactedString `yaml:"password" json:"password"`
}

// acme-dns update endpoint
const acmeDnsUpdateEndpoint = "/update"

type acmeDnsUpdate struct {
	SubDomain string `json:"subdomain"`
	Txt       string `json:"txt"`
}

// postUpdate updates the acmeDnsResource to the specified content value on
// the acme-dns server
func (service *Service) postUpdate(adr *acmeDnsResource, resourceContent string) error {
	// body of api call
	payload := acmeDnsUpdate{
		SubDomain: strings.Split(adr.FullDomain, ".")[0], // expected subdomain value for acme-dns
		Txt:       resourceContent,
	}

	// marshal for posting
	payloadJson, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// set auth headers
	header := make(http.Header)
	header.Set("X-Api-User", adr.Username.Unredacted())
	header.Set("X-Api-Key", adr.Password.Unredacted())

	// post to acme dns
	resp, err := service.httpClient.PostWithHeader(service.acmeDnsAddress+acmeDnsUpdateEndpoint, "application/json", bytes.NewBuffer(payloadJson), header)
	if err != nil {
		return err
	}

	// read body & close
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)

	// if not status 200, error
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("acme-dns failed to update %s (%d)", adr.FullDomain, resp.StatusCode)
	}

	return nil
}

// getAcmeDnsResource returns the acme dns resource for resourceName. If no
// record exists, an error is returned instead
func (service *Service) getAcmeDnsResource(resourceName string) (*acmeDnsResource, error) {
	// remove prepended _acme-challenge. string from resource name
	resourceNameRealDomain := strings.TrimPrefix(resourceName, "_acme-challenge.")

	for i := range service.acmeDnsResources {
		if resourceNameRealDomain == service.acmeDnsResources[i].RealDomain {
			return &service.acmeDnsResources[i], nil
		}
	}

	return nil, fmt.Errorf("acme-dns resource not found for %s", resourceName)
}

// Provision updates the acme-dns resource corresponding to resourceName with
// the resourceContent
func (service *Service) Provision(resourceName, resourceContent string) error {
	// get acme-dns resource
	adr, err := service.getAcmeDnsResource(resourceName)
	if err != nil {
		return err
	}

	// post update
	err = service.postUpdate(adr, resourceContent)
	if err != nil {
		return err
	}

	return nil
}

// Derovision updates the acme-dns resource corresponding to resourceName with
// a dummy value. This probably isn't really needed and this function could just
// be an empty stub, but clearing the data doesn't hurt.
func (service *Service) Deprovision(resourceName, resourceContent string) error {
	// get acme-dns resource
	adr, err := service.getAcmeDnsResource(resourceName)
	if err != nil {
		return err
	}

	// post update (dummy value when not in use)
	err = service.postUpdate(adr, "__VOID__")
	if err != nil {
		return err
	}

	return nil
}
