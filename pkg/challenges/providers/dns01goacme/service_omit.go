//go:build omitdns01goacme

package dns01goacme

import (
	"errors"
)

var errUnimplemented = errors.New("dns-01 go-acme is not available (omitted from this build)")

// fakeProvider is an unimplemented provider that only returns errors
type fakeProvider struct{}

func (fp *fakeProvider) Present(domain, token, keyAuth string) error {
	return errUnimplemented
}
func (fp *fakeProvider) CleanUp(domain, token, keyAuth string) error {
	return errUnimplemented
}

// NewService creates a new dummy service that only returns unimplemented errors
func NewService(app App, cfg *Config) (*Service, error) {
	// if no config, error
	if cfg == nil {
		return nil, errServiceComponent
	}

	service := new(Service)

	// logger
	service.logger = app.GetLogger()
	if service.logger == nil {
		return nil, errServiceComponent
	}

	// make go acme provider
	service.goacmeProvider = &fakeProvider{}

	return service, nil
}
