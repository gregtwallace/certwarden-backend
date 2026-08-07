package storage_test

import (
	"certwarden-backend/pkg/domain/acme_servers"
	"certwarden-backend/pkg/helpers_test"
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
				CreatedAt:    time.Unix(1780337479, 0),
				UpdatedAt:    time.Unix(1780338000, 0),
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
				CreatedAt:    time.Unix(1780337449, 0),
				UpdatedAt:    time.Unix(1780338040, 0),
			},
			helpers_test.NewTestErrorStringComp("UNIQUE constraint failed"),
			acme_servers.Server{},
			sql.ErrNoRows,
		},
		{ // incomplete payload
			acme_servers.NewPayload{
				Name:        new("its_a_new_server"),
				Description: new("wont work"),
				// DirectoryURL
				IsStaging: new(false),
				CreatedAt: time.Unix(1880337449, 0),
				UpdatedAt: time.Unix(1880338040, 0),
			},
			helpers_test.NewTestErrorStringComp("NOT NULL constraint failed"),
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
		t.Run(fmt.Sprintf("post name: %s", helpers_test.StringPointerToVal(tc.newPayload.Name)), func(t *testing.T) {
			server, err := storage.PostNewServer(tc.newPayload)
			if !helpers_test.ErrorsIs(err, tc.expectedPostErr) {
				t.Errorf("expected post error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedPostErr), helpers_test.ErrorToVal(err))
			}

			CompareAcmeServer(t, server, tc.expectedNew)

			server, err = storage.GetOneServerByName(server.Name)
			if !helpers_test.ErrorsIs(err, tc.expectedGetErr) {
				t.Errorf("expected get error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedGetErr), helpers_test.ErrorToVal(err))
			}

			CompareAcmeServer(t, server, tc.expectedNew)
		})
	}
}
