package migrations_test

import (
	"certwarden-backend/pkg/helpers_test"
	"certwarden-backend/pkg/storage/migrations"
	"certwarden-backend/pkg/storage/sqlite3"
	"context"
	"testing"

	"github.com/pressly/goose/v3"
)

func TestDBVersion1(t *testing.T) {
	thisTestFolder := tempFileStorage + "v1"

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

	// run the tests
	ctx, cancel := context.WithTimeout(context.Background(), migrations.MigrateDBTimeout)
	defer cancel()

	err = migrations.SetupGoose(ctx, db)
	if err != nil {
		t.Fatalf("goose setup failed: %s", err)
	}

	err = goose.UpToContext(ctx, db, "sql", 1)
	if err != nil {
		t.Fatalf("goose failed to up migrate v0 -> v1: %s", err)
	}

	// reverse
	err = goose.DownToContext(ctx, db, "sql", 0)
	if err != nil {
		t.Fatalf("goose failed to down migrate v1 -> v0: %s", err)
	}
}
