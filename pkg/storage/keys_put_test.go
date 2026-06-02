package storage_test

import (
	"certwarden-backend/pkg/domain/private_keys"
	"certwarden-backend/pkg/domain/private_keys/key_crypto"
	"certwarden-backend/pkg/test_helpers"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestPutKeyUpdate(t *testing.T) {
	testCases := []struct {
		payload        private_keys.UpdatePayload
		expectedKey    private_keys.Key
		expectedPutErr error
		expectedGetErr error
	}{
		{ // invalid key
			private_keys.UpdatePayload{
				ID: -1,
			},
			private_keys.Key{},
			sql.ErrNoRows,
			sql.ErrNoRows,
		},
		{ // invalid key
			private_keys.UpdatePayload{
				ID: 522,
			},
			private_keys.Key{},
			sql.ErrNoRows,
			sql.ErrNoRows,
		},
		{ // update all things
			private_keys.UpdatePayload{
				ID:        31,
				UpdatedAt: 100555,
			},
			private_keys.Key{
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
				UpdatedAt:      time.Unix(100555, 0),
			},
			nil,
			nil,
		},
		{ // update none of the things (except last update)
			private_keys.UpdatePayload{
				ID:        62,
				UpdatedAt: 1001111,
			},
			private_keys.Key{
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
				UpdatedAt:      time.Unix(1001111, 0),
			},
			nil,
			nil,
		},
		{ // update just key disabled
			private_keys.UpdatePayload{
				ID:             58,
				ApiKeyDisabled: new(false),
				UpdatedAt:      1751730000,
			},
			private_keys.Key{
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
				ApiKeyDisabled: false,
				ApiKeyViaUrl:   false,
				LastAccess:     time.Unix(0, 0),
				CreatedAt:      time.Unix(1743176647, 0),
				UpdatedAt:      time.Unix(1751730000, 0),
			},
			nil,
			nil,
		},
	}

	// create testing service
	storage, err := openStorageWithTestData(t, "putkeyupdate")
	if err != nil {
		t.Fatal(err)
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("#%d (id: %d)", i, tc.payload.ID), func(t *testing.T) {
			key, err := storage.PutKeyUpdate(tc.payload)
			if !errors.Is(err, tc.expectedPutErr) {
				t.Errorf("expected put error '%s' but got '%s'", test_helpers.ErrorToVal(tc.expectedPutErr), test_helpers.ErrorToVal(err))
			}

			KeyCompare(t, tc.expectedKey, key)

			key, err = storage.GetOneKeyById(tc.payload.ID)
			if !errors.Is(err, tc.expectedGetErr) {
				t.Errorf("expected get error '%s' but got '%s'", test_helpers.ErrorToVal(tc.expectedGetErr), test_helpers.ErrorToVal(err))
			}

			KeyCompare(t, key, tc.expectedKey)
		})
	}
}

func TestPutKeyApiKey(t *testing.T) {
	testCases := []struct {
		keyId          int
		apiKey         string
		updateTimeUnix int

		expectedKey    private_keys.Key
		expectedPutErr error
		expectedGetErr error
	}{
		{ // invalid key id
			-1,
			"fake",
			100005,
			private_keys.Key{},
			sql.ErrNoRows,
			sql.ErrNoRows,
		},
		{ // invalid key id
			500,
			"anotherfake",
			10005000,
			private_keys.Key{},
			sql.ErrNoRows,
			sql.ErrNoRows,
		},
		// do some updates
		{
			31,
			"fake31",
			1022005,
			private_keys.Key{
				ID:          31,
				Name:        "certwarden",
				Description: "localhost / dev work w/ real cert",
				Algorithm:   key_crypto.AlgorithmECDSAp256,
				Pem: `-----BEGIN EC PRIVATE KEY-----
red-31
-----END EC PRIVATE KEY-----
`,
				ApiKey:         "fake31",
				ApiKeyNew:      "key-api-new-key-4",
				ApiKeyDisabled: false,
				ApiKeyViaUrl:   true,
				LastAccess:     time.Unix(1745952074, 0),
				CreatedAt:      time.Unix(1709327549, 0),
				UpdatedAt:      time.Unix(1022005, 0),
			},
			nil,
			nil,
		},
		{
			62,
			"62thing",
			0,
			private_keys.Key{
				ID:          62,
				Name:        "SomeKEy",
				Description: "some desc",
				Algorithm:   key_crypto.AlgorithmECDSAp256,
				Pem: `-----BEGIN EC PRIVATE KEY-----
red-62
-----END EC PRIVATE KEY-----
`,
				ApiKey:         "62thing",
				ApiKeyNew:      "key-api-new-key-62",
				ApiKeyDisabled: false,
				ApiKeyViaUrl:   false,
				LastAccess:     time.Unix(1777555691, 0),
				CreatedAt:      time.Unix(1751738296, 0),
				UpdatedAt:      time.Unix(0, 0),
			},
			nil,
			nil,
		},
	}

	// create testing service
	storage, err := openStorageWithTestData(t, "putkeyapikey")
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("id: %d)", tc.keyId), func(t *testing.T) {
			err := storage.PutKeyApiKey(tc.keyId, tc.apiKey, tc.updateTimeUnix)
			if !errors.Is(err, tc.expectedPutErr) {
				t.Errorf("expected put error '%s' but got '%s'", test_helpers.ErrorToVal(tc.expectedPutErr), test_helpers.ErrorToVal(err))
			}

			key, err := storage.GetOneKeyById(tc.keyId)
			if !errors.Is(err, tc.expectedGetErr) {
				t.Errorf("expected get error '%s' but got '%s'", test_helpers.ErrorToVal(tc.expectedGetErr), test_helpers.ErrorToVal(err))
			}

			KeyCompare(t, key, tc.expectedKey)
		})
	}
}

