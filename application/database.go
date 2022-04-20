package application

import (
	"context"
	"database/sql"
	"time"
)

// function opens connection to the sqlite database
//   this will also cause the file to be created, if it does not exist
func OpenDB(cfg Config) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", cfg.Db.Dsn)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}
