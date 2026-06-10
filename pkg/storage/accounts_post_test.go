package storage_test

import (
	"certwarden-backend/pkg/domain/acme_accounts"
	"certwarden-backend/pkg/test_helpers"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestPostNewAcmeAccount(t *testing.T) {
	testCases := []struct {
		newPayload     acme_accounts.NewPayload
		expectPostErr  bool
		expectedNew    acme_accounts.Account
		expectedGetErr error
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
				CreatedAt:    1788837479,
				UpdatedAt:    1788838000,
				Kid:          "https://fake.example.com/1234",
			},
			false,
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
				CreatedAt:    1888837479,
				UpdatedAt:    1888838000,
				Kid:          "https://fake.example.com/123456",
			},
			true,
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
				CreatedAt:   1888837479,
				UpdatedAt:   1888838000,
				Kid:         "https://fake.example.com/123456",
			},
			true,
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
				CreatedAt:   1888837479,
				UpdatedAt:   1888838000,
				Kid:         "https://fake.example.com/123456",
			},
			true,
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
		t.Run(fmt.Sprintf("post name: %s", test_helpers.StringPointerToVal(tc.newPayload.Name)), func(t *testing.T) {
			acct, err := storage.PostNewAcmeAccount(tc.newPayload)
			if (err != nil && !tc.expectPostErr) || (err == nil && tc.expectPostErr) {
				t.Errorf("expected post error '%t' but got err '%s'", tc.expectPostErr, test_helpers.ErrorToVal(err))
			}

			CompareAcmeAccount(t, acct, tc.expectedNew)

			acct, err = storage.GetOneAcmeAccountByName(acct.Name)
			if !errors.Is(err, tc.expectedGetErr) {
				t.Errorf("expected get error '%s' but got '%s'", test_helpers.ErrorToVal(tc.expectedGetErr), test_helpers.ErrorToVal(err))
			}

			CompareAcmeAccount(t, acct, tc.expectedNew)
		})
	}
}
