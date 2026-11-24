package storage

import (
	"certwarden-backend/pkg/domain/app/auth"
	"context"
)

// dbToUser converts the user db object to app object
func (userDb *userDb) dbToUser() (user auth.User) {
	return auth.User{
		ID:           userDb.id,
		Username:     userDb.username,
		PasswordHash: userDb.passwordHash,
		CreatedAt:    userDb.createdAt,
		UpdatedAt:    userDb.updatedAt,
	}
}

// GetOneUserByName returns a user from the db based on
// username
func (store Storage) GetOneUserByName(username string) (auth.User, error) {
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
		return auth.User{}, err
	}

	convertedUser := user.dbToUser()

	return convertedUser, nil
}
