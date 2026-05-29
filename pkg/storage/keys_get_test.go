package storage_test

import (
	"certwarden-backend/pkg/domain/private_keys"
	"certwarden-backend/pkg/domain/private_keys/key_crypto"
	"certwarden-backend/pkg/pagination_sort"
	"certwarden-backend/pkg/test_helpers"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"
)

// keys for comparisons
var (
	key31 = private_keys.Key{
		ID:          31,
		Name:        "certwarden",
		Description: "localhost / dev work w/ real cert",
		Algorithm:   key_crypto.AlgorithmECDSAp256,
		Pem: `-----BEGIN EC PRIVATE KEY-----
red-31
-----END EC PRIVATE KEY-----
`,
		ApiKey:         "key-api-key-4",
		ApiKeyNew:      "key-api-new-key-4",
		ApiKeyDisabled: false,
		ApiKeyViaUrl:   true,
		LastAccess:     time.Unix(1745952074, 0),
		CreatedAt:      time.Unix(1709327549, 0),
		UpdatedAt:      time.Unix(1732748628, 0),
	}

	key58 = private_keys.Key{
		ID:          58,
		Name:        "_Buypass_Staging",
		Description: "",
		Algorithm:   key_crypto.AlgorithmECDSAp256,
		Pem: `-----BEGIN EC PRIVATE KEY-----
red-58
-----END EC PRIVATE KEY-----
`,
		ApiKey:         "key-api-key-58",
		ApiKeyNew:      "",
		ApiKeyDisabled: true,
		ApiKeyViaUrl:   false,
		LastAccess:     time.Unix(0, 0),
		CreatedAt:      time.Unix(1743176647, 0),
		UpdatedAt:      time.Unix(1751563349, 0),
	}

	key62 = private_keys.Key{
		ID:          62,
		Name:        "SomeKEy",
		Description: "some desc",
		Algorithm:   key_crypto.AlgorithmECDSAp256,
		Pem: `-----BEGIN EC PRIVATE KEY-----
red-62
-----END EC PRIVATE KEY-----
`,
		ApiKey:         "key-api-key-62",
		ApiKeyNew:      "key-api-new-key-62",
		ApiKeyDisabled: false,
		ApiKeyViaUrl:   false,
		LastAccess:     time.Unix(1777555691, 0),
		CreatedAt:      time.Unix(1751738296, 0),
		UpdatedAt:      time.Unix(1751738296, 0),
	}

	key63 = private_keys.Key{
		ID:          63,
		Name:        "_Another_Test_Acct_LE_Staging",
		Description: "",
		Algorithm:   key_crypto.AlgorithmRSA2048,
		Pem: `-----BEGIN RSA PRIVATE KEY-----
red-63
-----END RSA PRIVATE KEY-----
`,
		ApiKey:         "key-api-key-63",
		ApiKeyNew:      "",
		ApiKeyDisabled: true,
		ApiKeyViaUrl:   false,
		LastAccess:     time.Unix(0, 0),
		CreatedAt:      time.Unix(1752254918, 0),
		UpdatedAt:      time.Unix(1752254918, 0),
	}

	key64 = private_keys.Key{
		ID:          64,
		Name:        "_Another_Test_Acct_LE_Staging_Roll",
		Description: "",
		Algorithm:   key_crypto.AlgorithmECDSAp256,
		Pem: `-----BEGIN EC PRIVATE KEY-----
red-64
-----END EC PRIVATE KEY-----
`,
		ApiKey:         "key-api-key-64",
		ApiKeyNew:      "",
		ApiKeyDisabled: true,
		ApiKeyViaUrl:   false,
		LastAccess:     time.Unix(0, 0),
		CreatedAt:      time.Unix(1752258426, 0),
		UpdatedAt:      time.Unix(1752258426, 0),
	}

	key67 = private_keys.Key{
		ID:          67,
		Name:        "_GC3",
		Description: "",
		Algorithm:   key_crypto.AlgorithmRSA3072,
		Pem: `-----BEGIN RSA PRIVATE KEY-----
red-67
-----END RSA PRIVATE KEY-----
`,
		ApiKey:         "key-api-key-67",
		ApiKeyNew:      "",
		ApiKeyDisabled: true,
		ApiKeyViaUrl:   false,
		LastAccess:     time.Unix(0, 0),
		CreatedAt:      time.Unix(1752418131, 0),
		UpdatedAt:      time.Unix(1752418131, 0),
	}

	key69 = private_keys.Key{
		ID:          69,
		Name:        "STAGING_persist--test007.test.example2.com",
		Description: "",
		Algorithm:   key_crypto.AlgorithmECDSAp256,
		Pem: `-----BEGIN EC PRIVATE KEY-----
red-69
-----END EC PRIVATE KEY-----
`,
		ApiKey:         "key-api-key-69",
		ApiKeyNew:      "",
		ApiKeyDisabled: false,
		ApiKeyViaUrl:   false,
		LastAccess:     time.Unix(1777555692, 0),
		CreatedAt:      time.Unix(1775761592, 0),
		// TODO: Research why this is 0 in the db. Might find while writing other tests. Also possible this is just junk data from a previous bug.
		UpdatedAt: time.Unix(0, 0),
	}
)

