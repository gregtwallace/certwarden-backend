package storage_test

import (
	"certwarden-backend/pkg/domain/private_keys"
	"certwarden-backend/pkg/domain/private_keys/key_crypto"
	"certwarden-backend/pkg/helpers_test"
	"certwarden-backend/pkg/pagination_sort"
	"database/sql"
	"fmt"
	"testing"
	"time"
)

// keys for comparisons
var (
	key1 = private_keys.Key{
		ID:          1,
		Name:        "LE_Staging",
		Description: "",
		Algorithm:   key_crypto.AlgorithmECDSAp384,
		Pem: `-----BEGIN EC PRIVATE KEY-----
red-1
-----END EC PRIVATE KEY-----
`,
		ApiKey:         "key-api-key-1",
		ApiKeyNew:      "",
		ApiKeyDisabled: true,
		ApiKeyViaUrl:   false,
		LastAccess:     nil,
		CreatedAt:      time.Unix(1697142029, 0),
		UpdatedAt:      time.Unix(1697142029, 0),
	}

	key4 = private_keys.Key{
		ID:          4,
		Name:        "LE_Production",
		Description: "some desc goes here",
		Algorithm:   key_crypto.AlgorithmECDSAp384,
		Pem: `-----BEGIN EC PRIVATE KEY-----
red-4
-----END EC PRIVATE KEY-----
`,
		ApiKey:         "key-api-key-4",
		ApiKeyNew:      "new-key-4-here",
		ApiKeyDisabled: false,
		ApiKeyViaUrl:   true,
		LastAccess:     new(time.Unix(12345678, 0)),
		CreatedAt:      time.Unix(1697142945, 0),
		UpdatedAt:      time.Unix(1700593381, 0),
	}

	key31 = private_keys.Key{
		ID:          31,
		Name:        "certwarden",
		Description: "localhost / dev work w/ real cert",
		Algorithm:   key_crypto.AlgorithmECDSAp256,
		Pem: `-----BEGIN EC PRIVATE KEY-----
red-31
-----END EC PRIVATE KEY-----
`,
		ApiKey:         "key-api-key-31",
		ApiKeyNew:      "key-api-new-key-31",
		ApiKeyDisabled: false,
		ApiKeyViaUrl:   true,
		LastAccess:     new(time.Unix(1745952074, 0)),
		CreatedAt:      time.Unix(1709327549, 0),
		UpdatedAt:      time.Unix(1732748628, 0),
	}

	key55 = private_keys.Key{
		ID:          55,
		Name:        "test008.test.example.com",
		Description: "",
		Algorithm:   key_crypto.AlgorithmECDSAp256,
		Pem: `-----BEGIN EC PRIVATE KEY-----
red-55
-----END EC PRIVATE KEY-----
`,
		ApiKey:         "key-api-key-55",
		ApiKeyNew:      "",
		ApiKeyDisabled: false,
		ApiKeyViaUrl:   false,
		LastAccess:     nil,
		CreatedAt:      time.Unix(1743170701, 0),
		UpdatedAt:      time.Unix(0, 0),
	}

	key56 = private_keys.Key{
		ID:          56,
		Name:        "test008.test.example.com-p",
		Description: "",
		Algorithm:   key_crypto.AlgorithmECDSAp256,
		Pem: `-----BEGIN EC PRIVATE KEY-----
red-56
-----END EC PRIVATE KEY-----
`,
		ApiKey:         "key-api-key-56",
		ApiKeyNew:      "",
		ApiKeyDisabled: false,
		ApiKeyViaUrl:   false,
		LastAccess:     nil,
		CreatedAt:      time.Unix(1743171060, 0),
		UpdatedAt:      time.Unix(0, 0),
	}

	key57 = private_keys.Key{
		ID:          57,
		Name:        "a0.alias.test.example.com",
		Description: "",
		Algorithm:   key_crypto.AlgorithmECDSAp256,
		Pem: `-----BEGIN EC PRIVATE KEY-----
red-57
-----END EC PRIVATE KEY-----
`,
		ApiKey:         "key-api-key-57",
		ApiKeyNew:      "",
		ApiKeyDisabled: false,
		ApiKeyViaUrl:   false,
		LastAccess:     nil,
		CreatedAt:      time.Unix(1743173262, 0),
		UpdatedAt:      time.Unix(0, 0),
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
		LastAccess:     nil,
		CreatedAt:      time.Unix(1743176647, 0),
		UpdatedAt:      time.Unix(1751563349, 0),
	}

	key61 = private_keys.Key{
		ID:          61,
		Name:        "_Google_Cloud_Staging_Acct",
		Description: "",
		Algorithm:   key_crypto.AlgorithmECDSAp384,
		Pem: `-----BEGIN EC PRIVATE KEY-----
red-61
-----END EC PRIVATE KEY-----
`,
		ApiKey:         "key-api-key-61",
		ApiKeyNew:      "key-api-new-key-61",
		ApiKeyDisabled: true,
		ApiKeyViaUrl:   false,
		LastAccess:     nil,
		CreatedAt:      time.Unix(1751561584, 0),
		UpdatedAt:      time.Unix(1751561584, 0),
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
		LastAccess:     new(time.Unix(1777555691, 0)),
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
		LastAccess:     nil,
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
		LastAccess:     nil,
		CreatedAt:      time.Unix(1752258426, 0),
		UpdatedAt:      time.Unix(1752258426, 0),
	}

	key65 = private_keys.Key{
		ID:          65,
		Name:        "_Google_Cloud_Staging2",
		Description: "",
		Algorithm:   key_crypto.AlgorithmRSA4096,
		Pem: `-----BEGIN RSA PRIVATE KEY-----
red-65
-----END RSA PRIVATE KEY-----
`,
		ApiKey:         "key-api-key-65",
		ApiKeyNew:      "",
		ApiKeyDisabled: false,
		ApiKeyViaUrl:   false,
		LastAccess:     nil,
		CreatedAt:      time.Unix(1752262145, 0),
		UpdatedAt:      time.Unix(1752262145, 0),
	}

	key66 = private_keys.Key{
		ID:          66,
		Name:        "_Google_Cloud_Staging2b",
		Description: "",
		Algorithm:   key_crypto.AlgorithmECDSAp256,
		Pem: `-----BEGIN EC PRIVATE KEY-----
red-66
-----END EC PRIVATE KEY-----
`,
		ApiKey:         "key-api-key-66",
		ApiKeyNew:      "",
		ApiKeyDisabled: true,
		ApiKeyViaUrl:   false,
		LastAccess:     nil,
		CreatedAt:      time.Unix(1752266414, 0),
		UpdatedAt:      time.Unix(1752266418, 0),
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
		LastAccess:     nil,
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
		LastAccess:     new(time.Unix(1777555692, 0)),
		CreatedAt:      time.Unix(1775761592, 0),
		UpdatedAt:      time.Unix(0, 0),
	}

	key70 = private_keys.Key{
		ID:          70,
		Name:        "STAGING_persist--test005.test.example2.com",
		Description: "",
		Algorithm:   key_crypto.AlgorithmECDSAp256,
		Pem: `-----BEGIN EC PRIVATE KEY-----
red-70
-----END EC PRIVATE KEY-----
`,
		ApiKey:         "key-api-key-70",
		ApiKeyNew:      "",
		ApiKeyDisabled: false,
		ApiKeyViaUrl:   false,
		LastAccess:     nil,
		CreatedAt:      time.Unix(1779386740, 0),
		UpdatedAt:      time.Unix(0, 0),
	}
)

