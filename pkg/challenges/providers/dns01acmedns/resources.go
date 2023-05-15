package dns01acmedns

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

var (
	ErrDomainNotConfigured = errors.New("dns01acmedns domain name not configured (restart lego if config was updated)")
	ErrUpdateFailed        = errors.New("dns01acmedns failed to update domain ")
)

// acmeDnsResource contains the needed configuration to update
// and acme-dns record and also the corresponding 'real' domain
// that will have a certificate issued for it.
type acmeDnsResource struct {
	RealDomain string `yaml:"real_domain"`
	FullDomain string `yaml:"full_domain"`
	Username   string `yaml:"username"`
	Password   string `yaml:"password"`
}

// subdomain returns just the first piece of the subdomain of FullDomain
func (adr *acmeDnsResource) subdomain() string {
	fd := strings.Split(adr.FullDomain, ".")
	return fd[0]
}

// acme-dns update endpoint
const acmeDnsUpdateEndpoint = "/update"

type acmeDnsUpdate struct {
	SubDomain string `json:"subdomain"`
	Txt       string `json:"txt"`
}

// updateRequest creates an update request for the adr hosted by the
// acmeDnsAddress using the new resource content value specified
func (service *Service) updateRequest(adr *acmeDnsResource, resourceContent string) (*http.Request, error) {
	// body of api call
	payload := acmeDnsUpdate{
		SubDomain: adr.subdomain(),
		Txt:       resourceContent,
	}

	// marshal for posting
	payloadJson, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	// prepare api call to acme-dns to set the txt record
	req, err := service.httpClient.NewRequest(http.MethodPost, service.acmeDnsAddress+acmeDnsUpdateEndpoint, bytes.NewBuffer(payloadJson))
	if err != nil {
		return nil, err
	}

	// set auth headers
	req.Header.Set("X-Api-User", adr.Username)
	req.Header.Set("X-Api-Key", adr.Password)

	return req, nil
}

// Provision updates the acme-dns resource record with the correct content
func (service *Service) Provision(resourceName string, resourceContent string) error {
	// check if resource exists
	adr := acmeDnsResource{}
	found := false
	for _, r := range service.acmeDnsResources {
		if "_acme-challenge."+r.RealDomain == resourceName {
			adr = r
			found = true
			break
		}
	}
	// this package does not support registering new records, so fail if
	// record was not found
	if !found {
		return ErrDomainNotConfigured
	}

	// make request
	req, err := service.updateRequest(&adr, resourceContent)
	if err != nil {
		return err
	}

	// send request
	resp, err := service.httpClient.Do(req)
	if err != nil {
		return err
	}

	// if not status 200
	if resp.StatusCode != http.StatusOK {
		return ErrUpdateFailed
	}

	return nil
}

// Derovision updates the acme-dns resource record with blank content. This probably
// isn't really needed and this could be an empty stub. Clearing the data doesn't
// hurt though.
func (service *Service) Deprovision(resourceName string, resourceContent string) error {
	// check if resource exists
	adr := acmeDnsResource{}
	found := false
	for _, r := range service.acmeDnsResources {
		if "_acme-challenge."+r.RealDomain == resourceName {
			adr = r
			found = true
			break
		}
	}
	// this package does not support registering new records, so fail if
	// record was not found
	if !found {
		return ErrDomainNotConfigured
	}

	// make request (dummy text value when not in use)
	req, err := service.updateRequest(&adr, "VOID_____VOID______VOID_______VOID_____VOID")
	if err != nil {
		return err
	}

	// send request
	resp, err := service.httpClient.Do(req)
	if err != nil {
		return err
	}

	// if not status 200
	if resp.StatusCode != http.StatusOK {
		return ErrUpdateFailed
	}

	return nil
}
