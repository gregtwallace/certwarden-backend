package acme_accounts

import (
	"context"
	"legocerthub-backend/private_keys"
)

// dbGetAllAccounts returns a slice of all of the acme accounts in the database
func (accountsApp *AccountsApp) dbGetAllAccounts() ([]account, error) {
	ctx, cancel := context.WithTimeout(context.Background(), accountsApp.DB.Timeout)
	defer cancel()

	query := `SELECT aa.id, aa.name, aa.description, pk.id, pk.name, aa.status, aa.email, aa.is_staging 
	FROM
		acme_accounts aa
		LEFT JOIN private_keys pk on (aa.private_key_id = pk.id)
	ORDER BY aa.id`

	rows, err := accountsApp.DB.Database.QueryContext(ctx, query)
	if err != nil {
		accountsApp.Logger.Println(err)
		return nil, err
	}
	defer rows.Close()

	var allAccounts []account
	for rows.Next() {
		var oneAccount accountDb
		err = rows.Scan(
			&oneAccount.id,
			&oneAccount.name,
			&oneAccount.description,
			&oneAccount.privateKeyId,
			&oneAccount.privateKeyName,
			&oneAccount.status,
			&oneAccount.email,
			&oneAccount.isStaging,
		)
		if err != nil {
			accountsApp.Logger.Println(err)
			return nil, err
		}

		convertedAccount, err := oneAccount.accountDbToAcc()
		if err != nil {
			accountsApp.Logger.Println(err)
			return nil, err
		}

		allAccounts = append(allAccounts, convertedAccount)
	}

	return allAccounts, nil
}

// dbGetOneAccount returns an acmeAccount based on its unique id
func (accountsApp *AccountsApp) dbGetOneAccount(id int) (account, error) {
	ctx, cancel := context.WithTimeout(context.Background(), accountsApp.DB.Timeout)
	defer cancel()

	query := `SELECT aa.id, aa.name, aa.description, pk.id, pk.name, aa.status, aa.email, aa.is_staging,
	aa.accepted_tos, aa.kid, aa.created_at, aa.updated_at
	FROM
		acme_accounts aa
		LEFT JOIN private_keys pk on (aa.private_key_id = pk.id)
	WHERE aa.id = $1
	ORDER BY aa.id`

	row := accountsApp.DB.Database.QueryRowContext(ctx, query, id)

	var oneAccount accountDb
	err := row.Scan(
		&oneAccount.id,
		&oneAccount.name,
		&oneAccount.description,
		&oneAccount.privateKeyId,
		&oneAccount.privateKeyName,
		&oneAccount.status,
		&oneAccount.email,
		&oneAccount.isStaging,
		&oneAccount.acceptedTos,
		&oneAccount.kid,
		&oneAccount.createdAt,
		&oneAccount.updatedAt,
	)
	if err != nil {
		accountsApp.Logger.Println(err)
		return account{}, err
	}

	convertedAccount, err := oneAccount.accountDbToAcc()
	if err != nil {
		return account{}, err
	}

	return convertedAccount, nil
}

// dbPutExistingAccount overwrites an existing acme account using specified data
// certain fields cannot be updated, per rfc8555
func (accountsApp *AccountsApp) dbPutExistingAccount(accountDb accountDb) error {
	ctx, cancel := context.WithTimeout(context.Background(), accountsApp.DB.Timeout)
	defer cancel()

	query := `
	UPDATE
		acme_accounts
	SET
		name = $1,
		description = $2,
		email = $3,
		accepted_tos = case when $4 is null then accepted_tos else $4 end,
		updated_at = $5
	WHERE
		id = $6`

	_, err := accountsApp.DB.Database.ExecContext(ctx, query,
		accountDb.name,
		accountDb.description,
		accountDb.email,
		accountDb.acceptedTos,
		accountDb.updatedAt,
		accountDb.id)
	if err != nil {
		return err
	}

	// TODO: Handle 0 rows updated.
	return nil

}

// dbGetAvailableKeys returns a slice of private keys that exist but are not already associated
//  with a known ACME account or certificate
func (accountsApp *AccountsApp) dbGetAvailableKeys() ([]private_keys.Key, error) {
	ctx, cancel := context.WithTimeout(context.Background(), accountsApp.DB.Timeout)
	defer cancel()

	// TODO - Once certs are added, need to check that table as well for keys in use
	query := `
		SELECT pk.id, pk.name, pk.description, pk.algorithm
		FROM
		  private_keys pk
		WHERE
			NOT EXISTS(
				SELECT
					aa.private_key_id
				FROM
					acme_accounts aa
				WHERE
					pk.id = aa.private_key_id
			)
	`

	rows, err := accountsApp.DB.Database.QueryContext(ctx, query)
	if err != nil {
		accountsApp.Logger.Println(err)
		return nil, err
	}
	defer rows.Close()

	var availableKeys []private_keys.Key
	for rows.Next() {
		var oneKey private_keys.KeyDb

		err = rows.Scan(
			&oneKey.ID,
			&oneKey.Name,
			&oneKey.Description,
			&oneKey.AlgorithmValue,
		)
		if err != nil {
			accountsApp.Logger.Println(err)
			return nil, err
		}

		convertedKey := oneKey.KeyDbToKey()

		availableKeys = append(availableKeys, convertedKey)
	}

	return availableKeys, nil
}