// TestGetAllKeys does spot checking of expected results
func TestGetAllKeys(t *testing.T) {
	testCases := []struct {
		q                 pagination_sort.Query
		expectedTotalCt   int
		expectedResultLen int
		testIndx          int
		expectedKeyAtIndx private_keys.Key
	}{
		{pagination_sort.Query{}, 19, 19, 0, key63},
		{QueryBuilderForTest(5, 15, "algorithm", true), 19, 4, 2, key67},
		{QueryBuilderForTest(10, 0, "last_access", false), 19, 10, 2, key31},
	}

	// create testing service
	storage, err := openStorageWithTestData(t, "getallkeys")
	if err != nil {
		t.Fatal(err)
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("#%d (%s)", i, tc.expectedKeyAtIndx.Name), func(t *testing.T) {
			keys, totalCt, err := storage.GetAllKeys(tc.q)
			if err != nil {
				t.Errorf("get all keys failed")
				return
			}

			if totalCt != tc.expectedTotalCt {
				t.Errorf("get all keys returned incorrect total count, expected '%d' but got '%d'", tc.expectedTotalCt, totalCt)
			}
			if len(keys) != tc.expectedResultLen {
				t.Errorf("get all keys returned incorrect keys length, expected '%d' but got '%d'", tc.expectedResultLen, len(keys))
			}
			if tc.testIndx <= len(keys)-1 {
				KeyCompare(t, keys[tc.testIndx], tc.expectedKeyAtIndx)
			} else {
				t.Errorf("couldnt test key at index '%d' because length of key array was only '%d'", tc.testIndx, len(keys))
			}
		})
	}
}

func TestGetOneKeyById(t *testing.T) {
	testCases := []struct {
		id          int
		expectedErr error
		expectedKey private_keys.Key
	}{
		{22, sql.ErrNoRows, private_keys.Key{}},
		{31, nil, key31},
		{67, nil, key67},
	}

	// create testing service
	storage, err := openStorageWithTestData(t, "getonekeybyid")
	if err != nil {
		t.Fatal(err)
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("#%d (id: %d)", i, tc.id), func(t *testing.T) {
			key, err := storage.GetOneKeyById(tc.id)
			if !errors.Is(err, tc.expectedErr) {
				t.Errorf("expected error '%s' but got '%s'", tc.expectedErr, test_helpers.ErrorToVal(err))
			}

			KeyCompare(t, tc.expectedKey, key)
		})
	}
}

func TestGetOneKeyByName(t *testing.T) {
	testCases := []struct {
		name        string
		expectedErr error
		expectedKey private_keys.Key
	}{
		{"fake-bad-name", sql.ErrNoRows, private_keys.Key{}},
		{"certwarden", nil, key31},
		{"_Another_Test_Acct_LE_Staging", nil, key63},
	}

	// create testing service
	storage, err := openStorageWithTestData(t, "getonekeybyname")
	if err != nil {
		t.Fatal(err)
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("#%d (name: %s)", i, tc.name), func(t *testing.T) {
			key, err := storage.GetOneKeyByName(tc.name)
			if !errors.Is(err, tc.expectedErr) {
				t.Errorf("expected error '%s' but got '%s'", tc.expectedErr, test_helpers.ErrorToVal(err))
			}

			KeyCompare(t, tc.expectedKey, key)
		})
	}
}

func TestGetAvailableKeys(t *testing.T) {
	// create testing service
	storage, err := openStorageWithTestData(t, "getavailablekeys")
	if err != nil {
		t.Fatal(err)
	}

	keys, err := storage.GetAvailableKeys()
	if err != nil {
		t.Errorf("get available keys failed")
		return
	}

	expectedResultLen := 3
	if len(keys) != expectedResultLen {
		t.Errorf("get available keys returned incorrect keys length, expected '%d' but got '%d'", expectedResultLen, len(keys))
	}

	expectedKeys := []private_keys.Key{key58, key62, key69}
	for i, expectedKey := range expectedKeys {
		if i > len(keys)-1 {
			t.Errorf("expected key id '%d' at index '%d' but result was too short", expectedKey.ID, i)
			continue
		}

		KeyCompare(t, keys[i], expectedKey)
	}
}
