package acme_accounts

import (
	"context"
	"legocerthub-backend/private_keys"
)

// dbGetAllAcmeAccounts returns a slice of all of the acme accounts in the database
func (acmeAccountsApp *AcmeAccountsApp) dbGetAllAcmeAccounts() ([]*acmeAccount, error) {
	ctx, cancel := context.WithTimeout(context.Background(), acmeAccountsApp.Timeout)
	defer cancel()

	query := `SELECT aa.id, aa.name, aa.description, pk.id, pk.name, aa.status, aa.email, aa.is_staging 
	FROM
		acme_accounts aa
		LEFT JOIN private_keys pk on (aa.private_key_id = pk.id)
	ORDER BY aa.id`

	rows, err := acmeAccountsApp.Database.QueryContext(ctx, query)
	if err != nil {
		acmeAccountsApp.Logger.Println(err)
		return nil, err
	}
	defer rows.Close()

	var allAccounts []*acmeAccount
	for rows.Next() {
		var oneAccount acmeAccountDb
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
			acmeAccountsApp.Logger.Println(err)
			return nil, err
		}

		convertedAccount, err := oneAccount.acmeAccountDbToAcc()
		if err != nil {
			acmeAccountsApp.Logger.Println(err)
			return nil, err
		}

		allAccounts = append(allAccounts, convertedAccount)
	}

	return allAccounts, nil
}

// dbGetOneAcmeAccount returns an acmeAccount based on its unique id
func (acmeAccountsApp *AcmeAccountsApp) dbGetOneAcmeAccount(id int) (*acmeAccount, error) {
	ctx, cancel := context.WithTimeout(context.Background(), acmeAccountsApp.Timeout)
	defer cancel()

	query := `SELECT aa.id, aa.name, aa.description, pk.id, pk.name, aa.status, aa.email, aa.is_staging,
	aa.accepted_tos, aa.kid, aa.created_at, aa.updated_at
	FROM
		acme_accounts aa
		LEFT JOIN private_keys pk on (aa.private_key_id = pk.id)
	WHERE aa.id = $1
	ORDER BY aa.id`

	row := acmeAccountsApp.Database.QueryRowContext(ctx, query, id)

	var oneAccount acmeAccountDb
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
		acmeAccountsApp.Logger.Println(err)
		return nil, err
	}

	convertedAccount, err := oneAccount.acmeAccountDbToAcc()
	if err != nil {
		return nil, err
	}

	return convertedAccount, nil
}

// dbPutExistingAcmeAccount overwrites an existing acme account using specified data
// certain fields cannot be updated, per rfc8555
func (acmeAccountsApp *AcmeAccountsApp) dbPutExistingAcmeAccount(acmeAccount acmeAccountDb) error {
	ctx, cancel := context.WithTimeout(context.Background(), acmeAccountsApp.Timeout)
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

	_, err := acmeAccountsApp.Database.ExecContext(ctx, query,
		acmeAccount.name,
		acmeAccount.description,
		acmeAccount.email,
		acmeAccount.acceptedTos,
		acmeAccount.updatedAt,
		acmeAccount.id)
	if err != nil {
		return err
	}

	// TODO: Handle 0 rows updated.
	return nil

}

// dbGetAvailableKeys returns a slice of private keys that exist but are not already associated
//  with a known ACME account or certificate
func (acmeAccountsApp *AcmeAccountsApp) dbGetAvailableKeys() ([]private_keys.PrivateKey, error) {
	ctx, cancel := context.WithTimeout(context.Background(), acmeAccountsApp.Timeout)
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

	rows, err := acmeAccountsApp.Database.QueryContext(ctx, query)
	if err != nil {
		acmeAccountsApp.Logger.Println(err)
		return nil, err
	}
	defer rows.Close()

	var availableKeys []private_keys.PrivateKey
	for rows.Next() {
		var oneKey private_keys.PrivateKeyDb

		err = rows.Scan(
			&oneKey.ID,
			&oneKey.Name,
			&oneKey.Description,
			&oneKey.AlgorithmValue,
		)
		if err != nil {
			acmeAccountsApp.Logger.Println(err)
			return nil, err
		}

		convertedKey := oneKey.PrivateKeyDbToPk()

		availableKeys = append(availableKeys, *convertedKey)
	}

	return availableKeys, nil
}

// query := `SELECT aa.id, aa.name, aa.description, pk.id, pk.name, aa.status, aa.email, aa.is_staging,
// aa.accepted_tos, aa.kid, aa.created_at, aa.updated_at
// FROM
// 	acme_accounts aa
// 	LEFT JOIN private_keys pk on (aa.private_key_id = pk.id)
// WHERE aa.id = $1
// ORDER BY aa.id`

// ID          int       `json:"id"`
// Name        string    `json:"name"`
// Description string    `json:"description"`
// Algorithm   algorithm `json:"algorithm"`
