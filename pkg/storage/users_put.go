package storage

import (
	"context"
	"time"
)

// PutUserPasswordHash updates the specified user's password hash to the specified hash.
func (store *Storage) PutUserPasswordHash(username, passwordHash string, updatedAt time.Time) (userId int, err error) {
	// database action
	ctx, cancel := context.WithTimeout(store.shutdownContext, store.timeout)
	defer cancel()

	query := `
	UPDATE
		users
	SET
		password_hash = $1,
		updated_at = $2
	WHERE
		username = $3
	`

	res, err := store.db.ExecContext(ctx, query,
		passwordHash,
		updatedAt.Unix(),
		username,
	)
	if err != nil {
		return -2, err
	}

	// verify update actually happened
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return -2, err
	}
	if rowsAffected != 1 {
		return -2, errorWrongAffectedRowCount(1, rowsAffected)
	}

	// get updated key to return
	updatedUser, err := store.GetOneUserByUsername(username)
	if err != nil {
		return -2, err
	}

	return updatedUser.ID, nil
}
