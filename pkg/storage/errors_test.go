package storage_test

import (
	"certwarden-backend/pkg/storage"
	"errors"
	"testing"
)

func TestErrorWrongRowCount(t *testing.T) {
	err := storage.ErrorWrongRowCount(3, 2)
	if !errors.Is(err, storage.ErrWrongRowCount) {
		t.Errorf("returned wrong error type")
	}
}

func TestErrorWrongAffectedRowCount(t *testing.T) {
	err := storage.ErrorWrongAffectedRowCount(3, 2)
	if !errors.Is(err, storage.ErrWrongAffectedRowCount) {
		t.Errorf("returned wrong error type")
	}
}
