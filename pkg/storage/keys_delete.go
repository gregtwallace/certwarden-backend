package storage

import (
	"context"
	"database/sql"
	"errors"
)

// KeyInUse returns a bool if the specified key is in use, it returns
// an error if the key does not exist or any other error occurs. NOTE: This check
// includes if the key is assigned an acme account, certificate, OR is the private key
// for a current order (but the cert has been reconfigured to a different key). Therefore,
// it is possible for a key to return TRUE for inUse, but may also be in the "available"
// key list, returned by the GET available keys function.
func (store *Storage) KeyInUse(id int) (inUse bool, err error) {
	ctx, cancel := context.WithTimeout(store.shutdownContext, store.timeout)
	defer cancel()

	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	// check key exists
	query := `
	SELECT id
	FROM private_keys
	WHERE id = $1
	`

	row := tx.QueryRowContext(ctx, query, id)
	_discardVar := -2
	err = row.Scan(&_discardVar)
	if err != nil {
		// sql.ErrNoRows included here
		return false, err
	}

	// check not in use in accounts
	// if scan in succeeds, record exists in acme_accounts
	query = `
	SELECT id
	FROM acme_accounts
	WHERE private_key_id = $1
	`

	row = tx.QueryRowContext(ctx, query, id)
	err = row.Scan(&_discardVar)
	if !errors.Is(err, sql.ErrNoRows) {
		return true, err
	}

	// check not in use in certs
	// if scan in succeeds, record exists in certificates
	// this confirms a cert isn't trying to use this key in future orders
	query = `
	SELECT id
	FROM certificates
	WHERE private_key_id = $1
	`

	row = tx.QueryRowContext(ctx, query, id)
	err = row.Scan(&_discardVar)
	if !errors.Is(err, sql.ErrNoRows) {
		return true, err
	}

	// Even if the key isnt in use on a certificate, it may be in use on the certificate's most recent
	// valid order. In that case, the key is being actively served to CW clients. This query checks if
	// the key is in use on any such order by first getting the most recent valid order for each cert
	// then filtering to just orders using the key we're checking. If there is any result, the key is
	// in use by a most recent order.
	query = `
	SELECT
		certificate_id
	FROM
		acme_orders
	WHERE 
		status = "valid"
		AND
		known_revoked = 0
		AND
		pem NOT NULL
		AND
		certificate_id not null
	GROUP BY
		certificate_id
	HAVING
		MAX(created_at)
		AND
		finalized_key_id = $1
	`

	row = tx.QueryRowContext(ctx, query, id)
	err = row.Scan(&_discardVar)
	if !errors.Is(err, sql.ErrNoRows) {
		return true, err
	}

	err = tx.Commit()
	if err != nil {
		return false, err
	}

	return false, nil
}

// DeleteKey deletes a private key from the database
func (store *Storage) DeleteKey(id int) error {
	// check that delete is safe
	inUse, err := store.KeyInUse(id)
	if err != nil {
		return err
	}
	if inUse {
		return ErrInUse
	}

	ctx, cancel := context.WithTimeout(store.shutdownContext, store.timeout)
	defer cancel()

	// delete
	query := `
	DELETE FROM
		private_keys
	WHERE
		id = $1
	`

	res, err := store.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	// verify update actually happened
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected != 1 {
		return errorWrongAffectedRowCount(1, rowsAffected)
	}

	return nil
}
