//go:build windows

package dns01acmesh

import (
	"certwarden-backend/pkg/acme"
	"errors"
)

var errWindows = errors.New("acme.sh is not supported in windows")

// acme.sh doesn't work in windows, so implement dummy code that doesn't do anything

type App interface {
}

type Service struct {
}

func (service *Service) AcmeChallengeType() acme.ChallengeType {
	return acme.ChallengeTypeUnknown
}

func (service *Service) Stop() error { return errWindows }

// Configuration options
type Config struct {
}

func NewService(app App, cfg *Config) (*Service, error) {
	return nil, errWindows
}

func (service *Service) UpdateService(app App, cfg *Config) error {
	return errWindows
}
