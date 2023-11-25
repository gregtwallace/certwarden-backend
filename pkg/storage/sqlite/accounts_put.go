package sqlite

import (
	"context"
	"legocerthub-backend/pkg/domain/acme_accounts"
)

// PutNameDescAccount only updates the name and desc in the database
// TODO: refactor to more generic for anything that can be updated??
func (store *Storage) PutNameDescAccount(payload acme_accounts.NameDescPayload) (acme_accounts.Account, error) {
	// database update
	ctx, cancel := context.WithTimeout(context.Background(), store.timeout)
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

	_, err := store.db.ExecContext(ctx, query,
		payload.Name,
		payload.Description,
		payload.UpdatedAt,
		payload.ID,
	)

	if err != nil {
		return acme_accounts.Account{}, err
	}

	// get updated account to return
	updatedAccount, err := store.GetOneAccountById(payload.ID)
	if err != nil {
		return acme_accounts.Account{}, err
	}

	return updatedAccount, nil
}

// PutAcmeAccountResponse populates an account with data that is returned by LE when
// an account is POSTed to
func (store *Storage) PutAcmeAccountResponse(payload acme_accounts.AcmeAccount) (acme_accounts.Account, error) {
	ctx, cancel := context.WithTimeout(context.Background(), store.timeout)
	defer cancel()

	query := `
	UPDATE
		acme_accounts
	SET
		status = $1,
		email = $2,
		created_at = case when $3 is 0 or $3 is null then created_at else $3 end,
		updated_at = $4,
		kid = case when $5 is "" or $5 is null then kid else $5 end
	WHERE
		id = $6`

	_, err := store.db.ExecContext(ctx, query,
		payload.Status,
		payload.Email(),
		payload.CreatedAt.ToUnixTime(),
		payload.UpdatedAt,
		payload.Location,
		payload.ID,
	)
	if err != nil {
		return acme_accounts.Account{}, err
	}
	// TODO: Handle 0 rows updated.

	// get updated account to return
	updatedAccount, err := store.GetOneAccountById(payload.ID)
	if err != nil {
		return acme_accounts.Account{}, err
	}

	return updatedAccount, nil
}

// PutNewAccountKey updates the specified account to the new key id
func (store *Storage) PutNewAccountKey(payload acme_accounts.RolloverKeyPayload) (acme_accounts.Account, error) {
	ctx, cancel := context.WithTimeout(context.Background(), store.timeout)
	defer cancel()

	query := `
	UPDATE
		acme_accounts
	SET
		private_key_id = $1,
		updated_at = $2
	WHERE
		id = $3
	`

	_, err := store.db.ExecContext(ctx, query,
		payload.PrivateKeyID,
		payload.UpdatedAt,
		payload.ID,
	)
	if err != nil {
		return acme_accounts.Account{}, err
	}
	// TODO: Handle 0 rows updated.

	// get updated account to return
	updatedAccount, err := store.GetOneAccountById(payload.ID)
	if err != nil {
		return acme_accounts.Account{}, err
	}

	return updatedAccount, nil
}
