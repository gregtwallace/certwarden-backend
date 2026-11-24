//go:build windows

package dns01acmesh

import "errors"

var errWindows = errors.New("acme.sh is not supported in windows, disable it")

func NewService(app App, cfg *Config) (*Service, error) {
	return nil, errWindows
}
