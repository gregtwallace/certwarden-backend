package storage_test

import (
	"certwarden-backend/pkg/test_helpers"
	"database/sql"
	"errors"
	"fmt"
	"testing"
)

func TestKeyInUse(t *testing.T) {
	testCases := []struct {
		keyID         int
		expectedInUse bool
		expectedErr   error
	}{
		{2, false, sql.ErrNoRows},
		{58, false, nil},
		{62, false, nil},
		{1, true, nil},  // acct
		{4, true, nil},  // acct
		{31, true, nil}, // cert
		{55, true, nil}, // cert
		{64, true, nil}, // cert
		{69, true, nil}, // newest order (but isn't in use on a cert)
	}

	// create testing service
	storage, err := openStorageWithTestData(t, "keyinuse")
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("key id: %d", tc.keyID), func(t *testing.T) {
			inUse, err := storage.KeyInUse(tc.keyID)
			if !errors.Is(err, tc.expectedErr) {
				t.Errorf("expected error '%s' but got '%s'", tc.expectedErr, test_helpers.ErrorToVal(err))
			}

			if inUse != tc.expectedInUse {
				t.Errorf("key id expected inuse '%t' but got '%t'", tc.expectedInUse, inUse)
			}
		})
	}
}
