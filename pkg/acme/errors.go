package acme

import (
	"errors"
	"fmt"
)

var (
	ErrARIUnsupported      = errors.New("acme: server does not support ARI (directory missing 'renewalInfo' key)")
	ErrARIIdentifierFailed = errors.New("acme: failed to generate ari unique identifier")
)

func errorARIIdentifierFailed(detail string) error {
	return fmt.Errorf("%w: %s", ErrARIIdentifierFailed, detail)
}
