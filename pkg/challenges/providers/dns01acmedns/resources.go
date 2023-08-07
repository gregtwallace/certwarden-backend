package dns01acmedns

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
)

var (
	errDomainNotConfigured = errors.New("dns01acmedns domain name not configured (restart lego if config was updated)")
	errUpdateFailed        = errors.New("dns01acmedns failed to update domain ")
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

// postUpdate posts the requested update to the acme dns server
func (service *Service) postUpdate(adr *acmeDnsResource, resourceContent string) (*http.Response, error) {
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

	// set auth headers
	header := make(http.Header)
	header.Set("X-Api-User", adr.Username)
	header.Set("X-Api-Key", adr.Password)

	// post to acme dns
	resp, err := service.httpClient.PostWithHeader(service.acmeDnsAddress+acmeDnsUpdateEndpoint, "application/json", bytes.NewBuffer(payloadJson), header)
	if err != nil {
		return nil, err
	}

	return resp, nil
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
		return errDomainNotConfigured
	}

	// post update
	resp, err := service.postUpdate(&adr, resourceContent)
	if err != nil {
		return err
	}

	// read body & close
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)

	// if not status 200
	if resp.StatusCode != http.StatusOK {
		return errUpdateFailed
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
		return errDomainNotConfigured
	}

	// make request (dummy text value when not in use)
	resp, err := service.postUpdate(&adr, "VOID_____VOID______VOID_______VOID_____VOID")
	if err != nil {
		return err
	}

	// read body & close
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)

	// if not status 200
	if resp.StatusCode != http.StatusOK {
		return errUpdateFailed
	}

	return nil
}
