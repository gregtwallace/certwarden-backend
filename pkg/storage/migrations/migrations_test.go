package migrations_test

import (
	"certwarden-backend/pkg/helpers_test"
	"certwarden-backend/pkg/storage/migrations"
	"certwarden-backend/pkg/storage/sqlite3"
	"context"
	"io"
	"os"
	"testing"

	"github.com/pressly/goose/v3"
)

// TODO: TestSetupGoose (using new empty file && various versions of pragma)

func TestDo(t *testing.T) {
	// test empty file
	t.Run("empty file", func(t *testing.T) {
		thisTestFolder := tempFileStorage + "migrate_do_empty"

		// make temp data folder
		helpers_test.MakeTempStorage(t, thisTestFolder)

		// file will be new (empty)
		a := newFakeApp(t, thisTestFolder)

		db, cleanup, err := sqlite3.OpenSqlite3Database(a)
		if err != nil {
			t.Fatalf("failed to open sqlite3 db: %s", err)
		}
		t.Cleanup(cleanup)
		t.Cleanup(func() {
			err := db.Close()
			if err != nil {
				t.Errorf("failed to close storage (%s)", err)
			}
		})

		ctx, cancel := context.WithTimeout(context.Background(), migrations.MigrateDBTimeout)
		defer cancel()

		err = migrations.Do(ctx, db, a.GetLogger())
		if err != nil {
			t.Fatalf("failed to run migrations: %s", err)
		}

		// test: confirm we're on the latest version now
		err = goose.UpByOneContext(ctx, db, "sql")
		expectErr := helpers_test.NewTestErrorStringComp("no next version found")
		if !helpers_test.ErrorsIs(err, expectErr) {
			t.Errorf("expected err '%s' but got '%s'", expectErr, err)
		}
	})

	// test testdata_v11 file (doesnt contain `goose_db_version` table)
	t.Run("testdata_v11", func(t *testing.T) {
		thisTestFolder := tempFileStorage + "migrate_do_v11"

		// make temp data folder
		helpers_test.MakeTempStorage(t, thisTestFolder)

		// file will be new (empty)
		a := newFakeApp(t, thisTestFolder)

		// copy testdata_v11 file to temp folder

		testDataF, err := os.Open("../../../test_data/testdata_v11.db")
		if err != nil {
			t.Fatalf("failed to open test data file '%s'", err)
		}
		t.Cleanup(func() {
			err := testDataF.Close()
			if err != nil {
				t.Errorf("failed to close testDataF (%s)", err)
			}
		})

		testDataCopyF, err := os.Create(thisTestFolder + "/appdata.db")
		if err != nil {
			t.Fatalf("failed to create test data file copy '%s'", err)
		}
		t.Cleanup(func() {
			err := testDataCopyF.Close()
			if err != nil {
				t.Errorf("failed to close testDataCopyF (%s)", err)
			}
		})

		_, err = io.Copy(testDataCopyF, testDataF)
		if err != nil {
			t.Fatalf("failed to copy test data '%s'", err)
		}

		err = testDataCopyF.Sync()
		if err != nil {
			t.Fatalf("failed to sync test data '%s'", err)
		}

		// open db
		db, cleanup, err := sqlite3.OpenSqlite3Database(a)
		if err != nil {
			t.Fatalf("failed to open sqlite3 db: %s", err)
		}
		t.Cleanup(cleanup)
		t.Cleanup(func() {
			err := db.Close()
			if err != nil {
				t.Errorf("failed to close storage (%s)", err)
			}
		})

		ctx, cancel := context.WithTimeout(context.Background(), migrations.MigrateDBTimeout)
		defer cancel()

		err = migrations.Do(ctx, db, a.GetLogger())
		if err != nil {
			t.Fatalf("failed to run migrations: %s", err)
		}

		// test: confirm we're on the latest version now
		err = goose.UpByOneContext(ctx, db, "sql")
		expectErr := helpers_test.NewTestErrorStringComp("no next version found")
		if !helpers_test.ErrorsIs(err, expectErr) {
			t.Errorf("expected err '%s' but got '%s'", expectErr, err)
		}
	})

	// TODO: test next version where `goose_db_version` table MUST exist
}
