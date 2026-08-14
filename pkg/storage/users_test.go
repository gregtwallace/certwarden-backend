package storage_test

import (
	"certwarden-backend/pkg/domain/app/auth"
	"testing"
)

// compareUser compares user to expectedUser and throws appropriate errors for any differences
func compareUser(t *testing.T, user, expectedUser *auth.User) {
	if user == nil && expectedUser == nil {
		return
	}
	if user == nil && expectedUser != nil {
		t.Errorf("user: user is nil but expectedUser is non-nil")
		return
	}
	if user != nil && expectedUser == nil {
		t.Errorf("user: user is non-nil but expectedUser is nil")
		return
	}

	if user.ID != expectedUser.ID {
		t.Errorf("user: id expected '%d' but got '%d'", expectedUser.ID, user.ID)
	}

	if user.Username != expectedUser.Username {
		t.Errorf("user: username expected '%s' but got '%s'", expectedUser.Username, user.Username)
	}

	if user.PasswordHash != expectedUser.PasswordHash {
		t.Errorf("user: passwordhash expected '%s' but got '%s'", expectedUser.PasswordHash, user.PasswordHash)
	}

	if !user.CreatedAt.Equal(expectedUser.CreatedAt) {
		t.Errorf("key: created at expected '%s' but got '%s'", expectedUser.CreatedAt.UTC(), user.CreatedAt.UTC())
	}

	if !user.UpdatedAt.Equal(expectedUser.UpdatedAt) {
		t.Errorf("key: updated at expected '%s' but got '%s'", expectedUser.UpdatedAt.UTC(), user.UpdatedAt.UTC())
	}
}
