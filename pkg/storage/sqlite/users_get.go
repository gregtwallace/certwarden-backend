package sqlite

import (
	"context"
	"database/sql"
	"legocerthub-backend/pkg/domain/app/users"
	"legocerthub-backend/pkg/storage"
)

// dbToUser converts the user db object to app object
func (userDb *userDb) dbToUser() (user users.User) {
	return users.User{
		ID:           userDb.id,
		Username:     userDb.username,
		PasswordHash: userDb.passwordHash,
		CreatedAt:    userDb.createdAt,
		UpdatedAt:    userDb.updatedAt,
	}
}

// GetOneUserByName returns a user from the db based on
// username
func (store Storage) GetOneUserByName(username string) (users.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), store.Timeout)
	defer cancel()

	query := `
	SELECT
		id, username, password_hash, created_at, updated_at
	FROM
		users
	WHERE
		username = $1
	`

	row := store.Db.QueryRowContext(ctx, query, username)

	var user userDb
	err := row.Scan(
		&user.id,
		&user.username,
		&user.passwordHash,
		&user.createdAt,
		&user.updatedAt,
	)

	if err != nil {
		// if no record exists
		if err == sql.ErrNoRows {
			err = storage.ErrNoRecord
		}
		return users.User{}, err
	}

	convertedUser := user.dbToUser()

	return convertedUser, nil
}
