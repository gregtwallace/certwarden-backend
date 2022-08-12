package sqlite

// userDb represents how users are stored in the db
type userDb struct {
	id           int
	username     string
	passwordHash string
	createdAt    int
	updatedAt    int
}
