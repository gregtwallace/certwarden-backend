package auth

type User struct {
	ID           int
	Username     string
	PasswordHash string
	CreatedAt    int
	UpdatedAt    int
}
