package storage_test

import (
	"certwarden-backend/pkg/domain/acme_servers"
	"certwarden-backend/pkg/helpers_test"
	"certwarden-backend/pkg/storage"
	"fmt"
	"testing"
	"time"
)

// backupCheckErrOK triggers an error on t if err does not match the expected err
// Note: this includes deadline expiration as a lock error
func backupCheckErrOK(t *testing.T, err error, expectLockErr bool) {
	if expectLockErr {
		expectedErr := helpers_test.NewTestErrorStringComp("database is locked")
		if !helpers_test.ErrorsIs(err, expectedErr) {
			t.Errorf("err expected '%s' but got '%s'", expectedErr, helpers_test.ErrorToVal(err))
		}

	} else if err != nil {
		t.Errorf("err expected '%s' but got '%s'", helpers_test.ErrorToVal(nil), helpers_test.ErrorToVal(err))
	}
}

// backupTestBattery is the group of tests run both while db is locked and while it is unlocked
func backupTestBattery(t *testing.T, store *storage.Storage, expectLocked bool) {
	lockedStateTxt := "unlocked"
	if expectLocked {
		lockedStateTxt = "locked"
	}

	// read only
	t.Run(fmt.Sprintf("%s: get all acme accounts", lockedStateTxt), func(t *testing.T) {
		_, _, err := store.GetAllAcmeAccounts(queryBuilderForTest(5, 0, "", false))
		backupCheckErrOK(t, err, false)
	})

	t.Run(fmt.Sprintf("%s: get one key by id", lockedStateTxt), func(t *testing.T) {
		_, err := store.GetOneKeyById(62)
		backupCheckErrOK(t, err, false)
	})

	// trying to write
	t.Run(fmt.Sprintf("%s: put key api key", lockedStateTxt), func(t *testing.T) {
		err := store.PutKeyApiKey(1, "xyz", time.Unix(123, 0))
		backupCheckErrOK(t, err, expectLocked)
	})

	t.Run(fmt.Sprintf("%s: put acme server update", lockedStateTxt), func(t *testing.T) {
		payload := acme_servers.UpdatePayload{
			ID:        1,
			UpdatedAt: time.Unix(6323444, 0),
		}

		_, err := store.PutServerUpdate(&payload)
		backupCheckErrOK(t, err, expectLocked)
	})
}

func TestLockDBForBackup(t *testing.T) {
	// create testing service
	store := openStorageWithTestData(t, "lockdbforbackup")

	unlock, err := store.LockDBForBackup()
	if err != nil {
		t.Fatalf("failed to lock db: %s", err)
	}

	// try various operations (locked)
	backupTestBattery(t, store, true)

	// verify things work after unlock
	unlock()

	// try various operations (unlocked)
	backupTestBattery(t, store, false)
}
