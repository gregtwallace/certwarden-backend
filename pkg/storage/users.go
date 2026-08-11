package storage

import (
	"certwarden-backend/pkg/domain/app/auth"
	"time"
)

// userDb represents how users are stored in the db
type userDb struct {
	id           int
	username     string
	passwordHash string
	createdAt    int64
	updatedAt    int64
}

// dbToUser converts the user db object to app object
func (userDb *userDb) dbToUser() (user auth.User) {
	return auth.User{
		ID:           userDb.id,
		Username:     userDb.username,
		PasswordHash: userDb.passwordHash,
		CreatedAt:    time.Unix(userDb.createdAt, 0),
		UpdatedAt:    time.Unix(userDb.updatedAt, 0),
	}
}
