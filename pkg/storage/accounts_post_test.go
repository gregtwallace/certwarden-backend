package storage_test

import (
	"certwarden-backend/pkg/domain/acme_accounts"
	"certwarden-backend/pkg/helpers_test"
	"database/sql"
	"fmt"
	"testing"
	"time"
)

func TestPostNewAcmeAccount(t *testing.T) {
	testCases := []struct {
		newPayload      acme_accounts.NewPayload
		expectedPostErr error
		expectedNew     acme_accounts.Account
		expectedGetErr  error
	}{
		{ // valid insertion
			acme_accounts.NewPayload{
				Name:         new("NewAcct"),
				Description:  new("some acct"),
				AcmeServerID: new(1),
				PrivateKeyID: new(58),
				Status:       "a status",
				Email:        new("anemail@example.com"),
				AcceptedTos:  new(false),
				CreatedAt:    time.Unix(1788837479, 0),
				UpdatedAt:    time.Unix(1788838000, 0),
				Kid:          "https://fake.example.com/1234",
			},
			nil,
			acme_accounts.Account{
				ID:          30,
				Name:        "NewAcct",
				Description: "some acct",
				AcmeServer:  acmeServer1,
				AccountKey:  key58,
				Status:      "a status",
				Email:       "anemail@example.com",
				AcceptedTos: false,
				CreatedAt:   time.Unix(1788837479, 0),
				UpdatedAt:   time.Unix(1788838000, 0),
				Kid:         "https://fake.example.com/1234",
			},
			nil,
		},
		{ // duplicate name (non-case sensitive)
			acme_accounts.NewPayload{
				Name:         new("le_staging_account"),
				Description:  new("wont work"),
				AcmeServerID: new(1),
				PrivateKeyID: new(62),
				Status:       "status",
				Email:        new("email@example.com"),
				AcceptedTos:  new(true),
				CreatedAt:    time.Unix(1888837479, 0),
				UpdatedAt:    time.Unix(1888838000, 0),
				Kid:          "https://fake.example.com/123456",
			},
			helpers_test.NewTestErrorStringComp("UNIQUE constraint failed"),
			acme_accounts.Account{},
			sql.ErrNoRows,
		},
		{ // incomplete payload
			acme_accounts.NewPayload{
				Name:         new("its_a_new_acct"),
				Description:  new("wont work 2"),
				AcmeServerID: new(1),
				// PrivateKeyID:
				Status:      "status",
				Email:       new("fake2@example.com"),
				AcceptedTos: new(true),
				CreatedAt:   time.Unix(1888837479, 0),
				UpdatedAt:   time.Unix(1888838000, 0),
				Kid:         "https://fake.example.com/123456",
			},
			helpers_test.NewTestErrorStringComp("NOT NULL constraint failed"),
			acme_accounts.Account{},
			sql.ErrNoRows,
		},
		{ // incomplete payload
			acme_accounts.NewPayload{
				Name:         new("its_a_new_acct"),
				Description:  new("wont work 2"),
				AcmeServerID: new(1),
				PrivateKeyID: new(62),
				Status:       "status",
				// Email:
				AcceptedTos: new(true),
				CreatedAt:   time.Unix(1888837479, 0),
				UpdatedAt:   time.Unix(1888838000, 0),
				Kid:         "https://fake.example.com/123456",
			},
			helpers_test.NewTestErrorStringComp("NOT NULL constraint failed"),
			acme_accounts.Account{},
			sql.ErrNoRows,
		},
	}

	// create testing service
	storage, err := openStorageWithTestData(t, "postnewaccount")
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("post name: %s", helpers_test.StringPointerToVal(tc.newPayload.Name)), func(t *testing.T) {
			acct, err := storage.PostNewAcmeAccount(tc.newPayload)
			if !helpers_test.ErrorsIs(err, tc.expectedPostErr) {
				t.Errorf("expected post error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedPostErr), helpers_test.ErrorToVal(err))
			}

			compareAcmeAccount(t, acct, tc.expectedNew)

			acct, err = storage.GetOneAcmeAccountByName(acct.Name)
			if !helpers_test.ErrorsIs(err, tc.expectedGetErr) {
				t.Errorf("expected get error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedGetErr), helpers_test.ErrorToVal(err))
			}

			compareAcmeAccount(t, acct, tc.expectedNew)
		})
	}
}