func TestPutKeyNewApiKey(t *testing.T) {
	testCases := []struct {
		keyId          int
		apiKeyNew      string
		updateTimeUnix int

		expectedKey    private_keys.Key
		expectedPutErr error
		expectedGetErr error
	}{
		{ // invalid key id
			-1,
			"",
			1099905,
			private_keys.Key{},
			sql.ErrNoRows,
			sql.ErrNoRows,
		},
		{ // invalid key id
			500,
			"anotherfake",
			10005000,
			private_keys.Key{},
			sql.ErrNoRows,
			sql.ErrNoRows,
		},
		// do some updates
		{
			69,
			"fakenew69",
			1022005,
			private_keys.Key{
				ID:          69,
				Name:        "STAGING_persist--test007.test.example2.com",
				Description: "",
				Algorithm:   key_crypto.AlgorithmECDSAp256,
				Pem: `-----BEGIN EC PRIVATE KEY-----
red-69
-----END EC PRIVATE KEY-----
`,
				ApiKey:         "key-api-key-69",
				ApiKeyNew:      "fakenew69",
				ApiKeyDisabled: false,
				ApiKeyViaUrl:   false,
				LastAccess:     time.Unix(1777555692, 0),
				CreatedAt:      time.Unix(1775761592, 0),
				UpdatedAt:      time.Unix(1022005, 0),
			},
			nil,
			nil,
		},
		{
			67,
			"otherfakenew67",
			0,
			private_keys.Key{
				ID:          67,
				Name:        "_GC3",
				Description: "",
				Algorithm:   key_crypto.AlgorithmRSA3072,
				Pem: `-----BEGIN RSA PRIVATE KEY-----
red-67
-----END RSA PRIVATE KEY-----
`,
				ApiKey:         "key-api-key-67",
				ApiKeyNew:      "otherfakenew67",
				ApiKeyDisabled: true,
				ApiKeyViaUrl:   false,
				LastAccess:     time.Unix(0, 0),
				CreatedAt:      time.Unix(1752418131, 0),
				UpdatedAt:      time.Unix(0, 0),
			},
			nil,
			nil,
		},
	}

	// create testing service
	storage, err := openStorageWithTestData(t, "putkeynewapikey")
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("id: %d)", tc.keyId), func(t *testing.T) {
			err := storage.PutKeyNewApiKey(tc.keyId, tc.apiKeyNew, tc.updateTimeUnix)
			if !errors.Is(err, tc.expectedPutErr) {
				t.Errorf("expected put error '%s' but got '%s'", test_helpers.ErrorToVal(tc.expectedPutErr), test_helpers.ErrorToVal(err))
			}

			key, err := storage.GetOneKeyById(tc.keyId)
			if !errors.Is(err, tc.expectedGetErr) {
				t.Errorf("expected get error '%s' but got '%s'", test_helpers.ErrorToVal(tc.expectedGetErr), test_helpers.ErrorToVal(err))
			}

			KeyCompare(t, key, tc.expectedKey)
		})
	}
}
