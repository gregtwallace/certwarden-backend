package helpers_test

import (
	"os"
	"testing"
)

// MakeTempStorage creates a temporary storage folder and registers appropriate
// cleanup tasks.
func MakeTempStorage(t *testing.T, path string) {
	_, err := os.Stat(path)
	if err == nil {
		err := os.RemoveAll(path)
		if err != nil {
			t.Errorf("failed to delete '%s'", path)
		}
	} else if !ErrorsIs(err, os.ErrNotExist) {
		t.Fatalf("failed to stat temp folder '%s'", path)
	}

	err = os.MkdirAll(path, os.FileMode(0o777))
	if err != nil {
		t.Fatalf("failed to make temp folder '%s'", path)
	}
	t.Cleanup(func() {
		err := os.RemoveAll(path)
		if err != nil {
			t.Errorf("failed to delete '%s'", path)
		}
	})
}
