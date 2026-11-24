package storage_test

import (
	"certwarden-backend/pkg/domain/acme_servers"
	"certwarden-backend/pkg/storage"
	"certwarden-backend/pkg/test_helpers"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestPutServerUpdate(t *testing.T) {
	testCases := []struct {
		payload           acme_servers.UpdatePayload
		expectedPutResult acme_servers.Server
		expectedPutErr    error
		getId             int
		expectedGetResult acme_servers.Server
		expectedGetErr    error
	}{
		{ // invalid server
			acme_servers.UpdatePayload{
				ID: -1,
			},
			acme_servers.Server{},
			storage.ErrWrongUpdateRowCount,
			-1,
			acme_servers.Server{},
			sql.ErrNoRows,
		},
		{ // invalid server
			acme_servers.UpdatePayload{
				ID: 522,
			},
			acme_servers.Server{},
			storage.ErrWrongUpdateRowCount,
			-1,
			acme_servers.Server{},
			sql.ErrNoRows,
		},
		{ // update all things
			acme_servers.UpdatePayload{
				ID:           1,
				Name:         new("Updated"),
				Description:  new("new desc"),
				DirectoryURL: new("https://example-new.com/directory"),
				IsStaging:    new(false),
				UpdatedAt:    1733265750,
			},
			acme_servers.Server{
				ID:           1,
				Name:         "Updated",
				Description:  "new desc",
				DirectoryURL: "https://example-new.com/directory",
				IsStaging:    false,
				CreatedAt:    time.Unix(1697139774, 0),
				UpdatedAt:    time.Unix(1733265750, 0),
			},
			nil,
			1,
			acme_servers.Server{
				ID:           1,
				Name:         "Updated",
				Description:  "new desc",
				DirectoryURL: "https://example-new.com/directory",
				IsStaging:    false,
				CreatedAt:    time.Unix(1697139774, 0),
				UpdatedAt:    time.Unix(1733265750, 0),
			},
			nil,
		},
		{ // update none of the things (except last update)
			acme_servers.UpdatePayload{
				ID:        19,
				UpdatedAt: 11121111,
			},
			acme_servers.Server{
				ID:           19,
				Name:         "Google_Cloud_Staging",
				Description:  "Google Cloud PreProd",
				DirectoryURL: "https://dv.acme-v02.test-api.pki.goog/directory",
				IsStaging:    true,
				CreatedAt:    time.Unix(1745080146, 0),
				UpdatedAt:    time.Unix(11121111, 0),
			},
			nil,
			19,
			acme_servers.Server{
				ID:           19,
				Name:         "Google_Cloud_Staging",
				Description:  "Google Cloud PreProd",
				DirectoryURL: "https://dv.acme-v02.test-api.pki.goog/directory",
				IsStaging:    true,
				CreatedAt:    time.Unix(1745080146, 0),
				UpdatedAt:    time.Unix(11121111, 0),
			},
			nil,
		},
		{ // update just directory URL
			acme_servers.UpdatePayload{
				ID:           4,
				DirectoryURL: new("https://example-put.com/directory"),
				UpdatedAt:    100800111,
			},
			acme_servers.Server{
				ID:           4,
				Name:         "Google_Prod",
				Description:  "",
				DirectoryURL: "https://example-put.com/directory",
				IsStaging:    false,
				CreatedAt:    time.Unix(1699565933, 0),
				UpdatedAt:    time.Unix(100800111, 0),
			},
			nil,
			4,
			acme_servers.Server{
				ID:           4,
				Name:         "Google_Prod",
				Description:  "",
				DirectoryURL: "https://example-put.com/directory",
				IsStaging:    false,
				CreatedAt:    time.Unix(1699565933, 0),
				UpdatedAt:    time.Unix(100800111, 0),
			},
			nil,
		},
	}

	// create testing service
	storage, err := openStorageWithTestData(t, "putserverupdate")
	if err != nil {
		t.Fatal(err)
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("#%d (id: %d)", i, tc.payload.ID), func(t *testing.T) {
			server, err := storage.PutServerUpdate(tc.payload)
			if !errors.Is(err, tc.expectedPutErr) {
				t.Errorf("expected put error '%s' but got '%s'", test_helpers.ErrorToVal(tc.expectedPutErr), test_helpers.ErrorToVal(err))
			}

			CompareAcmeServer(t, server, tc.expectedPutResult)

			server, err = storage.GetOneServerById(tc.getId)
			if !errors.Is(err, tc.expectedGetErr) {
				t.Errorf("expected get error '%s' but got '%s'", test_helpers.ErrorToVal(tc.expectedGetErr), test_helpers.ErrorToVal(err))
			}

			CompareAcmeServer(t, server, tc.expectedGetResult)
		})
	}
}
