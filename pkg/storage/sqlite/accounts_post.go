package sqlite

import (
	"context"
	"legocerthub-backend/pkg/domain/acme_accounts"
)

// PostNewAccount inserts a new account into the db
func (store *Storage) PostNewAccount(payload acme_accounts.NewPayload) (id int, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), store.Timeout)
	defer cancel()

	// don't check for in use in storage. main app business logic should
	// take care of it

	// insert the new account
	query := `
	INSERT INTO acme_accounts (name, description, private_key_id, status, email, is_staging, accepted_tos, created_at, updated_at, kid)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	RETURNING id
	`

	err = store.Db.QueryRowContext(ctx, query,
		payload.Name,
		payload.Description,
		payload.PrivateKeyID,
		payload.Status,
		payload.Email,
		payload.IsStaging,
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
