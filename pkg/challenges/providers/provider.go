package providers

import (
	"certwarden-backend/pkg/acme"
	"time"
)

// providerConfig is the interface provider configs must satisfy
type providerConfig interface{}

// service is an interface for a child provider service
type Service interface {
	AcmeChallengeType() acme.ChallengeType
	Provision(domain string, token string, keyAuth acme.KeyAuth) (err error)
	Deprovision(domain string, token string, keyAuth acme.KeyAuth) (err error)
	Stop() error
}

// provider is the structure of a provider that is being managed
type provider struct {
	ID                       int      `json:"id"`
	Tag                      string   `json:"tag"`
	Type                     string   `json:"type"`
	Domains                  []string `json:"domains"`
	PostProvisionWaitSeconds int      `json:"post_resource_provision_wait"`
	Config                   any      `json:"config"`
	Service                  `json:"-"`
}

// PostProvisionResourceWait returns a duration that should be slept after a resource is
// provisioned, to ensure the resource has completely propagated.
func (p *provider) PostProvisionResourceWait() time.Duration {
	return time.Duration(p.PostProvisionWaitSeconds * int(time.Second))
}
