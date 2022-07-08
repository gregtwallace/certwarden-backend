package sqlite

import (
	"context"
	"database/sql"
)

// DeleteKey deletes a private key from the database
func (storage *Storage) DeleteAccount(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), storage.Timeout)
	defer cancel()

	query := `
	DELETE FROM
		acme_accounts
	WHERE
		id = $1
	`

	// TODO: Ensure can't delete a key that is in use on an account or certificate
	// this is already checked by FK constraint, but perhaps update this later.

	result, err := storage.Db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	resultRows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	// if the record wasn't found, return no rows error
	// not exactly the purpose of this error, but it'll do
	if resultRows == 0 {
		return sql.ErrNoRows
	}

	return nil
}
