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

func TestPostNewKey(t *testing.T) {
	testCases := []struct {
		newKeyPayload  private_keys.NewPayload
		expectPostErr  bool
		expectedNewKey private_keys.Key
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
				CreatedAt:      1780336479,
				UpdatedAt:      1780337000,
			},
			false,
			private_keys.Key{
				ID:             72,
				Name:           "new_key_xyz",
				Description:    "",
				Algorithm:      key_crypto.AlgorithmECDSAp256,
				Pem:            "-- new pem xyz --",
				ApiKey:         "apikeyxyz",
				ApiKeyNew:      "",
				ApiKeyDisabled: true,
				ApiKeyViaUrl:   true,
				LastAccess:     time.Unix(0, 0),
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
				CreatedAt:      1780336477,
				UpdatedAt:      1780337010,
			},
			true,
			private_keys.Key{},
			sql.ErrNoRows,
		},
		{ // incomplete payload
			private_keys.NewPayload{
				Name:           new("its_a_new_key_valid"),
				AlgorithmValue: new(key_crypto.AlgorithmRSA2048.StorageValue()),
				PemContent:     new("-- new pem again --"),
				ApiKey:         "irrelevant2",
				ApiKeyDisabled: new(false),
				ApiKeyViaUrl:   false,
				CreatedAt:      1780336480,
				UpdatedAt:      1780337001,
			},
			true,
			private_keys.Key{},
			sql.ErrNoRows,
		},
	}

	// create testing service
	storage, err := openStorageWithTestData(t, "postnewkey")
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("post name: %s", test_helpers.StringPointerToVal(tc.newKeyPayload.Name)), func(t *testing.T) {
			key, err := storage.PostNewKey(tc.newKeyPayload)
			if (err != nil && !tc.expectPostErr) || (err == nil && tc.expectPostErr) {
				t.Errorf("expected post error '%t' but got err '%s'", tc.expectPostErr, test_helpers.ErrorToVal(err))
			}

			KeyCompare(t, key, tc.expectedNewKey)

			key, err = storage.GetOneKeyById(tc.expectedNewKey.ID)
			if !errors.Is(err, tc.expectedGetErr) {
				t.Errorf("expected get error '%s' but got '%s'", test_helpers.ErrorToVal(tc.expectedGetErr), test_helpers.ErrorToVal(err))
			}

			KeyCompare(t, key, tc.expectedNewKey)
		})
	}
}
