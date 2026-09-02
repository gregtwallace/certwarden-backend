package migrations

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"time"

	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
)

const (
	migrateDBTimeout = 10 * time.Second
)

//go:embed sql/*.sql
var embedMigrations embed.FS

// gooseLogger is a wrapper for *zap.SugaredLogger that implements goose.Logger
type gooseLogger struct {
	l *zap.SugaredLogger
}

func (gl *gooseLogger) Fatalf(format string, v ...any) {
	gl.l.Fatalf(format, v)
}

func (gl *gooseLogger) Printf(format string, v ...any) {
	gl.l.Infof(format, v)
}

// gooseLogger -- end

// Do makes all necessary db modifications to bring db to the current schema version.
// If db does not contain a `goose_db_version` table, one is created using the db's
// existing pragma version.
func Do(ctx context.Context, db *sql.DB, logger *zap.SugaredLogger) error {
	gl := &gooseLogger{
		l: logger,
	}
	goose.SetLogger(gl)

	err := setupGoose(ctx, db)
	if err != nil {
		return fmt.Errorf("storage: failed to setup goose (%w)", err)
	}

	goose.SetBaseFS(embedMigrations)
	err = goose.Up(db, "sql")
	if err != nil {
		return fmt.Errorf("storage: migration failed (%w)", err)
	}

	return nil
}

// setupGoose creates the `goose_db_version` using the db's existing pragma
// version. If the db already has the `goose_db_version` table, this is a no-op
func setupGoose(ctx context.Context, db *sql.DB) error {
	ctx, cancel := context.WithTimeout(ctx, migrateDBTimeout)
	defer cancel()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// check if `goose_db_version` exists
	q := `
		SELECT * 
		FROM goose_db_version
		LIMIT 1;
	`
	_, err = tx.ExecContext(ctx, q)
	// SHOULD error here if table doesn't exist (and thus we should continue); no error = exists = done
	if err == nil {
		return nil
	}

	// get pragma
	query := `PRAGMA user_version`
	row := tx.QueryRowContext(ctx, query)
	fileUserVersion := -1
	err = row.Scan(
		&fileUserVersion,
	)
	// error, assume pragma 0 and pray for the best
	if err != nil {
		fileUserVersion = 0
	}

	// insert goose table
	query = `CREATE TABLE 'goose_db_version' (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		version_id INTEGER NOT NULL,
		is_applied INTEGER NOT NULL,
		tstamp TIMESTAMP DEFAULT (datetime('now'))
	)`
	_, err = tx.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	// add the appropriate version entries
	for i := 0; i <= fileUserVersion; i++ {
		query = `
		INSERT INTO 'goose_db_version' (
			version_id, is_applied
		)
		VALUES (
			$1, 1
		);
	`
		_, err = tx.ExecContext(ctx, query,
			i,
		)
		if err != nil {
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
