package storage_test

import (
	"certwarden-backend/pkg/domain/acme_accounts"
	"certwarden-backend/pkg/helpers_test"
	"certwarden-backend/pkg/storage"
	"database/sql"
	"fmt"
	"testing"
	"time"
)

func TestPutAcmeAccountUpdate(t *testing.T) {
	testCases := []struct {
		payload           acme_accounts.UpdatePayload
		expectedPutResult *acme_accounts.Account
		expectedPutErr    error
		getId             int
		expectedGetResult *acme_accounts.Account
		expectedGetErr    error
	}{
		{ // invalid
			acme_accounts.UpdatePayload{
				ID: -1,
			},
			nil,
			storage.ErrWrongUpdateRowCount,
			-1,
			nil,
			sql.ErrNoRows,
		},
		{ // invalid
			acme_accounts.UpdatePayload{
				ID: 522,
			},
			nil,
			storage.ErrWrongUpdateRowCount,
			522,
			nil,
			sql.ErrNoRows,
		},
		{ // update all things
			acme_accounts.UpdatePayload{
				ID:          1,
				Name:        new("update-1"),
				Description: new("new desc 1"),
				UpdatedAt:   time.Unix(1000265751, 0),
			},
			&acme_accounts.Account{
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
			&acme_accounts.Account{
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
			acme_accounts.UpdatePayload{
				ID:        2,
				UpdatedAt: time.Unix(11021111, 0),
			},
			&acme_accounts.Account{
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
			&acme_accounts.Account{
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
			acme_accounts.UpdatePayload{
				ID:          23,
				Description: new("new thing"),
				UpdatedAt:   time.Unix(107800111, 0),
			},
			&acme_accounts.Account{
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
			&acme_accounts.Account{
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
			acme_accounts.UpdatePayload{
				ID:        16,
				Name:      new("le_production_account"),
				UpdatedAt: time.Unix(107800777, 0),
			},
			nil,
			helpers_test.NewTestErrorStringComp("UNIQUE constraint failed"),
			16,
			&acmeAcct16,
			nil,
		},
		// more tests for acme account updates
		{ // update all acme acct fields
			acme_accounts.UpdatePayload{
				ID:        29,
				Status:    new("deactivated"),
				Email:     new("newfake3@example.com"),
				KID:       new("https://fake.example.com/account/new-kid-1234"),
				CreatedAt: new(time.Unix(80808080, 0)),
				UpdatedAt: time.Unix(107800000, 0),
			},
			&acme_accounts.Account{
				ID:          29,
				Name:        "GC3",
				Description: "",
				AcmeServer:  acmeServer19,
				AccountKey:  key67,
				Status:      "deactivated",
				Email:       "newfake3@example.com",
				AcceptedTos: true,
				CreatedAt:   time.Unix(80808080, 0),
				UpdatedAt:   time.Unix(107800000, 0),
				Kid:         "https://fake.example.com/account/new-kid-1234",
			},
			nil,
			29,
			&acme_accounts.Account{
				ID:          29,
				Name:        "GC3",
				Description: "",
				AcmeServer:  acmeServer19,
				AccountKey:  key67,
				Status:      "deactivated",
				Email:       "newfake3@example.com",
				AcceptedTos: true,
				CreatedAt:   time.Unix(80808080, 0),
				UpdatedAt:   time.Unix(107800000, 0),
				Kid:         "https://fake.example.com/account/new-kid-1234",
			},
			nil,
		},
		{ // update just 1 acme acct fields
			acme_accounts.UpdatePayload{
				ID:        28,
				Status:    new("deactivated"),
				UpdatedAt: time.Unix(155500000, 0),
			},
			&acme_accounts.Account{
				ID:          28,
				Name:        "Google_Cloud_Staging2a",
				Description: "",
				AcmeServer:  acmeServer19,
				AccountKey:  key65,
				Status:      "deactivated",
				Email:       "fake2@example.com",
				AcceptedTos: true,
				CreatedAt:   time.Unix(1752418101, 0),
				UpdatedAt:   time.Unix(155500000, 0),
				Kid:         "https://dv.acme-v02.test-api.pki.goog/account/red-28",
			},
			nil,
			28,
			&acme_accounts.Account{
				ID:          28,
				Name:        "Google_Cloud_Staging2a",
				Description: "",
				AcmeServer:  acmeServer19,
				AccountKey:  key65,
				Status:      "deactivated",
				Email:       "fake2@example.com",
				AcceptedTos: true,
				CreatedAt:   time.Unix(1752418101, 0),
				UpdatedAt:   time.Unix(155500000, 0),
				Kid:         "https://dv.acme-v02.test-api.pki.goog/account/red-28",
			},
			nil,
		},
	}

	// create testing service
	store, err := openStorageWithTestData(t, "putacmeacctupdate")
	if err != nil {
		t.Fatal(err)
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("#%d (id: %d)", i, tc.payload.ID), func(t *testing.T) {
			acct, err := store.PutAcmeAccountUpdate(&tc.payload)
			if !helpers_test.ErrorsIs(err, tc.expectedPutErr) {
				t.Errorf("expected put error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedPutErr), helpers_test.ErrorToVal(err))
			}

			compareAcmeAccount(t, acct, tc.expectedPutResult)

			acct, err = store.GetOneAcmeAccountById(tc.getId)
			if !helpers_test.ErrorsIs(err, tc.expectedGetErr) {
				t.Errorf("expected get error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedGetErr), helpers_test.ErrorToVal(err))
			}

			compareAcmeAccount(t, acct, tc.expectedGetResult)
		})
	}
}

func TestPutAcmeAccountNewKey(t *testing.T) {
	testCases := []struct {
		payload           acme_accounts.RolloverKeyPayload
		expectedPutResult *acme_accounts.Account
		expectedPutErr    error
		getId             int
		expectedGetResult *acme_accounts.Account
		expectedGetErr    error
	}{
		{ // invalid
			acme_accounts.RolloverKeyPayload{
				ID: -1,
			},
			nil,
			storage.ErrWrongUpdateRowCount,
			-1,
			nil,
			sql.ErrNoRows,
		},
		{ // invalid
			acme_accounts.RolloverKeyPayload{
				ID: 522,
			},
			nil,
			storage.ErrWrongUpdateRowCount,
			522,
			nil,
			sql.ErrNoRows,
		},
		// ok updates
		{
			acme_accounts.RolloverKeyPayload{
				ID:           28,
				PrivateKeyID: new(58),
				UpdatedAt:    time.Unix(1088265758, 0),
			},
			&acme_accounts.Account{
				ID:          28,
				Name:        "Google_Cloud_Staging2a",
				Description: "",
				AcmeServer:  acmeServer19,
				AccountKey:  key58,
				Status:      "valid",
				Email:       "fake2@example.com",
				AcceptedTos: true,
				CreatedAt:   time.Unix(1752418101, 0),
				UpdatedAt:   time.Unix(1088265758, 0),
				Kid:         "https://dv.acme-v02.test-api.pki.goog/account/red-28",
			},
			nil,
			28,
			&acme_accounts.Account{
				ID:          28,
				Name:        "Google_Cloud_Staging2a",
				Description: "",
				AcmeServer:  acmeServer19,
				AccountKey:  key58,
				Status:      "valid",
				Email:       "fake2@example.com",
				AcceptedTos: true,
				CreatedAt:   time.Unix(1752418101, 0),
				UpdatedAt:   time.Unix(1088265758, 0),
				Kid:         "https://dv.acme-v02.test-api.pki.goog/account/red-28",
			},
			nil,
		},
		// bad updates
		{
			acme_accounts.RolloverKeyPayload{
				ID:           1,
				PrivateKeyID: new(-1), // invalid key (doesnt exist)
				UpdatedAt:    time.Unix(1088000751, 0),
			},
			nil,
			helpers_test.NewTestErrorStringComp("FOREIGN KEY constraint failed"),
			1,
			&acmeAcct1,
			nil,
		},
		{
			acme_accounts.RolloverKeyPayload{
				ID:           1,
				PrivateKeyID: new(678), // invalid key (doesnt exist)
				UpdatedAt:    time.Unix(1088000752, 0),
			},
			nil,
			helpers_test.NewTestErrorStringComp("FOREIGN KEY constraint failed"),
			1,
			&acmeAcct1,
			nil,
		},
		{
			acme_accounts.RolloverKeyPayload{
				ID:           23,
				PrivateKeyID: new(1), // invalid key (unique constraint fails)
				UpdatedAt:    time.Unix(1088000753, 0),
			},
			nil,
			helpers_test.NewTestErrorStringComp("UNIQUE constraint failed"),
			23,
			&acmeAcct23,
			nil,
		},
	}

	// create testing service
	store, err := openStorageWithTestData(t, "putacmeacctnewkey")
	if err != nil {
		t.Fatal(err)
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("#%d (id: %d)", i, tc.payload.ID), func(t *testing.T) {
			acct, err := store.PutAcmeAccountNewKey(&tc.payload)
			if !helpers_test.ErrorsIs(err, tc.expectedPutErr) {
				t.Errorf("expected put error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedPutErr), helpers_test.ErrorToVal(err))
			}

			compareAcmeAccount(t, acct, tc.expectedPutResult)

			acct, err = store.GetOneAcmeAccountById(tc.getId)
			if !helpers_test.ErrorsIs(err, tc.expectedGetErr) {
				t.Errorf("expected get error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedGetErr), helpers_test.ErrorToVal(err))
			}

			compareAcmeAccount(t, acct, tc.expectedGetResult)
		})
	}
}
