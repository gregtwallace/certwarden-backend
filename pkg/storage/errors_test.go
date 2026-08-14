package storage_test

import (
	"certwarden-backend/pkg/storage"
	"errors"
	"testing"
)

func TestErrorWrongUpdateRowCount(t *testing.T) {
	err := storage.ErrorWrongUpdateRowCount(3, 2)
	if !errors.Is(err, storage.ErrWrongUpdateRowCount) {
		t.Errorf("returned wrong error type")
	}
}
