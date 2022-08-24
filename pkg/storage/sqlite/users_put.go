package sqlite

import "context"

// UpdateUserPassword updates the specified user's password hash to the specified
// hash.
func (store *Storage) UpdateUserPassword(username string, newPasswordHash string) (userId int, err error) {
	// database action
	ctx, cancel := context.WithTimeout(context.Background(), store.Timeout)
	defer cancel()

	query := `
	UPDATE
		users
	SET
		password_hash = $1,
		updated_at = $2
	WHERE
		username = $3
	RETURNING
		id
	`

	// update password and return id
	err = store.Db.QueryRowContext(ctx, query,
		newPasswordHash,
		timeNow(),
		username,
	).Scan(&userId)

	if err != nil {
		return -2, err
	}

	return userId, nil
}
