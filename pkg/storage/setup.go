package storage

import (
	"certwarden-backend/pkg/domain/app/auth"
	"context"
	"database/sql"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// populateNewDb creates the tables in the db file and sets the db version
func (store *Storage) populateNewDb() error {
	ctx, cancel := context.WithTimeout(store.shutdownContext, store.timeout)
	defer cancel()

	// create sql transaction to roll back in the event an error occurs
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// set db user_version
	// No injection protection since const isn't user editable
	query := `PRAGMA user_version = ` + strconv.Itoa(DbCurrentUserVersion)

	_, err = tx.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	// create tables
	err = createDBTablesV11(tx)
	if err != nil {
		return err
	}

	// insert default ACME servers
	err = insertDefaultAcmeServers(tx)
	if err != nil {
		return err
	}

	// insert default user
	err = insertDefaultUser(tx)
	if err != nil {
		return err
	}

	// no errors, commit transaction
	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

// insertDefaultUser inserts the default user with the default password
func insertDefaultUser(tx *sql.Tx) error {
	// add default admin user
	// default username and password
	defaultUsername := "admin"
	defaultPassword := "password"

	// generate password hash
	defaultHashedPw, err := bcrypt.GenerateFromPassword([]byte(defaultPassword), auth.BcryptCost)
	if err != nil {
		return err
	}

	// insert
	query := `
	INSERT INTO
		users (username, password_hash, created_at, updated_at)
	VALUES (
		$1,
		$2,
		$3,
		$4
	)
	`

	_, err = tx.Exec(query,
		defaultUsername,
		defaultHashedPw,
		time.Now().Unix(),
		time.Now().Unix(),
	)

	if err != nil {
		return err
	}

	return nil
}
