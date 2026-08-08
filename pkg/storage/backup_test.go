package storage_test

import (
	"certwarden-backend/pkg/domain/acme_servers"
	"certwarden-backend/pkg/helpers_test"
	"certwarden-backend/pkg/storage"
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

// backupCheckErrOK triggers an error on t if err does not match the expected err
// Note: this includes deadline expiration as a lock error
func backupCheckErrOK(t *testing.T, err error, expectLockErr bool) {
	if expectLockErr {
		if !helpers_test.ErrorsIs(err, context.DeadlineExceeded) && !helpers_test.ErrorsIs(err, helpers_test.NewTestErrorStringComp("database is locked")) {
			t.Errorf("err expected '%s' but got '%s'", helpers_test.ErrorToVal(context.DeadlineExceeded), helpers_test.ErrorToVal(err))
		}
	} else {
		if err != nil {
			t.Errorf("err expected '%s' but got '%s'", helpers_test.ErrorToVal(nil), helpers_test.ErrorToVal(err))
		}
	}
}

// backupTestBattery is the group of tests run both while db is locked and while it is unlocked
func backupTestBattery(t *testing.T, storage *storage.Storage, expectLocked bool) {
	wg := sync.WaitGroup{}
	lockedStateTxt := "unlocked"
	if !expectLocked {
		lockedStateTxt = "locked"
	}

	// read only
	wg.Add(1)
	go t.Run(fmt.Sprintf("%s: get all acme accounts", lockedStateTxt), func(t *testing.T) {
		_, _, err := storage.GetAllAcmeAccounts(queryBuilderForTest(5, 0, "", true))
		backupCheckErrOK(t, err, false)
		wg.Done()
	})

	wg.Add(1)
	go t.Run(fmt.Sprintf("%s: get one key by id", lockedStateTxt), func(t *testing.T) {
		_, err := storage.GetOneKeyById(62)
		backupCheckErrOK(t, err, false)
		wg.Done()
	})

	wg.Wait()

	// trying to write
	wg.Add(1)
	go t.Run(fmt.Sprintf("%s: put key api key", lockedStateTxt), func(t *testing.T) {
		err := storage.PutKeyApiKey(1, "xyz", time.Unix(123, 0))
		backupCheckErrOK(t, err, expectLocked)
		wg.Done()
	})

	wg.Add(1)
	go t.Run(fmt.Sprintf("%s: put acme server update", lockedStateTxt), func(t *testing.T) {
		payload := acme_servers.UpdatePayload{
			ID:        1,
			UpdatedAt: time.Unix(6323444, 0),
		}

		_, err := storage.PutServerUpdate(payload)
		backupCheckErrOK(t, err, expectLocked)
		wg.Done()
	})

	// wait for all tests
	wg.Wait()
}

func TestLockDBForBackup(t *testing.T) {
	// create testing service
	storage, err := openStorageWithTestData(t, "lockdbforbackup")
	if err != nil {
		t.Fatal(err)
	}

	unlock, err := storage.LockDBForBackup()
	if err != nil {
		t.Fatalf("failed to lock db: %s", err)
	}

	// try various operations (locked)
	backupTestBattery(t, storage, true)

	// verify things work after unlock
	unlock()

	// try various operations (unlocked)
	backupTestBattery(t, storage, false)
}
