package storage_test

import (
	"certwarden-backend/pkg/domain/private_keys/key_crypto"
	"certwarden-backend/pkg/test_helpers"
	"database/sql"
	"errors"
	"testing"
	"time"
)

// TODO: TestGetAllKeys

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
		t.Errorf("getonekeybyid: 22: expected error '%s' but got '%s'", sql.ErrNoRows, test_helpers.ErrorToVal(err))
	}

	// a key
	key, err := storage.GetOneKeyById(31)
	if err != nil {
		t.Errorf("getonekeybyid: 31: expected key but received err '%s'", err)
	}

	if key.ID != 31 {
		t.Errorf("getonekeybyid: 31: ID: expected 31 but got %d", key.ID)
	}
	if key.Name != "certwarden" {
		t.Errorf("getonekeybyid: 31: Name: expected 'certwarden' but got '%s'", key.Name)
	}
	if key.Description != "localhost / dev work w/ real cert" {
		t.Errorf("getonekeybyid: 31: Name: expected 'localhost / dev work w/ real cert' but got '%s'", key.Description)
	}
	if key.Algorithm != key_crypto.AlgorithmECDSAp256 {
		t.Errorf("getonekeybyid: 31: Algorithm: expected 'ecdsap256' but got '%s'", key.Algorithm.StorageValue())
	}
	if key.Pem != `-----BEGIN EC PRIVATE KEY-----
red-31
-----END EC PRIVATE KEY-----
` {
		t.Errorf(`getonekeybyid: 31: Pem: expected '-----BEGIN EC PRIVATE KEY-----
red-31
-----END EC PRIVATE KEY-----
' but got '%s'`, key.Pem)
	}
	if key.ApiKey != "key-api-key-4" {
		t.Errorf("getonekeybyid: 31: api key: expected 'key-api-key-4' but got '%s'", key.ApiKey)
	}
	if key.ApiKeyNew != "key-api-new-key-4" {
		t.Errorf("getonekeybyid: 31: api key new: expected 'key-api-new-key-4' but got '%s'", key.ApiKeyNew)
	}
	if key.ApiKeyDisabled {
		t.Errorf("getonekeybyid: 31: api key disabled: expected 'false' but got '%t'", key.ApiKeyDisabled)
	}
	if !key.ApiKeyViaUrl {
		t.Errorf("getonekeybyid: 31: api key via url: expected 'true' but got '%t'", key.ApiKeyViaUrl)
	}
	if !key.LastAccess.Equal(time.Unix(1745952074, 0)) {
		t.Errorf("getonekeybyid: 31: last access: expected '%s' but got '%s'", time.Unix(1745952074, 0).UTC(), key.LastAccess.UTC())
	}
	if !key.CreatedAt.Equal(time.Unix(1709327549, 0)) {
		t.Errorf("getonekeybyid: 31: created at: expected '%s' but got '%s'", time.Unix(1709327549, 0).UTC(), key.CreatedAt.UTC())
	}
	if !key.UpdatedAt.Equal(time.Unix(1732748628, 0)) {
		t.Errorf("getonekeybyid: 31: updated at: expected '%s' but got '%s'", time.Unix(1732748628, 0).UTC(), key.UpdatedAt.UTC())
	}

	// another key
	key, err = storage.GetOneKeyById(67)
	if err != nil {
		t.Errorf("getonekeybyid: 67: expected key but received err '%s'", err)
	}

	if key.ID != 67 {
		t.Errorf("getonekeybyid: 67: ID: expected 67 but got %d", key.ID)
	}
	if key.Name != "_GC3" {
		t.Errorf("getonekeybyid: 67: Name: expected '_GC3' but got '%s'", key.Name)
	}
	if key.Description != "" {
		t.Errorf("getonekeybyid: 67: Name: expected '' but got '%s'", key.Description)
	}
	if key.Algorithm != key_crypto.AlgorithmRSA3072 {
		t.Errorf("getonekeybyid: 67: Algorithm: expected 'rsa3072' but got '%s'", key.Algorithm.StorageValue())
	}
	if key.Pem != `-----BEGIN RSA PRIVATE KEY-----
red-67
-----END RSA PRIVATE KEY-----
` {
		t.Errorf(`getonekeybyid: 67: Pem: expected '-----BEGIN RSA PRIVATE KEY-----
red-67
-----END RSA PRIVATE KEY-----
' but got '%s'`, key.Pem)
	}
	if key.ApiKey != "key-api-key-67" {
		t.Errorf("getonekeybyid: 67: api key: expected 'key-api-key-67' but got '%s'", key.ApiKey)
	}
	if key.ApiKeyNew != "" {
		t.Errorf("getonekeybyid: 67: api key new: expected '' but got '%s'", key.ApiKeyNew)
	}
	if !key.ApiKeyDisabled {
		t.Errorf("getonekeybyid: 67: api key disabled: expected 'true' but got '%t'", key.ApiKeyDisabled)
	}
	if key.ApiKeyViaUrl {
		t.Errorf("getonekeybyid: 67: api key via url: expected 'false' but got '%t'", key.ApiKeyViaUrl)
	}
	if !key.LastAccess.Equal(time.Unix(0, 0)) {
		t.Errorf("getonekeybyid: 67: last access: expected '%s' but got '%s'", time.Unix(0, 0).UTC(), key.LastAccess.UTC())
	}
	if !key.CreatedAt.Equal(time.Unix(1752418131, 0)) {
		t.Errorf("getonekeybyid: 67: created at: expected '%s' but got '%s'", time.Unix(1752418131, 0).UTC(), key.CreatedAt.UTC())
	}
	if !key.UpdatedAt.Equal(time.Unix(1752418131, 0)) {
		t.Errorf("getonekeybyid: 67: updated at: expected '%s' but got '%s'", time.Unix(1752418131, 0).UTC(), key.UpdatedAt.UTC())
	}
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
		t.Errorf("getonekeybyid: fake-bad-name: expected error '%s' but got '%s'", sql.ErrNoRows, test_helpers.ErrorToVal(err))
	}

	// a key
	key, err := storage.GetOneKeyByName("certwarden")
	if err != nil {
		t.Errorf("getonekeybyid: 31: expected key but received err '%s'", err)
	}

	if key.ID != 31 {
		t.Errorf("getonekeybyid: 31: ID: expected 31 but got %d", key.ID)
	}
	if key.Name != "certwarden" {
		t.Errorf("getonekeybyid: 31: Name: expected 'certwarden' but got '%s'", key.Name)
	}
	if key.Description != "localhost / dev work w/ real cert" {
		t.Errorf("getonekeybyid: 31: Name: expected 'localhost / dev work w/ real cert' but got '%s'", key.Description)
	}
	if key.Algorithm != key_crypto.AlgorithmECDSAp256 {
		t.Errorf("getonekeybyid: 31: Algorithm: expected 'ecdsap256' but got '%s'", key.Algorithm.StorageValue())
	}
	if key.Pem != `-----BEGIN EC PRIVATE KEY-----
red-31
-----END EC PRIVATE KEY-----
` {
		t.Errorf(`getonekeybyid: 31: Pem: expected '-----BEGIN EC PRIVATE KEY-----
red-31
-----END EC PRIVATE KEY-----
' but got '%s'`, key.Pem)
	}
	if key.ApiKey != "key-api-key-4" {
		t.Errorf("getonekeybyid: 31: api key: expected 'key-api-key-4' but got '%s'", key.ApiKey)
	}
	if key.ApiKeyNew != "key-api-new-key-4" {
		t.Errorf("getonekeybyid: 31: api key new: expected 'key-api-new-key-4' but got '%s'", key.ApiKeyNew)
	}
	if key.ApiKeyDisabled {
		t.Errorf("getonekeybyid: 31: api key disabled: expected 'false' but got '%t'", key.ApiKeyDisabled)
	}
	if !key.ApiKeyViaUrl {
		t.Errorf("getonekeybyid: 31: api key via url: expected 'true' but got '%t'", key.ApiKeyViaUrl)
	}
	if !key.LastAccess.Equal(time.Unix(1745952074, 0)) {
		t.Errorf("getonekeybyid: 31: last access: expected '%s' but got '%s'", time.Unix(1745952074, 0).UTC(), key.LastAccess.UTC())
	}
	if !key.CreatedAt.Equal(time.Unix(1709327549, 0)) {
		t.Errorf("getonekeybyid: 31: created at: expected '%s' but got '%s'", time.Unix(1709327549, 0).UTC(), key.CreatedAt.UTC())
	}
	if !key.UpdatedAt.Equal(time.Unix(1732748628, 0)) {
		t.Errorf("getonekeybyid: 31: updated at: expected '%s' but got '%s'", time.Unix(1732748628, 0).UTC(), key.UpdatedAt.UTC())
	}

	// another key
	key, err = storage.GetOneKeyByName("_Another_Test_Acct_LE_Staging")
	if err != nil {
		t.Errorf("getonekeybyid: _Another_Test_Acct_LE_Staging: expected key but received err '%s'", err)
	}

	if key.ID != 63 {
		t.Errorf("getonekeybyid: _Another_Test_Acct_LE_Staging: ID: expected 63 but got %d", key.ID)
	}
	if key.Name != "_Another_Test_Acct_LE_Staging" {
		t.Errorf("getonekeybyid: _Another_Test_Acct_LE_Staging: Name: expected '_GC3' but got '%s'", key.Name)
	}
	if key.Description != "" {
		t.Errorf("getonekeybyid: _Another_Test_Acct_LE_Staging: Name: expected '' but got '%s'", key.Description)
	}
	if key.Algorithm != key_crypto.AlgorithmRSA2048 {
		t.Errorf("getonekeybyid: _Another_Test_Acct_LE_Staging: Algorithm: expected 'rsa2048' but got '%s'", key.Algorithm.StorageValue())
	}
	if key.Pem != `-----BEGIN RSA PRIVATE KEY-----
red-63
-----END RSA PRIVATE KEY-----
` {
		t.Errorf(`getonekeybyid: _Another_Test_Acct_LE_Staging: Pem: expected '-----BEGIN RSA PRIVATE KEY-----
red-63
-----END RSA PRIVATE KEY-----
' but got '%s'`, key.Pem)
	}
	if key.ApiKey != "key-api-key-63" {
		t.Errorf("getonekeybyid: _Another_Test_Acct_LE_Staging: api key: expected 'key-api-key-63' but got '%s'", key.ApiKey)
	}
	if key.ApiKeyNew != "" {
		t.Errorf("getonekeybyid: _Another_Test_Acct_LE_Staging: api key new: expected '' but got '%s'", key.ApiKeyNew)
	}
	if !key.ApiKeyDisabled {
		t.Errorf("getonekeybyid: _Another_Test_Acct_LE_Staging: api key disabled: expected 'true' but got '%t'", key.ApiKeyDisabled)
	}
	if key.ApiKeyViaUrl {
		t.Errorf("getonekeybyid: _Another_Test_Acct_LE_Staging: api key via url: expected 'false' but got '%t'", key.ApiKeyViaUrl)
	}
	if !key.LastAccess.Equal(time.Unix(0, 0)) {
		t.Errorf("getonekeybyid: _Another_Test_Acct_LE_Staging: last access: expected '%s' but got '%s'", time.Unix(0, 0).UTC(), key.LastAccess.UTC())
	}
	if !key.CreatedAt.Equal(time.Unix(1752254918, 0)) {
		t.Errorf("getonekeybyid: _Another_Test_Acct_LE_Staging: created at: expected '%s' but got '%s'", time.Unix(1752254918, 0).UTC(), key.CreatedAt.UTC())
	}
	if !key.UpdatedAt.Equal(time.Unix(1752254918, 0)) {
		t.Errorf("getonekeybyid: _Another_Test_Acct_LE_Staging: updated at: expected '%s' but got '%s'", time.Unix(1752254918, 0).UTC(), key.UpdatedAt.UTC())
	}
}

// TODO: TestGetAvailableKeys
