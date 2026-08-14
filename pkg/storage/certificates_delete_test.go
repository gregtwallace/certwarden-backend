package storage_test

import (
	"certwarden-backend/pkg/domain/certificates"
	"certwarden-backend/pkg/helpers_test"
	"database/sql"
	"fmt"
	"testing"
)

func TestDeleteCert(t *testing.T) {
	testCases := []struct {
		id             int
		expectedDelErr error

		expectedGetCert *certificates.Certificate
		expectedGetErr  error
	}{
		{-12, sql.ErrNoRows, nil, sql.ErrNoRows}, // non-existent
		{2, sql.ErrNoRows, nil, sql.ErrNoRows},   // non-existent
		{18, nil, nil, sql.ErrNoRows},            // not in use, gets deleted
		{30, nil, nil, sql.ErrNoRows},            // not in use, gets deleted
		{32, nil, nil, sql.ErrNoRows},            // not in use, gets deleted
		{35, nil, nil, sql.ErrNoRows},            // not in use, gets deleted
	}

	// create testing service
	store, err := openStorageWithTestData(t, "deletecert")
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("id: %d", tc.id), func(t *testing.T) {
			err := store.DeleteCert(tc.id)
			if !helpers_test.ErrorsIs(err, tc.expectedDelErr) {
				t.Errorf("expected delete error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedDelErr), helpers_test.ErrorToVal(err))
			}

			cert, err := store.GetOneCertById(tc.id)
			if !helpers_test.ErrorsIs(err, tc.expectedGetErr) {
				t.Errorf("expected get error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedGetErr), helpers_test.ErrorToVal(err))
			}

			compareCertificate(t, cert, tc.expectedGetCert)
		})
	}
}
