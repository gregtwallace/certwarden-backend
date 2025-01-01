package dns01acmedns

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"certwarden-backend/pkg/acme"
)

// acmeDnsResource contains the needed configuration to update
// and acme-dns record and also the corresponding 'real' domain
// that will have a certificate issued for it.
type acmeDnsResource struct {
	RealDomain string `yaml:"real_domain" json:"real_domain"`
	FullDomain string `yaml:"full_domain" json:"full_domain"`
	Username   string `yaml:"username" json:"username"`
	Password   string `yaml:"password" json:"password"`
}

// acme-dns update endpoint
const acmeDnsUpdateEndpoint = "/update"

type acmeDnsUpdate struct {
	SubDomain string `json:"subdomain"`
	Txt       string `json:"txt"`
}

// postUpdate updates the acmeDnsResource to the specified content value on
// the acme-dns server
func (service *Service) postUpdate(adr *acmeDnsResource, dnsRecordValue string) error {
	// body of api call
	payload := acmeDnsUpdate{
		SubDomain: strings.Split(adr.FullDomain, ".")[0], // expected subdomain value for acme-dns
		Txt:       dnsRecordValue,
	}

	// marshal for posting
	payloadJson, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// make request
	req, err := http.NewRequest(http.MethodPost, service.acmeDnsAddress+acmeDnsUpdateEndpoint, bytes.NewBuffer(payloadJson))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-User", adr.Username)
	req.Header.Set("X-Api-Key", adr.Password)

	// post to acme dns
	resp, err := service.httpClient.Do(req)
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

// getAcmeDnsResource returns the acme dns resource for domain. If no
// record exists, an error is returned instead
func (service *Service) getAcmeDnsResource(domain string) (*acmeDnsResource, error) {
	for i := range service.acmeDnsResources {
		if domain == service.acmeDnsResources[i].RealDomain {
			return &service.acmeDnsResources[i], nil
		}
	}

	return nil, fmt.Errorf("acme-dns resource not found for %s", domain)
}

// Provision updates the acme-dns resource corresponding to domain with
// the new value calculated from keyAuth
func (service *Service) Provision(domain string, _ string, keyAuth acme.KeyAuth) error {
	// get acme-dns resource
	adr, err := service.getAcmeDnsResource(domain)
	if err != nil {
		return err
	}

	// get dns record value for update
	_, dnsRecordValue := acme.ValidationResourceDns01(domain, keyAuth)

	// post update
	err = service.postUpdate(adr, dnsRecordValue)
	if err != nil {
		return err
	}

	return nil
}

// Derovision updates the acme-dns resource corresponding to domain with
// a dummy value. This probably isn't really needed and this function could just
// be an empty stub, but clearing the data doesn't hurt.
func (service *Service) Deprovision(domain string, _ string, _ acme.KeyAuth) error {
	// get acme-dns resource
	adr, err := service.getAcmeDnsResource(domain)
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
