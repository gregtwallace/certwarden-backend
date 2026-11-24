package storage_test

import (
	"certwarden-backend/pkg/domain/private_keys"
	"certwarden-backend/pkg/domain/private_keys/key_crypto"
	"certwarden-backend/pkg/helpers_test"
	"database/sql"
	"fmt"
	"testing"
	"time"
)

func TestPostNewKey(t *testing.T) {
	testCases := []struct {
		newKeyPayload   private_keys.NewPayload
		expectedPostKey *private_keys.Key
		expectedPostErr error

		getName        string
		expectedGetKey *private_keys.Key
		expectedGetErr error
	}{
		{ // valid insertion
			private_keys.NewPayload{
				Name:           new("new_key_xyz"),
				Description:    new(""),
				AlgorithmValue: new(key_crypto.AlgorithmECDSAp256.StorageValue()),
				PemContent:     new("-- new pem xyz --"),
				ApiKey:         "apikeyxyz",
				ApiKeyDisabled: new(true),
				ApiKeyViaUrl:   true,
				CreatedAt:      time.Unix(1780336479, 0),
				UpdatedAt:      time.Unix(1780337000, 0),
			},
			&private_keys.Key{
				ID:             71,
				Name:           "new_key_xyz",
				Description:    "",
				Algorithm:      key_crypto.AlgorithmECDSAp256,
				Pem:            "-- new pem xyz --",
				ApiKey:         "apikeyxyz",
				ApiKeyNew:      "",
				ApiKeyDisabled: true,
				ApiKeyViaUrl:   true,
				LastAccess:     nil,
				CreatedAt:      time.Unix(1780336479, 0),
				UpdatedAt:      time.Unix(1780337000, 0),
			},
			nil,
			"new_key_xyz",
			&private_keys.Key{
				ID:             71,
				Name:           "new_key_xyz",
				Description:    "",
				Algorithm:      key_crypto.AlgorithmECDSAp256,
				Pem:            "-- new pem xyz --",
				ApiKey:         "apikeyxyz",
				ApiKeyNew:      "",
				ApiKeyDisabled: true,
				ApiKeyViaUrl:   true,
				LastAccess:     nil,
				CreatedAt:      time.Unix(1780336479, 0),
				UpdatedAt:      time.Unix(1780337000, 0),
			},
			nil,
		},
		{ // duplicate name (non-case sensitive)
			private_keys.NewPayload{
				Name:           new("le_staging"),
				Description:    new(""),
				AlgorithmValue: new(key_crypto.AlgorithmRSA2048.StorageValue()),
				PemContent:     new("-- new pem dupe --"),
				ApiKey:         "irrelevant",
				ApiKeyDisabled: new(false),
				ApiKeyViaUrl:   false,
				CreatedAt:      time.Unix(1780336477, 0),
				UpdatedAt:      time.Unix(1780337010, 0),
			},
			nil,
			helpers_test.NewTestErrorStringComp("UNIQUE constraint failed"),
			"le_staging",
			&key1,
			nil,
		},
		{ // incomplete payload
			private_keys.NewPayload{
				Name:           new("its_a_new_key_inco"),
				AlgorithmValue: new(key_crypto.AlgorithmRSA2048.StorageValue()),
				PemContent:     new("-- new pem again --"),
				ApiKey:         "irrelevant2",
				ApiKeyDisabled: new(false),
				ApiKeyViaUrl:   false,
				CreatedAt:      time.Unix(1780336480, 0),
				UpdatedAt:      time.Unix(1780337001, 0),
			},
			nil,
			helpers_test.NewTestErrorStringComp("NOT NULL constraint failed"),
			"its_a_new_key_inco",
			nil,
			sql.ErrNoRows,
		},
	}

	// create testing service
	store := openStorageWithTestData(t, "postnewkey")

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("post name: %s", helpers_test.StringPointerToVal(tc.newKeyPayload.Name)), func(t *testing.T) {
			key, err := store.PostNewKey(&tc.newKeyPayload)
			if !helpers_test.ErrorsIs(err, tc.expectedPostErr) {
				t.Errorf("expected post error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedPostErr), helpers_test.ErrorToVal(err))
			}

			compareKey(t, key, tc.expectedPostKey)

			key, err = store.GetOneKeyByName(tc.getName)
			if !helpers_test.ErrorsIs(err, tc.expectedGetErr) {
				t.Errorf("expected get error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedGetErr), helpers_test.ErrorToVal(err))
			}

			compareKey(t, key, tc.expectedGetKey)
		})
	}
}
