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

	query := `
	SELECT
		aa.id, aa.name, aa.description, aa.status, aa.email, aa.accepted_tos, aa.is_staging,
		aa.created_at, aa.updated_at, aa.kid,
		pk.id, pk.name, pk.description, pk.algorithm, pk.pem, pk.api_key, pk.api_key_via_url,
		pk.created_at, pk.updated_at
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
			&oneAccount.status,
			&oneAccount.email,
			&oneAccount.acceptedTos,
			&oneAccount.isStaging,
			&oneAccount.createdAt,
			&oneAccount.updatedAt,
			&oneAccount.kid,

			&oneAccount.accountKeyDb.id,
			&oneAccount.accountKeyDb.name,
			&oneAccount.accountKeyDb.description,
			&oneAccount.accountKeyDb.algorithmValue,
			&oneAccount.accountKeyDb.pem,
			&oneAccount.accountKeyDb.apiKey,
			&oneAccount.accountKeyDb.apiKeyViaUrl,
			&oneAccount.accountKeyDb.createdAt,
			&oneAccount.accountKeyDb.updatedAt,
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
func (store *Storage) GetOneAccountById(id int) (acme_accounts.Account, error) {
	return store.getOneAccount(id, "")
}

// GetOneAccountByName returns an Account based on its unique name
func (store *Storage) GetOneAccountByName(name string) (acme_accounts.Account, error) {
	return store.getOneAccount(-1, name)
}

// getOneAccount returns an Account based on either its unique id or its unique name
func (store *Storage) getOneAccount(id int, name string) (acme_accounts.Account, error) {
	ctx, cancel := context.WithTimeout(context.Background(), store.Timeout)
	defer cancel()

	query := `
	SELECT
		aa.id, aa.name, aa.description, aa.status, aa.email, aa.accepted_tos, aa.is_staging,
		aa.created_at, aa.updated_at, aa.kid,
		pk.id, pk.name, pk.description, pk.algorithm, pk.pem, pk.api_key, pk.api_key_via_url,
		pk.created_at, pk.updated_at
	FROM
		acme_accounts aa
		LEFT JOIN private_keys pk on (aa.private_key_id = pk.id)
	WHERE aa.id = $1 OR aa.name = $2
	ORDER BY aa.id`

	row := store.Db.QueryRowContext(ctx, query, id, name)

	var oneAccount accountDb

	err := row.Scan(
		&oneAccount.id,
		&oneAccount.name,
		&oneAccount.description,
		&oneAccount.status,
		&oneAccount.email,
		&oneAccount.acceptedTos,
		&oneAccount.isStaging,
		&oneAccount.createdAt,
		&oneAccount.updatedAt,
		&oneAccount.kid,

		&oneAccount.accountKeyDb.id,
		&oneAccount.accountKeyDb.name,
		&oneAccount.accountKeyDb.description,
		&oneAccount.accountKeyDb.algorithmValue,
		&oneAccount.accountKeyDb.pem,
		&oneAccount.accountKeyDb.apiKey,
		&oneAccount.accountKeyDb.apiKeyViaUrl,
		&oneAccount.accountKeyDb.createdAt,
		&oneAccount.accountKeyDb.updatedAt,
	)

	if err != nil {
		// if no record exists
		if err == sql.ErrNoRows {
			err = storage.ErrNoRecord
		}
		return acme_accounts.Account{}, err
	}

	return oneAccount.toAccount(), nil
}
