package storage_test

import (
	"certwarden-backend/pkg/domain/acme_accounts"
	"certwarden-backend/pkg/helpers_test"
	"certwarden-backend/pkg/storage"
	"database/sql"
	"fmt"
	"testing"
)

func TestAcmeAccountInUse(t *testing.T) {
	testCases := []struct {
		acctID        int
		expectedInUse bool
		expectedErr   error
	}{
		{-2, false, sql.ErrNoRows},
		{35, false, sql.ErrNoRows},

		{1, true, nil},
		{2, true, nil},
		{20, true, nil},

		{16, false, nil},
		{23, false, nil},
		{28, false, nil},
		{29, false, nil},
	}

	// create testing service
	store := openStorageWithTestData(t, "acmeaccountinuse")

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("id: %d", tc.acctID), func(t *testing.T) {
			inUse, err := store.AcmeAccountInUse(tc.acctID)
			if !helpers_test.ErrorsIs(err, tc.expectedErr) {
				t.Errorf("expected error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedErr), helpers_test.ErrorToVal(err))
			}

			if inUse != tc.expectedInUse {
				t.Errorf("expected inuse '%t' but got '%t'", tc.expectedInUse, inUse)
			}
		})
	}
}

func TestDeleteAcmeAccount(t *testing.T) {
	testCases := []struct {
		acctID            int
		expectedDelErr    error
		expectedGetResult *acme_accounts.Account
		expectedGetErr    error
	}{
		{-2, sql.ErrNoRows, nil, sql.ErrNoRows},  // non-existent
		{25, sql.ErrNoRows, nil, sql.ErrNoRows},  // non-existent
		{16, nil, nil, sql.ErrNoRows},            // not in use, gets deleted
		{28, nil, nil, sql.ErrNoRows},            // not in use, gets deleted
		{1, storage.ErrInUse, &acmeAcct1, nil},   // in use
		{2, storage.ErrInUse, &acmeAcct2, nil},   // in use
		{20, storage.ErrInUse, &acmeAcct20, nil}, // in use
	}

	// create testing service
	store := openStorageWithTestData(t, "deleteacmeaccount")

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("id: %d", tc.acctID), func(t *testing.T) {
			err := store.DeleteAcmeAccount(tc.acctID)
			if !helpers_test.ErrorsIs(err, tc.expectedDelErr) {
				t.Errorf("expected delete error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedDelErr), helpers_test.ErrorToVal(err))
			}

			acct, err := store.GetOneAcmeAccountById(tc.acctID)
			if !helpers_test.ErrorsIs(err, tc.expectedGetErr) {
				t.Errorf("expected get error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedGetErr), helpers_test.ErrorToVal(err))
			}

			compareAcmeAccount(t, acct, tc.expectedGetResult)
		})
	}
}
