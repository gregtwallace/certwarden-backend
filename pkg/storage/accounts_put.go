package storage

import (
	"certwarden-backend/pkg/domain/acme_accounts"
	"context"
	"errors"
	"fmt"
)

// PutAcmeAccountNameDesc only updates the name and desc in the database
func (store *Storage) PutAcmeAccountNameDesc(payload acme_accounts.NameDescPayload) (acme_accounts.Account, error) {
	// database update
	ctx, cancel := context.WithTimeout(store.shutdownContext, store.timeout)
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

	res, err := store.db.ExecContext(ctx, query,
		payload.Name,
		payload.Description,
		payload.UpdatedAt,
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

// PutAcmeAccountResponse populates an account with data that is returned by an ACME server when
// an account is POSTed to
func (store *Storage) PutAcmeAccountResponse(payload acme_accounts.AcmeAccountUpdate) (acme_accounts.Account, error) {
	ctx, cancel := context.WithTimeout(store.shutdownContext, store.timeout)
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

	res, err := store.db.ExecContext(ctx, query,
		payload.Status,
		payload.Email(),
		payload.CreatedAt.Unix(),
		payload.UpdatedAt,
		payload.Location,
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
		payload.UpdatedAt,
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
