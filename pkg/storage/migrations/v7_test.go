package migrations_test

import (
	"certwarden-backend/pkg/storage/migrations"
	"context"
	"testing"

	"github.com/pressly/goose/v3"
)

func TestDBVersion7(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), migrations.MigrateDBTimeout)
	defer cancel()

	db := setupDBV0(t, ctx, "migrate_v7")

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

	err = goose.UpToContext(ctx, db, "sql", 3)
	if err != nil {
		t.Fatalf("goose failed to up migrate v2 -> v3: %s", err)
	}

	insertDataV3(t, db)
	validateDataV3(t, db)

	err = goose.UpToContext(ctx, db, "sql", 4)
	if err != nil {
		t.Fatalf("goose failed to up migrate v3 -> v4: %s", err)
	}

	insertDataV4(t, db)
	validateDataV4(t, db)

	err = goose.UpToContext(ctx, db, "sql", 5)
	if err != nil {
		t.Fatalf("goose failed to up migrate v4 -> v5: %s", err)
	}

	insertDataV5(t, db)
	validateDataV5(t, db)

	err = goose.UpToContext(ctx, db, "sql", 6)
	if err != nil {
		t.Fatalf("goose failed to up migrate v5 -> v6: %s", err)
	}

	validateDataV6(t, db)

	err = goose.UpToContext(ctx, db, "sql", 7)
	if err != nil {
		t.Fatalf("goose failed to up migrate v6 -> v7: %s", err)
	}

	insertDataV7(t, db)
	validateDataV7(t, db)

	//
	// REVERSE AND GO DOWN
	//
	err = goose.DownToContext(ctx, db, "sql", 6)
	if err != nil {
		t.Fatalf("goose failed to down migrate v7 -> v6: %s", err)
	}

	validateDataV6(t, db)

	err = goose.DownToContext(ctx, db, "sql", 5)
	if err != nil {
		t.Fatalf("goose failed to down migrate v6 -> v5: %s", err)
	}

	validateDataV5(t, db)

	err = goose.DownToContext(ctx, db, "sql", 4)
	if err != nil {
		t.Fatalf("goose failed to down migrate v5 -> v4: %s", err)
	}

	validateDataV4(t, db)

	err = goose.DownToContext(ctx, db, "sql", 3)
	if err != nil {
		t.Fatalf("goose failed to down migrate v4 -> v3: %s", err)
	}

	validateDataV3(t, db)

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
