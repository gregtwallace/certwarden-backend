package backup_test

import (
	"certwarden-backend/pkg/domain/app/backup"
	"errors"
	"testing"
)

func TestErrorFileError(t *testing.T) {
	err := backup.ErrorFileError("somefile.nfo", "create")
	if !errors.Is(err, backup.ErrFileError) {
		t.Errorf("returned wrong error type")
	}
}
