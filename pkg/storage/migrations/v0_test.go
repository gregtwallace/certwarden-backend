package migrations_test

import (
	"certwarden-backend/pkg/storage/migrations"
	"context"
	"testing"
)

func TestDBVersion0(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), migrations.MigrateDBTimeout)
	defer cancel()

	db := setupDBV0(t, ctx, "migrate_v0")

	validateDataV0(t, db)
}
