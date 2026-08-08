package storage_test

import (
	"certwarden-backend/pkg/domain/private_keys"
	"certwarden-backend/pkg/helpers_test"
	"certwarden-backend/pkg/storage"
	"database/sql"
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
		t.Run(fmt.Sprintf("id: %d", tc.keyID), func(t *testing.T) {
			inUse, err := storage.KeyInUse(tc.keyID)
			if !helpers_test.ErrorsIs(err, tc.expectedErr) {
				t.Errorf("expected error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedErr), helpers_test.ErrorToVal(err))
			}

			if inUse != tc.expectedInUse {
				t.Errorf("expected inuse '%t' but got '%t'", tc.expectedInUse, inUse)
			}
		})
	}
}

func TestDeleteKey(t *testing.T) {
	testCases := []struct {
		keyID             int
		expectedDelErr    error
		expectedGetResult private_keys.Key
		expectedGetErr    error
	}{
		{2, sql.ErrNoRows, private_keys.Key{}, sql.ErrNoRows}, // non-existent
		{58, nil, private_keys.Key{}, sql.ErrNoRows},          // not in use, gets deleted
		{62, nil, private_keys.Key{}, sql.ErrNoRows},          // not in use, gets deleted
		{63, storage.ErrInUse, key63, nil},                    // in use by acct
		{64, storage.ErrInUse, key64, nil},                    // in use by cert
		{69, storage.ErrInUse, key69, nil},                    // in use by newest order (but isn't in use on a cert)
	}

	// create testing service
	storage, err := openStorageWithTestData(t, "deletekey")
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("id: %d", tc.keyID), func(t *testing.T) {
			err := storage.DeleteKey(tc.keyID)
			if !helpers_test.ErrorsIs(err, tc.expectedDelErr) {
				t.Errorf("expected delete error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedDelErr), helpers_test.ErrorToVal(err))
			}

			key, err := storage.GetOneKeyById(tc.keyID)
			if !helpers_test.ErrorsIs(err, tc.expectedGetErr) {
				t.Errorf("expected get error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedGetErr), helpers_test.ErrorToVal(err))
			}

			CompareKey(t, key, tc.expectedGetResult)
		})
	}
}
