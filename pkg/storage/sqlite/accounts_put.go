package sqlite

import (
	"context"
	"legocerthub-backend/pkg/domain/acme_accounts"
)

// PutNameDescAccount only updates the name and desc in the database
// TODO: refactor to more generic for anything that can be updated??
func (store *Storage) PutNameDescAccount(payload acme_accounts.NameDescPayload) (err error) {
	// database update
	ctx, cancel := context.WithTimeout(context.Background(), store.Timeout)
	defer cancel()

	query := `
	UPDATE
		acme_accounts
	SET
		name = case when $1 is null then name else $1 end,
		description = case when $2 is null then description else $2 end,
		updated_at = $3
	WHERE
		id = $4
	`

	_, err = store.Db.ExecContext(ctx, query,
		payload.Name,
		payload.Description,
		payload.UpdatedAt,
		payload.ID,
	)

	if err != nil {
		return err
	}

	return nil
}

// PutLEAccountResponse populates an account with data that is returned by LE when
// an account is POSTed to
func (store *Storage) PutAcmeAccountResponse(payload acme_accounts.AcmeAccount) error {
	ctx, cancel := context.WithTimeout(context.Background(), store.Timeout)
	defer cancel()

	query := `
	UPDATE
		acme_accounts
	SET
		status = $1,
		email = $2,
		created_at = $3,
		updated_at = $4,
		kid = case when $5 is "" or $5 is null then kid else $5 end
	WHERE
		id = $6`

	_, err := store.Db.ExecContext(ctx, query,
		payload.Status,
		payload.Email(),
		payload.CreatedAt.ToUnixTime(),
		payload.UpdatedAt,
		payload.Location,
		payload.ID,
	)
	if err != nil {
		return err
	}

	// TODO: Handle 0 rows updated.
	return nil
}
