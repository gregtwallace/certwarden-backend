package acme_test

import (
	"certwarden-backend/pkg/acme"
	"errors"
	"testing"
)

func TestErrorARIIdentifierFailed(t *testing.T) {
	err := acme.ErrorARIIdentifierFailed("some issue")
	if !errors.Is(err, acme.ErrARIIdentifierFailed) {
		t.Errorf("returned wrong error type")
	}
}
