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
		id                int
		expectedDelErr    error
		expectedGetResult certificates.Certificate
		expectedGetErr    error
	}{
		{-12, sql.ErrNoRows, certificates.Certificate{}, sql.ErrNoRows}, // non-existent
		{2, sql.ErrNoRows, certificates.Certificate{}, sql.ErrNoRows},   // non-existent
		{18, nil, certificates.Certificate{}, sql.ErrNoRows},            // not in use, gets deleted (Maybe TODO: Prevent delete from app to server's ssl cert)
		{30, nil, certificates.Certificate{}, sql.ErrNoRows},            // not in use, gets deleted
		{32, nil, certificates.Certificate{}, sql.ErrNoRows},            // not in use, gets deleted
		{35, nil, certificates.Certificate{}, sql.ErrNoRows},            // not in use, gets deleted
	}

	// create testing service
	storage, err := openStorageWithTestData(t, "deletecert")
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("id: %d", tc.id), func(t *testing.T) {
			err := storage.DeleteCert(tc.id)
			if !helpers_test.ErrorsIs(err, tc.expectedDelErr) {
				t.Errorf("expected delete error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedDelErr), helpers_test.ErrorToVal(err))
			}

			cert, err := storage.GetOneCertById(tc.id)
			if !helpers_test.ErrorsIs(err, tc.expectedGetErr) {
				t.Errorf("expected get error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedGetErr), helpers_test.ErrorToVal(err))
			}

			CompareCertificate(t, cert, tc.expectedGetResult)
		})
	}
}
