package sqlite

import (
	"context"
	"legocerthub-backend/pkg/storage"
)

// KeyInUse returns a bool if the specified key is in use, it returns
// an error if the key does not exist (or any other error)
func (store *Storage) KeyInUse(id int) (inUse bool, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), store.Timeout)
	defer cancel()

	// check key exists
	// if scan in succeeds, key exists
	query := `
	SELECT id
	FROM private_keys
	WHERE id = $1
	`

	row := store.Db.QueryRowContext(ctx, query, id)
	temp := -2
	row.Scan(&temp)
	if temp == -2 {
		return false, storage.ErrNoRecord
	}

	// check not in use in accounts
	// if scan in succeeds, record exists in acme_accounts
	query = `
	SELECT id
	FROM acme_accounts
	WHERE private_key_id = $1
	`

	row = store.Db.QueryRowContext(ctx, query, id)
	temp = -2
	row.Scan(&temp)
	if temp != -2 {
		return true, nil
	}

	// check not in use in certs
	// if scan in succeeds, record exists in certificates
	// this confirms a cert isn't trying to use this key in future orders
	query = `
	SELECT id
	FROM certificates
	WHERE private_key_id = $1
	`

	row = store.Db.QueryRowContext(ctx, query, id)
	temp = -2
	row.Scan(&temp)
	if temp != -2 {
		return true, nil
	}

	// check not in use by valid order on an existing cert_id with longest dated expiration
	// query groups valid orders by cert_id and then returns a result if the
	// order has the max valid_to for a particular key (the one being deleted)
	// This query is used to confirm the key isn't needed for any of the currently hosted cert records.
	// Needed because cert could be rotating key (cert record has a different key id now) and
	// don't want to delete key that may still be needed by pre-rotation cert
	query = `
	SELECT
		certificate_id
	FROM
		acme_orders
	WHERE 
		status = "valid"
		AND
		certificate_id not null
	GROUP BY
		certificate_id
	HAVING
		MAX(valid_to)
		AND
		finalized_key_id = $1
	
	`

	row = store.Db.QueryRowContext(ctx, query, id)
	temp = -2
	row.Scan(&temp)
	if temp != -2 {
		return true, nil
	}

	return false, nil
}

// DeleteKey deletes a private key from the database
func (store *Storage) DeleteKey(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), store.Timeout)
	defer cancel()

	// check that delete is safe
	inUse, err := store.KeyInUse(id)
	if err != nil {
		return err
	}
	if inUse {
		return storage.ErrInUse
	}

	// delete
	query := `
	DELETE FROM
		private_keys
	WHERE
		id = $1
	`

	_, err = store.Db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	return nil
}
