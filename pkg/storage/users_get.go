package storage

import (
	"certwarden-backend/pkg/domain/app/auth"
	"context"
)

// GetOneUserByUsername returns a user from the db with the specified username
func (store Storage) GetOneUserByUsername(username string) (*auth.User, error) {
	ctx, cancel := context.WithTimeout(store.shutdownContext, store.timeout)
	defer cancel()

	query := `
	SELECT
		id, username, password_hash, created_at, updated_at
	FROM
		users
	WHERE
		username = $1
	`

	row := store.db.QueryRowContext(ctx, query, username)

	var user userDb
	err := row.Scan(
		&user.id,
		&user.username,
		&user.passwordHash,
		&user.createdAt,
		&user.updatedAt,
	)

	if err != nil {
		return nil, err
	}

	convertedUser := user.dbToUser()

	return convertedUser, nil
}
