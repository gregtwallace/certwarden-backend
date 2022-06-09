package acme_accounts

import (
	"context"
	"database/sql"
	"errors"
	"legocerthub-backend/pkg/private_keys"
	"strconv"
)

// dbGetAllAccounts returns a slice of all of the acme accounts in the database
func (accountAppDb *AccountAppDb) getAllAccounts() ([]account, error) {
	ctx, cancel := context.WithTimeout(context.Background(), accountAppDb.Timeout)
	defer cancel()

	query := `SELECT aa.id, aa.name, aa.description, pk.id, pk.name, aa.status, aa.email, aa.is_staging 
	FROM
		acme_accounts aa
		LEFT JOIN private_keys pk on (aa.private_key_id = pk.id)
	ORDER BY aa.id`

	rows, err := accountAppDb.Database.QueryContext(ctx, query)
	if err != nil {
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
			return nil, err
		}

		convertedAccount, err := oneAccount.accountDbToAcc()
		if err != nil {
			return nil, err
		}

		allAccounts = append(allAccounts, convertedAccount)
	}

	return allAccounts, nil
}

// dbGetOneAccount returns an acmeAccount based on its unique id
func (accountAppDb *AccountAppDb) getOneAccount(id int) (account, error) {
	ctx, cancel := context.WithTimeout(context.Background(), accountAppDb.Timeout)
	defer cancel()

	query := `SELECT aa.id, aa.name, aa.description, pk.id, pk.name, aa.status, aa.email, aa.is_staging,
	aa.accepted_tos, aa.kid, aa.created_at, aa.updated_at
	FROM
		acme_accounts aa
		LEFT JOIN private_keys pk on (aa.private_key_id = pk.id)
	WHERE aa.id = $1
	ORDER BY aa.id`

	row := accountAppDb.Database.QueryRowContext(ctx, query, id)

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
func (accountAppDb *AccountAppDb) putExistingAccount(account accountDb) error {
	ctx, cancel := context.WithTimeout(context.Background(), accountAppDb.Timeout)
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

	_, err := accountAppDb.Database.ExecContext(ctx, query,
		account.name,
		account.description,
		account.email,
		account.acceptedTos,
		account.updatedAt,
		account.id)
	if err != nil {
		return err
	}

	// TODO: Handle 0 rows updated.
	return nil
}

// putLEAccountInfo populates an account with data that is returned by LE when
//  an account is POSTed to
func (accountAppDb *AccountAppDb) putLEAccountInfo(account accountDb) error {
	ctx, cancel := context.WithTimeout(context.Background(), accountAppDb.Timeout)
	defer cancel()

	query := `
	UPDATE
		acme_accounts
	SET
		status = $1,
		email = $2,
		created_at = $3,
		updated_at = $4,
		kid = $5
	WHERE
		id = $6`

	_, err := accountAppDb.Database.ExecContext(ctx, query,
		account.status,
		account.email,
		account.createdAt,
		account.updatedAt,
		account.kid,
		account.id)
	if err != nil {
		return err
	}

	// TODO: Handle 0 rows updated.
	return nil
}

// dbGetAvailableKeys returns a slice of private keys that exist but are not already associated
//  with a known ACME account or certificate
func (accountAppDb *AccountAppDb) getAvailableKeys() ([]private_keys.Key, error) {
	ctx, cancel := context.WithTimeout(context.Background(), accountAppDb.Timeout)
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

	rows, err := accountAppDb.Database.QueryContext(ctx, query)
	if err != nil {
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
			return nil, err
		}

		convertedKey := oneKey.KeyDbToKey()

		availableKeys = append(availableKeys, convertedKey)
	}

	return availableKeys, nil
}

func (accountAppDb *AccountAppDb) getAccountKeyPem(accountId string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), accountAppDb.Timeout)
	defer cancel()

	query := `SELECT pk.pem
	FROM
		acme_accounts aa
		LEFT JOIN private_keys pk on (aa.private_key_id = pk.id)
	WHERE
		aa.id = $1
	`

	row := accountAppDb.Database.QueryRowContext(ctx, query, accountId)
	var pem sql.NullString
	err := row.Scan(&pem)

	if err != nil {
		return "", err
	}

	return pem.String, nil
}

// postNewAccount inserts a new account into the db and returns the id of the new account
func (accoundAppDb *AccountAppDb) postNewAccount(account accountDb) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), accoundAppDb.Timeout)
	defer cancel()

	tx, err := accoundAppDb.Database.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}

	// insert the new account
	query := `
	INSERT INTO acme_accounts (name, description, private_key_id, status, email, accepted_tos, is_staging, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	result, err := tx.ExecContext(ctx, query,
		account.name,
		account.description,
		account.privateKeyId,
		account.status,
		account.email,
		account.acceptedTos,
		account.isStaging,
		account.createdAt,
		account.updatedAt,
	)
	if err != nil {
		tx.Rollback()
		return "", err
	}

	// id of the new account
	id, err := result.LastInsertId()
	if err != nil {
		tx.Rollback()
		return "", err
	}
	idStr := strconv.Itoa(int(id))

	// verify the new account does not have a cert that uses the same key
	query = `
		SELECT private_key_id
		FROM
		  certificates
		WHERE
			private_key_id = $1
	`

	row := tx.QueryRowContext(ctx, query, account.privateKeyId)

	var exists bool
	err = row.Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		tx.Rollback()
		return "", err
	} else if exists {
		tx.Rollback()
		return "", errors.New("private key in use by certificate")
	}

	err = tx.Commit()
	if err != nil {
		return "", err
	}

	return idStr, nil
}

func (accountAppDb *AccountAppDb) getAccountKid(accountId string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), accountAppDb.Timeout)
	defer cancel()

	query := `
		SELECT kid
		FROM
		  acme_accounts
		WHERE
			id = $1
	`

	row := accountAppDb.Database.QueryRowContext(ctx, query, accountId)
	var kid sql.NullString
	err := row.Scan(&kid)

	if err != nil {
		return "", err
	}

	return kid.String, nil
}

// putExistingAccountNameDesc only updates the name and desc in the database
func (db AccountAppDb) putExistingAccountNameDesc(account accountDb) error {
	ctx, cancel := context.WithTimeout(context.Background(), db.Timeout)
	defer cancel()

	query := `
	UPDATE
		acme_accounts
	SET
		name = $1,
		description = $2
	WHERE
		id = $3`

	_, err := db.Database.ExecContext(ctx, query,
		account.name,
		account.description,
		account.id)
	if err != nil {
		return err
	}

	// TODO: Handle 0 rows updated.
	return nil
}
