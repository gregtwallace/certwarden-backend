package storage_test

import (
	"certwarden-backend/pkg/domain/acme_servers"
	"certwarden-backend/pkg/pagination_sort"
	"certwarden-backend/pkg/test_helpers"
	"database/sql"
	"errors"
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
		expectedServerAtIndx acme_servers.Server
	}{
		{pagination_sort.Query{}, 5, 5, 3, acmeServer0},
		{QueryBuilderForTest(1, 1, "id", true), 5, 1, 0, acmeServer1},
		{QueryBuilderForTest(2, 1, "updated_at", false), 5, 2, 1, acmeServer4},
	}

	// create testing service
	storage, err := openStorageWithTestData(t, "getallacmeservers")
	if err != nil {
		t.Fatal(err)
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("#%d (%s)", i, tc.expectedServerAtIndx.Name), func(t *testing.T) {
			servers, totalCt, err := storage.GetAllAcmeServers(tc.q)
			if err != nil {
				t.Errorf("get all keys failed")
				return
			}

			if totalCt != tc.expectedTotalCt {
				t.Errorf("incorrect total count, expected '%d' but got '%d'", tc.expectedTotalCt, totalCt)
			}
			if len(servers) != tc.expectedResultLen {
				t.Errorf("incorrect servers length, expected '%d' but got '%d'", tc.expectedResultLen, len(servers))
			}
			if tc.testIndx <= len(servers)-1 {
				CompareAcmeServer(t, servers[tc.testIndx], tc.expectedServerAtIndx)
			} else {
				t.Errorf("couldnt test server at index '%d' because length of server array was only '%d'", tc.testIndx, len(servers))
			}
		})
	}
}

func TestGetOneServerById(t *testing.T) {
	testCases := []struct {
		id             int
		expectedErr    error
		expectedServer acme_servers.Server
	}{
		{-5, sql.ErrNoRows, acme_servers.Server{}},
		{50, sql.ErrNoRows, acme_servers.Server{}},
		{0, nil, acmeServer0},
		{1, nil, acmeServer1},
		{4, nil, acmeServer4},
	}

	// create testing service
	storage, err := openStorageWithTestData(t, "getoneserverbyid")
	if err != nil {
		t.Fatal(err)
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("#%d (id: %d)", i, tc.id), func(t *testing.T) {
			serv, err := storage.GetOneServerById(tc.id)
			if !errors.Is(err, tc.expectedErr) {
				t.Errorf("expected error '%s' but got '%s'", test_helpers.ErrorToVal(tc.expectedErr), test_helpers.ErrorToVal(err))
			}

			CompareAcmeServer(t, tc.expectedServer, serv)
		})
	}
}

func TestGetOneServerByName(t *testing.T) {
	testCases := []struct {
		name           string
		expectedErr    error
		expectedServer acme_servers.Server
	}{
		{"fake-bad-name", sql.ErrNoRows, acme_servers.Server{}},
		{"", sql.ErrNoRows, acme_servers.Server{}},
		{"Lets_Encrypt", nil, acmeServer0},
		{"Lets_Encrypt_Staging", nil, acmeServer1},
		{"Google_Prod", nil, acmeServer4},
	}

	// create testing service
	storage, err := openStorageWithTestData(t, "getoneserverbyname")
	if err != nil {
		t.Fatal(err)
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("#%d (name: %s)", i, tc.name), func(t *testing.T) {
			serv, err := storage.GetOneServerByName(tc.name)
			if !errors.Is(err, tc.expectedErr) {
				t.Errorf("expected error '%s' but got '%s'", test_helpers.ErrorToVal(tc.expectedErr), test_helpers.ErrorToVal(err))
			}

			CompareAcmeServer(t, tc.expectedServer, serv)
		})
	}
}
