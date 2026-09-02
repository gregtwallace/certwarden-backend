package migrations_test

import (
	"certwarden-backend/pkg/storage/migrations"
	"context"
	"testing"

	"github.com/pressly/goose/v3"
)

func TestDBVersion2(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), migrations.MigrateDBTimeout)
	defer cancel()

	db := setupDBV0(t, ctx, "migrate_v2")

	// GO UP

	validateDataV0(t, db)

	err := goose.UpToContext(ctx, db, "sql", 1)
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

	//
	// REVERSE AND GO DOWN
	//
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
