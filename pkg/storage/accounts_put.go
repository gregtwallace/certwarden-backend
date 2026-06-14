package storage

import (
	"certwarden-backend/pkg/domain/acme_accounts"
	"context"
	"errors"
	"fmt"
)

// PutAcmeAccountUpdate updates details about an acme account
func (store *Storage) PutAcmeAccountUpdate(payload acme_accounts.UpdatePayload) (acme_accounts.Account, error) {
	// database update
	ctx, cancel := context.WithTimeout(store.shutdownContext, store.timeout)
	defer cancel()

	query := `
	UPDATE
		acme_accounts
	SET
		name = case when $1 is null then name else $1 end,
		description = case when $2 is null then description else $2 end,
		status = case when $3 is null then status else $3 end,
		email = case when $4 is null then email else $4 end,
		kid = case when $5 is null then kid else $5 end,
		created_at = case when $6 is null then created_at else $6 end,
		updated_at = $7
	WHERE
		id = $8
	`

	// need special handling since created at could be null but needs to be converted to an int
	var createdAtUnix *int64
	if payload.CreatedAt != nil {
		// convert to unix time for storage
		createdAtUnix = new(payload.CreatedAt.Unix())
	}

	res, err := store.db.ExecContext(ctx, query,
		payload.Name,
		payload.Description,
		payload.Status,
		payload.Email,
		payload.KID,
		createdAtUnix,
		payload.UpdatedAt.Unix(),
		payload.ID,
	)
	if err != nil {
		return acme_accounts.Account{}, err
	}

	// verify update actually happened
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return acme_accounts.Account{}, err
	}
	if rowsAffected != 1 {
		return acme_accounts.Account{}, errors.Join(fmt.Errorf("expected 1 row update, but got '%d'", rowsAffected), ErrWrongUpdateRowCount)
	}

	// get updated account to return
	updatedAccount, err := store.GetOneAcmeAccountById(payload.ID)
	if err != nil {
		return acme_accounts.Account{}, err
	}

	return updatedAccount, nil
}

// PutAcmeAccountNewKey updates the specified account to the new key id
func (store *Storage) PutAcmeAccountNewKey(payload acme_accounts.RolloverKeyPayload) (acme_accounts.Account, error) {
	ctx, cancel := context.WithTimeout(store.shutdownContext, store.timeout)
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

	res, err := store.db.ExecContext(ctx, query,
		payload.PrivateKeyID,
		payload.UpdatedAt.Unix(),
		payload.ID,
	)
	if err != nil {
		return acme_accounts.Account{}, err
	}

	// verify update actually happened
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return acme_accounts.Account{}, err
	}
	if rowsAffected != 1 {
		return acme_accounts.Account{}, errors.Join(fmt.Errorf("expected 1 row update, but got '%d'", rowsAffected), ErrWrongUpdateRowCount)
	}

	// get updated account to return
	updatedAccount, err := store.GetOneAcmeAccountById(payload.ID)
	if err != nil {
		return acme_accounts.Account{}, err
	}

	return updatedAccount, nil
}
