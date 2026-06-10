package storage_test

import (
	"certwarden-backend/pkg/domain/acme_accounts"
	"certwarden-backend/pkg/storage"
	"certwarden-backend/pkg/test_helpers"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestPutAcmeAccountNameDesc(t *testing.T) {
	testCases := []struct {
		payload           acme_accounts.NameDescPayload
		expectedPutResult acme_accounts.Account
		expectedPutErr    error
		getId             int
		expectedGetResult acme_accounts.Account
		expectedGetErr    error
	}{
		{ // invalid acct
			acme_accounts.NameDescPayload{
				ID: -1,
			},
			acme_accounts.Account{},
			storage.ErrWrongUpdateRowCount,
			-1,
			acme_accounts.Account{},
			sql.ErrNoRows,
		},
		{ // invalid server
			acme_accounts.NameDescPayload{
				ID: 522,
			},
			acme_accounts.Account{},
			storage.ErrWrongUpdateRowCount,
			522,
			acme_accounts.Account{},
			sql.ErrNoRows,
		},
		{ // update all things
			acme_accounts.NameDescPayload{
				ID:          1,
				Name:        new("update-1"),
				Description: new("new desc 1"),
				UpdatedAt:   1000265751,
			},
			acme_accounts.Account{
				ID:          1,
				Name:        "update-1",
				Description: "new desc 1",
				AcmeServer:  acmeServer1,
				AccountKey:  key1,
				Status:      "valid",
				Email:       "",
				AcceptedTos: true,
				CreatedAt:   time.Unix(1697142144, 0),
				UpdatedAt:   time.Unix(1000265751, 0),
				Kid:         "https://acme-staging-v02.api.letsencrypt.org/acme/acct/red-1",
			},
			nil,
			1,
			acme_accounts.Account{
				ID:          1,
				Name:        "update-1",
				Description: "new desc 1",
				AcmeServer:  acmeServer1,
				AccountKey:  key1,
				Status:      "valid",
				Email:       "",
				AcceptedTos: true,
				CreatedAt:   time.Unix(1697142144, 0),
				UpdatedAt:   time.Unix(1000265751, 0),
				Kid:         "https://acme-staging-v02.api.letsencrypt.org/acme/acct/red-1",
			},
			nil,
		},
		{ // update none of the things (except last update)
			acme_accounts.NameDescPayload{
				ID:        2,
				UpdatedAt: 11021111,
			},
			acme_accounts.Account{
				ID:          2,
				Name:        "LE_Production_Account",
				Description: "LE Prod Account - Primary",
				AcmeServer:  acmeServer0,
				AccountKey:  key4,
				Status:      "valid",
				Email:       "fake@example.com",
				AcceptedTos: true,
				CreatedAt:   time.Unix(1697142971, 0),
				UpdatedAt:   time.Unix(11021111, 0),
				Kid:         "https://acme-v02.api.letsencrypt.org/acme/acct/red-2",
			},
			nil,
			2,
			acme_accounts.Account{
				ID:          2,
				Name:        "LE_Production_Account",
				Description: "LE Prod Account - Primary",
				AcmeServer:  acmeServer0,
				AccountKey:  key4,
				Status:      "valid",
				Email:       "fake@example.com",
				AcceptedTos: true,
				CreatedAt:   time.Unix(1697142971, 0),
				UpdatedAt:   time.Unix(11021111, 0),
				Kid:         "https://acme-v02.api.letsencrypt.org/acme/acct/red-2",
			},
			nil,
		},
		{ // update just desc
			acme_accounts.NameDescPayload{
				ID:          23,
				Description: new("new thing"),
				UpdatedAt:   107800111,
			},
			acme_accounts.Account{
				ID:          23,
				Name:        "Google_Cloud_Staging2",
				Description: "new thing",
				AcmeServer:  acmeServer19,
				AccountKey:  key66,
				Status:      "deactivated",
				Email:       "fake2@example.com",
				AcceptedTos: false,
				CreatedAt:   time.Unix(1752416890, 0),
				UpdatedAt:   time.Unix(107800111, 0),
				Kid:         "https://dv.acme-v02.test-api.pki.goog/account/red-23",
			},
			nil,
			23,
			acme_accounts.Account{
				ID:          23,
				Name:        "Google_Cloud_Staging2",
				Description: "new thing",
				AcmeServer:  acmeServer19,
				AccountKey:  key66,
				Status:      "deactivated",
				Email:       "fake2@example.com",
				AcceptedTos: false,
				CreatedAt:   time.Unix(1752416890, 0),
				UpdatedAt:   time.Unix(107800111, 0),
				Kid:         "https://dv.acme-v02.test-api.pki.goog/account/red-23",
			},
			nil,
		},
		{ // conflicting name change fail
			acme_accounts.NameDescPayload{
				ID:        16,
				Name:      new("le_production_account"),
				UpdatedAt: 107800777,
			},
			acme_accounts.Account{},
			test_helpers.ErrAnyType,
			16,
			acmeAcct16,
			nil,
		},
	}

	// create testing service
	storage, err := openStorageWithTestData(t, "putacctupdatenamedesc")
	if err != nil {
		t.Fatal(err)
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("#%d (id: %d)", i, tc.payload.ID), func(t *testing.T) {
			acct, err := storage.PutAcmeAccountNameDesc(tc.payload)
			if !test_helpers.ErrorsIs(err, tc.expectedPutErr) {
				t.Errorf("expected put error '%s' but got '%s'", test_helpers.ErrorToVal(tc.expectedPutErr), test_helpers.ErrorToVal(err))
			}

			CompareAcmeAccount(t, acct, tc.expectedPutResult)

			acct, err = storage.GetOneAcmeAccountById(tc.getId)
			if !errors.Is(err, tc.expectedGetErr) {
				t.Errorf("expected get error '%s' but got '%s'", test_helpers.ErrorToVal(tc.expectedGetErr), test_helpers.ErrorToVal(err))
			}

			CompareAcmeAccount(t, acct, tc.expectedGetResult)
		})
	}
}
