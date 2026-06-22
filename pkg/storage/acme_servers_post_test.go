package storage_test

import (
	"certwarden-backend/pkg/domain/acme_servers"
	"certwarden-backend/pkg/test_helpers"
	"database/sql"
	"fmt"
	"testing"
	"time"
)

func TestPostNewServer(t *testing.T) {
	testCases := []struct {
		newPayload      acme_servers.NewPayload
		expectedPostErr error
		expectedNew     acme_servers.Server
		expectedGetErr  error
	}{
		{ // valid insertion
			acme_servers.NewPayload{
				Name:         new("NewServer"),
				Description:  new("some service"),
				DirectoryURL: new("https://example.com/directory"),
				IsStaging:    new(true),
				CreatedAt:    1780337479,
				UpdatedAt:    1780338000,
			},
			nil,
			acme_servers.Server{
				ID:           21,
				Name:         "NewServer",
				Description:  "some service",
				DirectoryURL: "https://example.com/directory",
				IsStaging:    true,
				CreatedAt:    time.Unix(1780337479, 0),
				UpdatedAt:    time.Unix(1780338000, 0),
			},
			nil,
		},
		{ // duplicate name (non-case sensitive)
			acme_servers.NewPayload{
				Name:         new("lets_encrypt_staging"),
				Description:  new("some service wont work"),
				DirectoryURL: new("https://example2.com/directory"),
				IsStaging:    new(true),
				CreatedAt:    1780337449,
				UpdatedAt:    1780338040,
			},
			test_helpers.MakeTestErrorStringComp("UNIQUE constraint failed"),
			acme_servers.Server{},
			sql.ErrNoRows,
		},
		{ // incomplete payload
			acme_servers.NewPayload{
				Name:        new("its_a_new_server"),
				Description: new("wont work"),
				// DirectoryURL
				IsStaging: new(false),
				CreatedAt: 1880337449,
				UpdatedAt: 1880338040,
			},
			test_helpers.MakeTestErrorStringComp("NOT NULL constraint failed"),
			acme_servers.Server{},
			sql.ErrNoRows,
		},
	}

	// create testing service
	storage, err := openStorageWithTestData(t, "postnewserver")
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("post name: %s", test_helpers.StringPointerToVal(tc.newPayload.Name)), func(t *testing.T) {
			server, err := storage.PostNewServer(tc.newPayload)
			if !test_helpers.ErrorsIs(err, tc.expectedPostErr) {
				t.Errorf("expected post error '%s' but got '%s'", test_helpers.ErrorToVal(tc.expectedPostErr), test_helpers.ErrorToVal(err))
			}

			CompareAcmeServer(t, server, tc.expectedNew)

			server, err = storage.GetOneServerByName(server.Name)
			if !test_helpers.ErrorsIs(err, tc.expectedGetErr) {
				t.Errorf("expected get error '%s' but got '%s'", test_helpers.ErrorToVal(tc.expectedGetErr), test_helpers.ErrorToVal(err))
			}

			CompareAcmeServer(t, server, tc.expectedNew)
		})
	}
}
