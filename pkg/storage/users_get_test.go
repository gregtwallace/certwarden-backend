package storage_test

import (
	"certwarden-backend/pkg/domain/app/auth"
	"certwarden-backend/pkg/helpers_test"
	"database/sql"
	"fmt"
	"testing"
	"time"
)

var (
	user1 = auth.User{
		ID:           1,
		Username:     "admin",
		PasswordHash: "xYz",
		CreatedAt:    time.Unix(1697139775, 0),
		UpdatedAt:    time.Unix(1738009344, 0),
	}
	user2 = auth.User{
		ID:           2,
		Username:     "user2",
		PasswordHash: "abc",
		CreatedAt:    time.Unix(255111225, 0),
		UpdatedAt:    time.Unix(122544466, 0),
	}
	user4 = auth.User{
		ID:           4,
		Username:     "user4",
		PasswordHash: "1234b",
		CreatedAt:    time.Unix(88222222, 0),
		UpdatedAt:    time.Unix(22222222, 0),
	}
)

func TestGetOneUserByName(t *testing.T) {
	testCases := []struct {
		username     string
		expectedUser auth.User
		expectedErr  error
	}{
		{"", auth.User{}, sql.ErrNoRows},
		{"fake-bad-username", auth.User{}, sql.ErrNoRows},
		{"admin", user1, nil},
		// {"AdMiN", user1, nil}, // case is wrong TODO: make case insensitive
		{"user2", user2, nil},
		{"user4", user4, nil},
	}

	// create testing service
	store, err := openStorageWithTestData(t, "getoneuserbyusername")
	if err != nil {
		t.Fatal(err)
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("#%d (name: %s)", i, tc.username), func(t *testing.T) {
			user, err := store.GetOneUserByUsername(tc.username)
			if !helpers_test.ErrorsIs(err, tc.expectedErr) {
				t.Errorf("expected get username error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedErr), helpers_test.ErrorToVal(err))
			}

			compareUser(t, &user, &tc.expectedUser)
		})
	}
}
