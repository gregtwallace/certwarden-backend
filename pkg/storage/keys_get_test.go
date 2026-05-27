package storage_test

import (
	"certwarden-backend/pkg/domain/private_keys"
	"certwarden-backend/pkg/domain/private_keys/key_crypto"
	"certwarden-backend/pkg/pagination_sort"
	"certwarden-backend/pkg/test_helpers"
	"database/sql"
	"errors"
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
)

// TestGetAllKeys does spot checking of expected results
func TestGetAllKeys(t *testing.T) {
	// create testing service
	storage, err := openStorageWithTestData(t, "getallkeys")
	if err != nil {
		t.Fatal(err)
	}

	// no/default query
	keys, totalCt, err := storage.GetAllKeys(pagination_sort.Query{})
	if err != nil {
		t.Fatalf("get all keys failed")
	}
	if totalCt != 19 {
		t.Errorf("get all keys returned incorrect total count, expected '%d' but got '%d'", 19, totalCt)
	}
	if len(keys) != 19 {
		t.Errorf("get all keys returned incorrect keys length, expected '%d' but got '%d'", 19, len(keys))
	}
	if len(keys) > 0 {
		KeyCompare(t, keys[0], key63)
	}

	// alternate query
	keys, totalCt, err = storage.GetAllKeys(QueryBuilderForTest(5, 15, "algorithm", true))
	if err != nil {
		t.Fatalf("get all keys failed")
	}
	if totalCt != 19 {
		t.Errorf("get all keys returned incorrect total count, expected '%d' but got '%d'", 19, totalCt)
	}
	if len(keys) != 4 {
		t.Errorf("get all keys returned incorrect keys length, expected '%d' but got '%d'", 4, len(keys))
	}
	if len(keys) > 0 {
		KeyCompare(t, keys[2], key67)
	}

	// alternate query 2
	keys, totalCt, err = storage.GetAllKeys(QueryBuilderForTest(10, 0, "last_access", false))
	if err != nil {
		t.Fatalf("get all keys failed")
	}
	if totalCt != 19 {
		t.Errorf("get all keys returned incorrect total count, expected '%d' but got '%d'", 19, totalCt)
	}
	if len(keys) != 10 {
		t.Errorf("get all keys returned incorrect keys length, expected '%d' but got '%d'", 10, len(keys))
	}
	if len(keys) > 0 {
		KeyCompare(t, keys[1], key31)
	}
}

func TestGetOneKeyById(t *testing.T) {
	// create testing service
	storage, err := openStorageWithTestData(t, "getonekeybyid")
	if err != nil {
		t.Fatal(err)
	}

	// tests
	// doesnt exist
	_, err = storage.GetOneKeyById(22)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("expected error '%s' but got '%s'", sql.ErrNoRows, test_helpers.ErrorToVal(err))
	}

	// a key
	key, err := storage.GetOneKeyById(31)
	if err != nil {
		t.Errorf("expected key but received err '%s'", err)
	}

	KeyCompare(t, key, key31)

	// another key
	key, err = storage.GetOneKeyById(67)
	if err != nil {
		t.Errorf("expected key but received err '%s'", err)
	}

	KeyCompare(t, key, key67)
}

func TestGetOneKeyByName(t *testing.T) {
	// create testing service
	storage, err := openStorageWithTestData(t, "getonekeybyname")
	if err != nil {
		t.Fatal(err)
	}

	// tests
	// doesnt exist
	_, err = storage.GetOneKeyByName("fake-bad-name")
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("expected error '%s' but got '%s'", sql.ErrNoRows, test_helpers.ErrorToVal(err))
	}

	// a key
	key, err := storage.GetOneKeyByName("certwarden")
	if err != nil {
		t.Errorf("expected key but received err '%s'", err)
	}

	KeyCompare(t, key, key31)

	// another key
	key, err = storage.GetOneKeyByName("_Another_Test_Acct_LE_Staging")
	if err != nil {
		t.Errorf("expected key but received err '%s'", err)
	}

	KeyCompare(t, key, key63)
}

// TODO: TestGetAvailableKeys
