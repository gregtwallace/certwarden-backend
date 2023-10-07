package providers

import (
	"legocerthub-backend/pkg/acme"
)

// providerConfig is the interface provider configs must satisfy
type providerConfig interface {
	Domains() []string
}

// service is an interface for a child provider service
type Service interface {
	AcmeChallengeType() acme.ChallengeType
	Provision(resourceName, resourceContent string) (err error)
	Deprovision(resourceName, resourceContent string) (err error)
	Stop() error
}

// provider is the structure of a provider that is being managed
type provider struct {
	ID      int    `json:"id"`
	Tag     string `json:"tag"`
	Type    string `json:"type"`
	Config  any    `json:"config"`
	Service `json:"-"`
}
