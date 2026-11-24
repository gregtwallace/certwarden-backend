package storage_test

import (
	"certwarden-backend/pkg/domain/acme_servers"
	"certwarden-backend/pkg/storage"
	"certwarden-backend/pkg/test_helpers"
	"database/sql"
	"errors"
	"fmt"
	"testing"
)

func TestServerInUse(t *testing.T) {
	testCases := []struct {
		serverID      int
		expectedInUse bool
		expectedErr   error
	}{
		{2, false, sql.ErrNoRows},
		{35, false, sql.ErrNoRows},
		{0, true, nil},
		{1, true, nil},
		{4, false, nil},
		{19, true, nil},
		{20, false, nil},
	}

	// create testing service
	storage, err := openStorageWithTestData(t, "serverinuse")
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("server id: %d", tc.serverID), func(t *testing.T) {
			inUse, err := storage.ServerInUse(tc.serverID)
			if !errors.Is(err, tc.expectedErr) {
				t.Errorf("expected error '%s' but got '%s'", test_helpers.ErrorToVal(tc.expectedErr), test_helpers.ErrorToVal(err))
			}

			if inUse != tc.expectedInUse {
				t.Errorf("expected inuse '%t' but got '%t'", tc.expectedInUse, inUse)
			}
		})
	}
}

func TestDeleteServer(t *testing.T) {
	testCases := []struct {
		serverID          int
		expectedDelErr    error
		expectedGetResult acme_servers.Server
		expectedGetErr    error
	}{
		{2, sql.ErrNoRows, acme_servers.Server{}, sql.ErrNoRows},  // non-existent
		{25, sql.ErrNoRows, acme_servers.Server{}, sql.ErrNoRows}, // non-existent
		{4, nil, acme_servers.Server{}, sql.ErrNoRows},            // not in use, gets deleted
		{20, nil, acme_servers.Server{}, sql.ErrNoRows},           // not in use, gets deleted
		{0, storage.ErrInUse, acmeServer0, nil},                   // in use
		{1, storage.ErrInUse, acmeServer1, nil},                   // in use
		{19, storage.ErrInUse, acmeServer19, nil},                 // in use
	}

	// create testing service
	storage, err := openStorageWithTestData(t, "deleteserver")
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("server id: %d", tc.serverID), func(t *testing.T) {
			err := storage.DeleteServer(tc.serverID)
			if !errors.Is(err, tc.expectedDelErr) {
				t.Errorf("expected delete error '%s' but got '%s'", test_helpers.ErrorToVal(tc.expectedDelErr), test_helpers.ErrorToVal(err))
			}

			server, err := storage.GetOneServerById(tc.serverID)
			if !errors.Is(err, tc.expectedGetErr) {
				t.Errorf("expected get error '%s' but got '%s'", test_helpers.ErrorToVal(tc.expectedGetErr), test_helpers.ErrorToVal(err))
			}

			CompareAcmeServer(t, server, tc.expectedGetResult)
		})
	}
}
