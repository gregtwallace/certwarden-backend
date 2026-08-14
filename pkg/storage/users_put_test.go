package storage_test

import (
	"certwarden-backend/pkg/domain/app/auth"
	"certwarden-backend/pkg/helpers_test"
	"certwarden-backend/pkg/storage"
	"database/sql"
	"fmt"
	"testing"
	"time"
)

func TestPutUserPasswordHash(t *testing.T) {
	testCases := []struct {
		username        string
		newPasswordHash string
		updatedAt       time.Time

		expectedPutId   int
		expectedPutErr  error
		expectedGetUser *auth.User
		expectedGetErr  error
	}{
		{
			"",
			"somehash",
			time.Unix(12222222, 0),
			-2,
			storage.ErrWrongUpdateRowCount,
			nil,
			sql.ErrNoRows,
		},
		{
			"fake-bad-username",
			"somehash",
			time.Unix(12222223, 0),
			-2,
			storage.ErrWrongUpdateRowCount,
			nil,
			sql.ErrNoRows,
		},
		{
			"uSEr2", // case is wrong TODO: make case insensitive
			"newHAsh456",
			time.Unix(12244224, 0),
			-2,
			storage.ErrWrongUpdateRowCount,
			nil,
			sql.ErrNoRows,
		},
		{
			"user4",
			"anewhash",
			time.Unix(22422224, 0),
			4,
			nil,
			&auth.User{
				ID:           4,
				Username:     "user4",
				PasswordHash: "anewhash",
				CreatedAt:    time.Unix(88222222, 0),
				UpdatedAt:    time.Unix(22422224, 0),
			},
			nil,
		},
	}

	// create testing service
	store, err := openStorageWithTestData(t, "putuserpasswordhash")
	if err != nil {
		t.Fatal(err)
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("#%d (name: %s)", i, tc.username), func(t *testing.T) {
			userId, err := store.PutUserPasswordHash(tc.username, tc.newPasswordHash, tc.updatedAt)
			if !helpers_test.ErrorsIs(err, tc.expectedPutErr) {
				t.Errorf("expected put username passwordhash error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedPutErr), helpers_test.ErrorToVal(err))
			}

			if userId != tc.expectedPutId {
				t.Errorf("expected put username passwordhash return val '%d' but got '%d'", tc.expectedPutId, userId)
			}

			user, err := store.GetOneUserByUsername(tc.username)
			if !helpers_test.ErrorsIs(err, tc.expectedGetErr) {
				t.Errorf("expected get username error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedGetErr), helpers_test.ErrorToVal(err))
			}

			compareUser(t, user, tc.expectedGetUser)
		})
	}
}
