package sqlite

import (
	"context"
	"database/sql"
	"legocerthub-backend/pkg/domain/acme_accounts"
	"legocerthub-backend/pkg/storage"
)

// GetAllAccounts returns a slice of all of the Accounts in the database
func (store *Storage) GetAllAccounts() ([]acme_accounts.Account, error) {
	ctx, cancel := context.WithTimeout(context.Background(), store.Timeout)
	defer cancel()

	query := `SELECT aa.id, aa.name, aa.description, pk.id, pk.name,
	aa.status, aa.email, aa.is_staging, aa.accepted_tos
	FROM
		acme_accounts aa
		LEFT JOIN private_keys pk on (aa.private_key_id = pk.id)
	ORDER BY aa.name`

	rows, err := store.Db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var allAccounts []acme_accounts.Account
	for rows.Next() {
		var oneAccount accountDb
		err = rows.Scan(
			&oneAccount.id,
			&oneAccount.name,
			&oneAccount.description,
			&oneAccount.accountKeyDb.id,
			&oneAccount.accountKeyDb.name,
			&oneAccount.status,
			&oneAccount.email,
			&oneAccount.isStaging,
			&oneAccount.acceptedTos,
		)
		if err != nil {
			return nil, err
		}

		// convert to Account
		convertedAccount := oneAccount.toAccount()

		allAccounts = append(allAccounts, convertedAccount)
	}

	return allAccounts, nil
}

// GetOneAccountById returns an Account based on its unique id
func (store *Storage) GetOneAccountById(id int) (acme_accounts.AccountExtended, error) {
	return store.getOneAccount(id, "")
}

// GetOneAccountByName returns an Account based on its unique name
func (store *Storage) GetOneAccountByName(name string) (acme_accounts.AccountExtended, error) {
	return store.getOneAccount(-1, name)
}

// getOneAccount returns an Account based on either its unique id or its unique name
func (store *Storage) getOneAccount(id int, name string) (acme_accounts.AccountExtended, error) {
	ctx, cancel := context.WithTimeout(context.Background(), store.Timeout)
	defer cancel()

	query := `SELECT aa.id, aa.name, aa.description, pk.id, pk.name, pk.algorithm, pk.pem,
	aa.status, aa.email, aa.is_staging, aa.accepted_tos, aa.kid, aa.created_at, aa.updated_at
	FROM
		acme_accounts aa
		LEFT JOIN private_keys pk on (aa.private_key_id = pk.id)
	WHERE aa.id = $1 OR aa.name = $2
	ORDER BY aa.id`

	row := store.Db.QueryRowContext(ctx, query, id, name)

	var oneAccount accountDbExtended

	err := row.Scan(
		&oneAccount.id,
		&oneAccount.name,
		&oneAccount.description,
		&oneAccount.accountKeyDb.id,
		&oneAccount.accountKeyDb.name,
		&oneAccount.accountKeyDb.algorithmValue,
		&oneAccount.accountKeyDb.pem,
		&oneAccount.status,
		&oneAccount.email,
		&oneAccount.isStaging,
		&oneAccount.acceptedTos,
		&oneAccount.kid,
		&oneAccount.createdAt,
		&oneAccount.updatedAt,
	)

	if err != nil {
		// if no record exists
		if err == sql.ErrNoRows {
			err = storage.ErrNoRecord
		}
		return acme_accounts.AccountExtended{}, err
	}

	return oneAccount.toAccountExtended(), nil
}

// GetAvailableAccounts returns a slice of Accounts that exist and have a valid status
func (store *Storage) GetAvailableAccounts() (accts []acme_accounts.Account, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), store.Timeout)
	defer cancel()

	query := `
		SELECT
			aa.id, aa.name, aa.description, pk.id, pk.name,
			aa.status, aa.email, aa.is_staging, aa.accepted_tos
		FROM
			acme_accounts aa
			LEFT JOIN private_keys pk on (aa.private_key_id = pk.id)
		WHERE
			aa.status = "valid" AND
			aa.accepted_tos = true
		ORDER BY aa.name
	`

	rows, err := store.Db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var oneAccount accountDb

		err = rows.Scan(
			&oneAccount.id,
			&oneAccount.name,
			&oneAccount.description,
			&oneAccount.accountKeyDb.id,
			&oneAccount.accountKeyDb.name,
			&oneAccount.status,
			&oneAccount.email,
			&oneAccount.isStaging,
			&oneAccount.acceptedTos,
		)
		if err != nil {
			return nil, err
		}

		// convert to Account
		acct := oneAccount.toAccount()

		accts = append(accts, acct)
	}

	return accts, nil
}
