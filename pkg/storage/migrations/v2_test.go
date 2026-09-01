package migrations_test

import (
	"certwarden-backend/pkg/helpers_test"
	"certwarden-backend/pkg/storage/migrations"
	"certwarden-backend/pkg/storage/sqlite3"
	"context"
	"testing"

	"github.com/pressly/goose/v3"
)

func TestDBVersion2(t *testing.T) {
	thisTestFolder := tempFileStorage + "v2"

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

	// setup prev version db
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

	// run tests
	err = goose.UpToContext(ctx, db, "sql", 2)
	if err != nil {
		t.Fatalf("goose failed to up migrate v1 -> v2: %s", err)
	}

	// validate attribute was dropped
	q := `
		SELECT challenge_method
		FROM certificates
		LIMIT 1
	`
	_, err = db.ExecContext(ctx, q)
	if !helpers_test.ErrorsIs(err, helpers_test.NewTestErrorStringComp("no such column: challenge_method")) {
		t.Errorf("v2 should not have certificates.challenge_method attribute (err was %s)", helpers_test.ErrorToVal(err))
	}

	// reverse
	err = goose.DownToContext(ctx, db, "sql", 1)
	if err != nil {
		t.Fatalf("goose failed to down migrate v2 -> v1: %s", err)
	}

	// validate attribute is back
	q = `
		SELECT challenge_method
		FROM certificates
		LIMIT 1
	`
	_, err = db.ExecContext(ctx, q)
	if !helpers_test.ErrorsIs(err, nil) {
		t.Errorf("v1 should have certificates.challenge_method attribute (but err was %s)", helpers_test.ErrorToVal(err))
	}
}
