package sqlite

import (
	"context"
	"legocerthub-backend/pkg/domain/acme_accounts"
)

// PostNewAccount inserts a new account into the db
func (store *Storage) PostNewAccount(payload acme_accounts.NewPayload) (id int, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), store.timeout)
	defer cancel()

	// don't check for in use in storage. main app business logic should
	// take care of it

	// insert the new account
	query := `
	INSERT INTO acme_accounts (name, description, acme_server_id, private_key_id, status, email,
		accepted_tos, created_at, updated_at, kid)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	RETURNING id
	`

	err = store.db.QueryRowContext(ctx, query,
		payload.Name,
		payload.Description,
		payload.AcmeServerID,
		payload.PrivateKeyID,
		payload.Status,
		payload.Email,
		payload.AcceptedTos,
		payload.CreatedAt,
		payload.UpdatedAt,
		payload.Kid,
	).Scan(&id)

	if err != nil {
		return -2, err
	}

	return id, nil
}
