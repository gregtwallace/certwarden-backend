package storage_test

import (
	"certwarden-backend/pkg/domain/acme_servers"
	"certwarden-backend/pkg/helpers_test"
	"certwarden-backend/pkg/pagination_sort"
	"database/sql"
	"fmt"
	"testing"
	"time"
)

var (
	acmeServer0 = acme_servers.Server{
		ID:           0,
		Name:         "Lets_Encrypt",
		Description:  "Let's Encrypt Production Server",
		DirectoryURL: "https://acme-v02.api.letsencrypt.org/directory",
		IsStaging:    false,
		CreatedAt:    time.Unix(1697139774, 0),
		UpdatedAt:    time.Unix(1697139774, 0),
	}

	acmeServer1 = acme_servers.Server{
		ID:           1,
		Name:         "Lets_Encrypt_Staging",
		Description:  "Let's Encrypt Staging Server",
		DirectoryURL: "https://acme-staging-v02.api.letsencrypt.org/directory",
		IsStaging:    true,
		CreatedAt:    time.Unix(1697139774, 0),
		UpdatedAt:    time.Unix(1697139800, 0),
	}

	acmeServer4 = acme_servers.Server{
		ID:           4,
		Name:         "Google_Prod",
		Description:  "",
		DirectoryURL: "https://dv.acme-v02.api.pki.goog/directory",
		IsStaging:    false,
		CreatedAt:    time.Unix(1699565933, 0),
		UpdatedAt:    time.Unix(1699565933, 0),
	}

	acmeServer19 = acme_servers.Server{
		ID:           19,
		Name:         "Google_Cloud_Staging",
		Description:  "Google Cloud PreProd",
		DirectoryURL: "https://dv.acme-v02.test-api.pki.goog/directory",
		IsStaging:    true,
		CreatedAt:    time.Unix(1745080146, 0),
		UpdatedAt:    time.Unix(1745080146, 0),
	}
)

func TestGetAllAcmeServers(t *testing.T) {
	testCases := []struct {
		q                    pagination_sort.Query
		expectedTotalCt      int
		expectedResultLen    int
		testIndx             int
		expectedServerAtIndx *acme_servers.Server
	}{
		{pagination_sort.Query{}, 5, 5, 3, &acmeServer0},
		{queryBuilderForTest(1, 1, "id", true), 5, 1, 0, &acmeServer1},
		{queryBuilderForTest(2, 1, "updated_at", false), 5, 2, 1, &acmeServer4},
	}

	// create testing service
	store, err := openStorageWithTestData(t, "getallacmeservers")
	if err != nil {
		t.Fatal(err)
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("#%d (%s)", i, tc.expectedServerAtIndx.Name), func(t *testing.T) {
			servers, totalCt, err := store.GetAllAcmeServers(tc.q)
			if err != nil {
				t.Errorf("get all failed")
				return
			}

			if totalCt != tc.expectedTotalCt {
				t.Errorf("incorrect total count, expected '%d' but got '%d'", tc.expectedTotalCt, totalCt)
			}
			if len(servers) != tc.expectedResultLen {
				t.Errorf("incorrect result length, expected '%d' but got '%d'", tc.expectedResultLen, len(servers))
			}
			if tc.testIndx <= len(servers)-1 {
				compareAcmeServer(t, servers[tc.testIndx], tc.expectedServerAtIndx)
			} else {
				t.Errorf("couldnt test result at index '%d' because length of result array was only '%d'", tc.testIndx, len(servers))
			}
		})
	}
}

func TestGetOneServerById(t *testing.T) {
	testCases := []struct {
		id             int
		expectedErr    error
		expectedServer *acme_servers.Server
	}{
		{-5, sql.ErrNoRows, nil},
		{50, sql.ErrNoRows, nil},
		{0, nil, &acmeServer0},
		{1, nil, &acmeServer1},
		{4, nil, &acmeServer4},
	}

	// create testing service
	store, err := openStorageWithTestData(t, "getoneserverbyid")
	if err != nil {
		t.Fatal(err)
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("#%d (id: %d)", i, tc.id), func(t *testing.T) {
			serv, err := store.GetOneServerById(tc.id)
			if !helpers_test.ErrorsIs(err, tc.expectedErr) {
				t.Errorf("expected error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedErr), helpers_test.ErrorToVal(err))
			}

			compareAcmeServer(t, serv, tc.expectedServer)
		})
	}
}

func TestGetOneServerByName(t *testing.T) {
	testCases := []struct {
		name           string
		expectedErr    error
		expectedServer *acme_servers.Server
	}{
		{"fake-bad-name", sql.ErrNoRows, nil},
		{"", sql.ErrNoRows, nil},
		{"lets_encrypt", nil, &acmeServer0}, // case is wrong
		{"Lets_Encrypt_Staging", nil, &acmeServer1},
		{"Google_Prod", nil, &acmeServer4},
	}

	// create testing service
	store, err := openStorageWithTestData(t, "getoneserverbyname")
	if err != nil {
		t.Fatal(err)
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("#%d (name: %s)", i, tc.name), func(t *testing.T) {
			serv, err := store.GetOneServerByName(tc.name)
			if !helpers_test.ErrorsIs(err, tc.expectedErr) {
				t.Errorf("expected error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedErr), helpers_test.ErrorToVal(err))
			}

			compareAcmeServer(t, serv, tc.expectedServer)
		})
	}
}
