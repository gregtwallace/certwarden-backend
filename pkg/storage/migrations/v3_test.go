package migrations_test

import (
	"certwarden-backend/pkg/helpers_test"
	"certwarden-backend/pkg/storage/migrations"
	"certwarden-backend/pkg/storage/sqlite3"
	"context"
	"testing"

	"github.com/pressly/goose/v3"
)

func TestDBVersion3(t *testing.T) {
	thisTestFolder := tempFileStorage + "migrate_v3"

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

	// GO UP

	validateDataV0(t, db)

	err = goose.UpToContext(ctx, db, "sql", 1)
	if err != nil {
		t.Fatalf("goose failed to up migrate v0 -> v1: %s", err)
	}

	insertDataV1(t, db)
	validateDataV1(t, db, false)

	err = goose.UpToContext(ctx, db, "sql", 2)
	if err != nil {
		t.Fatalf("goose failed to up migrate v1 -> v2: %s", err)
	}

	insertDataV2(t, db)
	validateDataV2(t, db)

	err = goose.UpToContext(ctx, db, "sql", 3)
	if err != nil {
		t.Fatalf("goose failed to up migrate v2 -> v3: %s", err)
	}

	insertDataV3(t, db)
	validateDataV3(t, db, false)

	//
	// REVERSE AND GO DOWN
	//
	err = goose.DownToContext(ctx, db, "sql", 2)
	if err != nil {
		t.Fatalf("goose failed to down migrate v3 -> v2: %s", err)
	}

	validateDataV2(t, db)

	err = goose.DownToContext(ctx, db, "sql", 1)
	if err != nil {
		t.Fatalf("goose failed to down migrate v2 -> v1: %s", err)
	}

	validateDataV1(t, db, true)

	err = goose.DownToContext(ctx, db, "sql", 0)
	if err != nil {
		t.Fatalf("goose failed to down migrate v1 -> v0: %s", err)
	}

	validateDataV0(t, db)
}
