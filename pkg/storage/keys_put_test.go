package storage_test

import (
	"certwarden-backend/pkg/domain/private_keys"
	"certwarden-backend/pkg/domain/private_keys/key_crypto"
	"certwarden-backend/pkg/helpers_test"
	"certwarden-backend/pkg/storage"
	"database/sql"
	"fmt"
	"testing"
	"time"
)

func TestPutKeyUpdate(t *testing.T) {
	testCases := []struct {
		payload           private_keys.UpdatePayload
		expectedPutResult private_keys.Key
		expectedPutErr    error
		getId             int
		expectedGetResult private_keys.Key
		expectedGetErr    error
	}{
		{ // invalid key
			private_keys.UpdatePayload{
				ID: -1,
			},
			private_keys.Key{},
			storage.ErrWrongUpdateRowCount,
			-1,
			private_keys.Key{},
			sql.ErrNoRows,
		},
		{ // invalid key
			private_keys.UpdatePayload{
				ID: 522,
			},
			private_keys.Key{},
			storage.ErrWrongUpdateRowCount,
			522,
			private_keys.Key{},
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
				ApiKey:         "key-api-key-31",
				ApiKeyNew:      "key-api-new-key-31",
				ApiKeyDisabled: false,
				ApiKeyViaUrl:   true,
				LastAccess:     time.Unix(1745952074, 0),
				CreatedAt:      time.Unix(1709327549, 0),
				UpdatedAt:      time.Unix(100555, 0),
			},
			nil,
			31,
			private_keys.Key{
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
				LastAccess:     time.Unix(1745952074, 0),
				CreatedAt:      time.Unix(1709327549, 0),
				UpdatedAt:      time.Unix(100555, 0),
			},
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
			62,
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
			58,
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
			if !helpers_test.ErrorsIs(err, tc.expectedPutErr) {
				t.Errorf("expected put error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedPutErr), helpers_test.ErrorToVal(err))
			}

			CompareKey(t, key, tc.expectedPutResult)

			key, err = storage.GetOneKeyById(tc.getId)
			if !helpers_test.ErrorsIs(err, tc.expectedGetErr) {
				t.Errorf("expected get error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedGetErr), helpers_test.ErrorToVal(err))
			}

			CompareKey(t, key, tc.expectedGetResult)
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
			storage.ErrWrongUpdateRowCount,
			sql.ErrNoRows,
		},
		{ // invalid key id
			500,
			"anotherfake",
			10005000,
			private_keys.Key{},
			storage.ErrWrongUpdateRowCount,
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
				ApiKeyNew:      "key-api-new-key-31",
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
			if !helpers_test.ErrorsIs(err, tc.expectedPutErr) {
				t.Errorf("expected put error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedPutErr), helpers_test.ErrorToVal(err))
			}

			key, err := storage.GetOneKeyById(tc.keyId)
			if !helpers_test.ErrorsIs(err, tc.expectedGetErr) {
				t.Errorf("expected get error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedGetErr), helpers_test.ErrorToVal(err))
			}

			CompareKey(t, key, tc.expectedKey)
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
			storage.ErrWrongUpdateRowCount,
			sql.ErrNoRows,
		},
		{ // invalid key id
			500,
			"anotherfake",
			10005000,
			private_keys.Key{},
			storage.ErrWrongUpdateRowCount,
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
			if !helpers_test.ErrorsIs(err, tc.expectedPutErr) {
				t.Errorf("expected put error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedPutErr), helpers_test.ErrorToVal(err))
			}

			key, err := storage.GetOneKeyById(tc.keyId)
			if !helpers_test.ErrorsIs(err, tc.expectedGetErr) {
				t.Errorf("expected get error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedGetErr), helpers_test.ErrorToVal(err))
			}

			CompareKey(t, key, tc.expectedKey)
		})
	}
}

func TestPutKeyLastAccess(t *testing.T) {
	testCases := []struct {
		keyId              int
		lastAccessTimeUnix int64

		expectedKey    private_keys.Key
		expectedPutErr error
		expectedGetErr error
	}{
		{ // invalid key id
			-1,
			88888888,
			private_keys.Key{},
			storage.ErrWrongUpdateRowCount,
			sql.ErrNoRows,
		},
		{ // invalid key id
			500,
			88888888,
			private_keys.Key{},
			storage.ErrWrongUpdateRowCount,
			sql.ErrNoRows,
		},
		// do some updates
		{
			64,
			1022885,
			private_keys.Key{
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
				LastAccess:     time.Unix(1022885, 0),
				CreatedAt:      time.Unix(1752258426, 0),
				UpdatedAt:      time.Unix(1752258426, 0),
			},
			nil,
			nil,
		},
		{
			63,
			9999999,
			private_keys.Key{
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
				LastAccess:     time.Unix(9999999, 0),
				CreatedAt:      time.Unix(1752254918, 0),
				UpdatedAt:      time.Unix(1752254918, 0),
			},
			nil,
			nil,
		},
		{
			62,
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
				ApiKey:         "key-api-key-62",
				ApiKeyNew:      "key-api-new-key-62",
				ApiKeyDisabled: false,
				ApiKeyViaUrl:   false,
				LastAccess:     time.Unix(0, 0),
				CreatedAt:      time.Unix(1751738296, 0),
				UpdatedAt:      time.Unix(1751738296, 0),
			},
			nil,
			nil,
		},
	}

	// create testing service
	storage, err := openStorageWithTestData(t, "putkeylastaccess")
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("id: %d)", tc.keyId), func(t *testing.T) {
			err := storage.PutKeyLastAccess(tc.keyId, tc.lastAccessTimeUnix)
			if !helpers_test.ErrorsIs(err, tc.expectedPutErr) {
				t.Errorf("expected put error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedPutErr), helpers_test.ErrorToVal(err))
			}

			key, err := storage.GetOneKeyById(tc.keyId)
			if !helpers_test.ErrorsIs(err, tc.expectedGetErr) {
				t.Errorf("expected get error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedGetErr), helpers_test.ErrorToVal(err))
			}

			CompareKey(t, key, tc.expectedKey)
		})
	}
}
