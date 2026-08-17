package storage_test

import (
	"certwarden-backend/pkg/domain/private_keys"
	"testing"
)

// compareKey compares key to expectedKey and throws appropriate errors for any differences
func compareKey(t *testing.T, key, expectedKey *private_keys.Key) {
	if key == nil && expectedKey == nil {
		return
	}
	if key == nil && expectedKey != nil {
		t.Errorf("key: key is nil but expectedKey is non-nil")
		return
	}
	if key != nil && expectedKey == nil {
		t.Errorf("key: key is non-nil but expectedKey is nil")
		return
	}

	if key.ID != expectedKey.ID {
		t.Errorf("key: id expected '%d' but got '%d'", expectedKey.ID, key.ID)
	}

	if key.Name != expectedKey.Name {
		t.Errorf("key: name expected '%s' but got '%s'", expectedKey.Name, key.Name)
	}

	if key.Description != expectedKey.Description {
		t.Errorf("key: description expected '%s' but got '%s'", expectedKey.Description, key.Description)
	}

	if key.Algorithm != expectedKey.Algorithm {
		t.Errorf("key: algorithm expected '%s' but got '%s'", expectedKey.Algorithm.StorageValue(), key.Algorithm.StorageValue())
	}

	if key.Pem != expectedKey.Pem {
		t.Errorf("key: pem expected '%s' but got '%s'", expectedKey.Pem, key.Pem)
	}

	if key.ApiKey != expectedKey.ApiKey {
		t.Errorf("key: apikey expected '%s' but got '%s'", expectedKey.ApiKey, key.ApiKey)
	}

	if key.ApiKeyNew != expectedKey.ApiKeyNew {
		t.Errorf("key: apikey new expected '%s' but got '%s'", expectedKey.ApiKeyNew, key.ApiKeyNew)
	}

	if key.ApiKeyDisabled != expectedKey.ApiKeyDisabled {
		t.Errorf("key: api key disabled expected '%t' but got '%t'", expectedKey.ApiKeyDisabled, key.ApiKeyDisabled)
	}

	if key.ApiKeyViaUrl != expectedKey.ApiKeyViaUrl {
		t.Errorf("key: api key via url expected '%t' but got '%t'", expectedKey.ApiKeyViaUrl, key.ApiKeyViaUrl)
	}

	if !key.LastAccess.Equal(expectedKey.LastAccess) {
		t.Errorf("key: last access expected '%s' but got '%s'", expectedKey.LastAccess.UTC(), key.LastAccess.UTC())
	}

	if !key.CreatedAt.Equal(expectedKey.CreatedAt) {
		t.Errorf("key: created at expected '%s' but got '%s'", expectedKey.CreatedAt.UTC(), key.CreatedAt.UTC())
	}

	if !key.UpdatedAt.Equal(expectedKey.UpdatedAt) {
		t.Errorf("key: updated at expected '%s' but got '%s'", expectedKey.UpdatedAt.UTC(), key.UpdatedAt.UTC())
	}
}
