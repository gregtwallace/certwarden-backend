package storage_test

import (
	"certwarden-backend/pkg/domain/acme_accounts"
	"certwarden-backend/pkg/helpers_test"
	"certwarden-backend/pkg/pagination_sort"
	"database/sql"
	"fmt"
	"testing"
	"time"
)

var (
	acmeAcct1 = acme_accounts.Account{
		ID:          1,
		Name:        "LE_Staging_Account",
		Description: "",
		AcmeServer:  acmeServer1,
		AccountKey:  key1,
		Status:      "valid",
		Email:       "",
		AcceptedTos: true,
		CreatedAt:   time.Unix(1697142144, 0),
		UpdatedAt:   time.Unix(1752265759, 0),
		Kid:         "https://acme-staging-v02.api.letsencrypt.org/acme/acct/red-1",
	}

	acmeAcct2 = acme_accounts.Account{
		ID:          2,
		Name:        "LE_Production_Account",
		Description: "LE Prod Account - Primary",
		AcmeServer:  acmeServer0,
		AccountKey:  key4,
		Status:      "valid",
		Email:       "fake@example.com",
		AcceptedTos: true,
		CreatedAt:   time.Unix(1697142971, 0),
		UpdatedAt:   time.Unix(1746732527, 0),
		Kid:         "https://acme-v02.api.letsencrypt.org/acme/acct/red-2",
	}

	acmeAcct16 = acme_accounts.Account{
		ID:          16,
		Name:        "Google_Cloud_Staging",
		Description: "",
		AcmeServer:  acmeServer19,
		AccountKey:  key61,
		Status:      "valid",
		Email:       "fake@example.com",
		AcceptedTos: true,
		CreatedAt:   time.Unix(1751561598, 0),
		UpdatedAt:   time.Unix(1752253957, 0),
		Kid:         "https://dv.acme-v02.test-api.pki.goog/account/red-16",
	}

	acmeAcct20 = acme_accounts.Account{
		ID:          20,
		Name:        "_LE_Staging_Again",
		Description: "",
		AcmeServer:  acmeServer1,
		AccountKey:  key63,
		Status:      "valid",
		Email:       "",
		AcceptedTos: true,
		CreatedAt:   time.Unix(1752254932, 0),
		UpdatedAt:   time.Unix(1779474005, 0),
		Kid:         "https://acme-staging-v02.api.letsencrypt.org/acme/acct/red-20",
	}

	acmeAcct23 = acme_accounts.Account{
		ID:          23,
		Name:        "Google_Cloud_Staging2",
		Description: "",
		AcmeServer:  acmeServer19,
		AccountKey:  key66,
		Status:      "deactivated",
		Email:       "fake2@example.com",
		AcceptedTos: false,
		CreatedAt:   time.Unix(1752416890, 0),
		UpdatedAt:   time.Unix(1752416892, 0),
		Kid:         "https://dv.acme-v02.test-api.pki.goog/account/red-23",
	}
)

// TODO
func TestGetAllAcmeAccounts(t *testing.T) {
	testCases := []struct {
		q                  pagination_sort.Query
		expectedTotalCt    int
		expectedResultLen  int
		testIndx           int
		expectedAcctAtIndx acme_accounts.Account
	}{
		{pagination_sort.Query{}, 7, 7, 3, acmeAcct23},
		{QueryBuilderForTest(1, 1, "id", true), 7, 1, 0, acmeAcct2},
		{QueryBuilderForTest(2, 1, "servername", false), 7, 2, 1, acmeAcct2},
		{QueryBuilderForTest(2, 4, "servername", false), 7, 2, 0, acmeAcct23},
		{QueryBuilderForTest(6, 1, "keyname", true), 7, 6, 5, acmeAcct1},
	}

	// create testing service
	storage, err := openStorageWithTestData(t, "getallacmeaccounts")
	if err != nil {
		t.Fatal(err)
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("#%d (%s)", i, tc.expectedAcctAtIndx.Name), func(t *testing.T) {
			accts, totalCt, err := storage.GetAllAcmeAccounts(tc.q)
			if err != nil {
				t.Errorf("get all failed")
				return
			}

			if totalCt != tc.expectedTotalCt {
				t.Errorf("incorrect total count, expected '%d' but got '%d'", tc.expectedTotalCt, totalCt)
			}
			if len(accts) != tc.expectedResultLen {
				t.Errorf("incorrect result length, expected '%d' but got '%d'", tc.expectedResultLen, len(accts))
			}
			if tc.testIndx <= len(accts)-1 {
				CompareAcmeAccount(t, accts[tc.testIndx], tc.expectedAcctAtIndx)
			} else {
				t.Errorf("couldnt test result at index '%d' because length of result array was only '%d'", tc.testIndx, len(accts))
			}
		})
	}
}

func TestGetOneAccountById(t *testing.T) {
	testCases := []struct {
		id           int
		expectedErr  error
		expectedAcct acme_accounts.Account
	}{
		{-5, sql.ErrNoRows, acme_accounts.Account{}},
		{50, sql.ErrNoRows, acme_accounts.Account{}},
		{1, nil, acmeAcct1},
		{2, nil, acmeAcct2},
		{23, nil, acmeAcct23},
	}

	// create testing service
	storage, err := openStorageWithTestData(t, "getoneaccountbyid")
	if err != nil {
		t.Fatal(err)
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("#%d (id: %d)", i, tc.id), func(t *testing.T) {
			acct, err := storage.GetOneAcmeAccountById(tc.id)
			if !helpers_test.ErrorsIs(err, tc.expectedErr) {
				t.Errorf("expected error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedErr), helpers_test.ErrorToVal(err))
			}

			CompareAcmeAccount(t, acct, tc.expectedAcct)
		})
	}
}

func TestGetOneAccountByName(t *testing.T) {
	testCases := []struct {
		name         string
		expectedErr  error
		expectedAcct acme_accounts.Account
	}{
		{"", sql.ErrNoRows, acme_accounts.Account{}},
		{"fake-name", sql.ErrNoRows, acme_accounts.Account{}},
		{"le_staging_account", nil, acmeAcct1}, // case is wrong
		{"LE_Production_Account", nil, acmeAcct2},
		{"Google_Cloud_Staging2", nil, acmeAcct23},
	}

	// create testing service
	storage, err := openStorageWithTestData(t, "getoneaccountbyname")
	if err != nil {
		t.Fatal(err)
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("#%d (name: %s)", i, tc.name), func(t *testing.T) {
			acct, err := storage.GetOneAcmeAccountByName(tc.name)
			if !helpers_test.ErrorsIs(err, tc.expectedErr) {
				t.Errorf("expected error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedErr), helpers_test.ErrorToVal(err))
			}

			CompareAcmeAccount(t, acct, tc.expectedAcct)
		})
	}
}
