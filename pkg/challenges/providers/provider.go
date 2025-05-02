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
	ID                   int      `json:"id"`
	Tag                  string   `json:"tag"`
	Type                 string   `json:"type"`
	Domains              []string `json:"domains"`
	PreCheckWaitSeconds  int      `json:"precheck_wait"`
	PostCheckWaitSeconds int      `json:"postcheck_wait"`
	Config               any      `json:"config"`
	Service              `json:"-"`
}

// WaitDurationPreResourceCheck returns a duration that should be slept after a resource
// is provisioned, but before checks are performed to confirm the existence of the resource.
// This is useful to avoid unncessary early checking if it is known the resoucres take some
// minimum amount of time to propagate.
func (p *provider) WaitDurationPreResourceCheck() time.Duration {
	return time.Duration(p.PreCheckWaitSeconds * int(time.Second))
}

// WaitDurationPostResourceCheck returns a duration that should be slept after a resource
// is provisioned, and after checks are performed to confirm the existence of the resource
// and those checks confirmed the existence of the resource.
// This is useful to ensure full resource propation, such as cases where the checks may
// have confirmed existence but some additional time is desired to really make sure things
// completely propagated.
func (p *provider) WaitDurationPostResourceCheck() time.Duration {
	return time.Duration(p.PostCheckWaitSeconds * int(time.Second))
}