// TestGetAllKeys does spot checking of expected results
func TestGetAllKeys(t *testing.T) {
	testCases := []struct {
		q                 pagination_sort.Query
		expectedTotalCt   int
		expectedResultLen int
		testIndx          int
		expectedKeyAtIndx *private_keys.Key
	}{
		{pagination_sort.Query{}, 19, 19, 0, &key63},
		{queryBuilderForTest(5, 15, "algorithm", false), 19, 4, 2, &key67},
		{queryBuilderForTest(10, 0, "last_access", true), 19, 10, 2, &key31},
	}

	// create testing service
	store := openStorageWithTestData(t, "getallkeys")

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("#%d (%s)", i, tc.expectedKeyAtIndx.Name), func(t *testing.T) {
			keys, totalCt, err := store.GetAllKeys(tc.q)
			if err != nil {
				t.Errorf("get all failed")
				return
			}

			if totalCt != tc.expectedTotalCt {
				t.Errorf("incorrect total count, expected '%d' but got '%d'", tc.expectedTotalCt, totalCt)
			}
			if len(keys) != tc.expectedResultLen {
				t.Errorf("incorrect result length, expected '%d' but got '%d'", tc.expectedResultLen, len(keys))
			}
			if tc.testIndx <= len(keys)-1 {
				compareKey(t, keys[tc.testIndx], tc.expectedKeyAtIndx)
			} else {
				t.Errorf("couldnt test result at index '%d' because length of result array was only '%d'", tc.testIndx, len(keys))
			}
		})
	}
}

func TestGetOneKeyById(t *testing.T) {
	testCases := []struct {
		id          int
		expectedErr error
		expectedKey *private_keys.Key
	}{
		{-1, sql.ErrNoRows, nil},
		{22, sql.ErrNoRows, nil},
		{31, nil, &key31},
		{67, nil, &key67},
	}

	// create testing service
	store := openStorageWithTestData(t, "getonekeybyid")

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("#%d (id: %d)", i, tc.id), func(t *testing.T) {
			key, err := store.GetOneKeyById(tc.id)
			if !helpers_test.ErrorsIs(err, tc.expectedErr) {
				t.Errorf("expected error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedErr), helpers_test.ErrorToVal(err))
			}

			compareKey(t, key, tc.expectedKey)
		})
	}
}

func TestGetOneKeyByName(t *testing.T) {
	testCases := []struct {
		name        string
		expectedErr error
		expectedKey *private_keys.Key
	}{
		{"", sql.ErrNoRows, nil},
		{"fake-bad-name", sql.ErrNoRows, nil},
		{"cerTWarden", nil, &key31},
		{"_Another_TEST_Acct_le_Staging", nil, &key63}, // case is wrong
	}

	// create testing service
	store := openStorageWithTestData(t, "getonekeybyname")

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("#%d (name: %s)", i, tc.name), func(t *testing.T) {
			key, err := store.GetOneKeyByName(tc.name)
			if !helpers_test.ErrorsIs(err, tc.expectedErr) {
				t.Errorf("expected error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedErr), helpers_test.ErrorToVal(err))
			}

			compareKey(t, key, tc.expectedKey)
		})
	}
}

func TestGetAvailableKeys(t *testing.T) {
	// create testing service
	store := openStorageWithTestData(t, "getavailablekeys")

	keys, err := store.GetAvailableKeys()
	if err != nil {
		t.Errorf("get all failed")
		return
	}

	expectedResultLen := 3
	if len(keys) != expectedResultLen {
		t.Errorf("returned incorrect result length, expected '%d' but got '%d'", expectedResultLen, len(keys))
	}

	expectedKeys := []private_keys.Key{key58, key62, key69}
	for i, expectedKey := range expectedKeys {
		if i > len(keys)-1 {
			t.Errorf("expected id '%d' at index '%d' but result was too short", expectedKey.ID, i)
			continue
		}

		compareKey(t, keys[i], &expectedKey)
	}
}
