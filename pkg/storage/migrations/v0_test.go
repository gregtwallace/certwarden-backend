package migrations_test

import (
	"certwarden-backend/pkg/helpers_test"
	"certwarden-backend/pkg/storage/migrations"
	"certwarden-backend/pkg/storage/sqlite3"
	"context"
	"testing"
)

func TestDBVersion0(t *testing.T) {
	thisTestFolder := tempFileStorage + "migrate_v0"

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

	// setup empty db
	ctx, cancel := context.WithTimeout(context.Background(), migrations.MigrateDBTimeout)
	defer cancel()

	err = migrations.SetupGoose(ctx, db)
	if err != nil {
		t.Fatalf("goose setup failed: %s", err)
	}

	validateDataV0(t, db)
}
