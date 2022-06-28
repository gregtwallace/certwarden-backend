package sqlite

import (
	"context"
	"database/sql"
	"legocerthub-backend/pkg/domain/acme_accounts"
	"legocerthub-backend/pkg/domain/private_keys"
)

// accountDbToAcc turns the database representation of an Account into an Account
func (accountDb *accountDb) accountDbToAcc() (acme_accounts.Account, error) {
	var privateKey = new(private_keys.Key)
	var err error

	// convert embedded private key db
	if accountDb.privateKey != nil {
		*privateKey, err = accountDb.privateKey.keyDbToKey()
		if err != nil {
			return acme_accounts.Account{}, err
		}
	} else {
		privateKey = nil
	}

	return acme_accounts.Account{
		ID:          accountDb.id,
		Name:        accountDb.name,
		Description: accountDb.description.String,
		PrivateKey:  privateKey,
		Status:      accountDb.status.String,
		Email:       accountDb.email.String,
		AcceptedTos: accountDb.acceptedTos.Bool,
		IsStaging:   accountDb.isStaging.Bool,
		CreatedAt:   accountDb.createdAt,
		UpdatedAt:   accountDb.updatedAt,
		Kid:         accountDb.kid.String,
	}, nil
}

// GetAllAccounts returns a slice of all of the Accounts in the database
func (storage *Storage) GetAllAccounts() ([]acme_accounts.Account, error) {
	ctx, cancel := context.WithTimeout(context.Background(), storage.Timeout)
	defer cancel()

	query := `SELECT aa.id, aa.name, aa.description, pk.id, pk.name, pk.description,
	aa.status, aa.email, aa.is_staging 
	FROM
		acme_accounts aa
		LEFT JOIN private_keys pk on (aa.private_key_id = pk.id)
	ORDER BY aa.id`

	rows, err := storage.Db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var allAccounts []acme_accounts.Account
	for rows.Next() {
		var oneAccount accountDb
		// initialize keyDb pointer (or nil deref)
		oneAccount.privateKey = new(keyDb)
		err = rows.Scan(
			&oneAccount.id,
			&oneAccount.name,
			&oneAccount.description,
			&oneAccount.privateKey.id,
			&oneAccount.privateKey.name,
			&oneAccount.privateKey.description,
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

// GetOneAccountById returns an Account based on its unique id
func (storage *Storage) GetOneAccountById(id int) (acme_accounts.Account, error) {
	return storage.getOneAccount(id, "")
}

// GetOneAccountByName returns an Account based on its unique name
func (storage *Storage) GetOneAccountByName(name string) (acme_accounts.Account, error) {
	return storage.getOneAccount(-1, name)
}

// getOneAccount returns an Account based on either its unique id or its unique name
func (storage *Storage) getOneAccount(id int, name string) (acme_accounts.Account, error) {
	ctx, cancel := context.WithTimeout(context.Background(), storage.Timeout)
	defer cancel()

	query := `SELECT aa.id, aa.name, aa.description, pk.id, pk.name, pk.description, pk.algorithm, pk.pem, 
	aa.status, aa.email, aa.is_staging, aa.accepted_tos, aa.kid, aa.created_at, aa.updated_at
	FROM
		acme_accounts aa
		LEFT JOIN private_keys pk on (aa.private_key_id = pk.id)
	WHERE aa.id = $1 OR aa.name = $2
	ORDER BY aa.id`

	row := storage.Db.QueryRowContext(ctx, query, id, name)

	var oneAccount accountDb
	// initialize keyDb pointer (or nil deref)
	oneAccount.privateKey = new(keyDb)

	err := row.Scan(
		&oneAccount.id,
		&oneAccount.name,
		&oneAccount.description,
		&oneAccount.privateKey.id,
		&oneAccount.privateKey.name,
		&oneAccount.privateKey.description,
		&oneAccount.privateKey.algorithmValue,
		&oneAccount.privateKey.pem,
		&oneAccount.status,
		&oneAccount.email,
		&oneAccount.isStaging,
		&oneAccount.acceptedTos,
		&oneAccount.kid,
		&oneAccount.createdAt,
		&oneAccount.updatedAt,
	)
	if err != nil {
		return acme_accounts.Account{}, err
	}

	convertedAccount, err := oneAccount.accountDbToAcc()
	if err != nil {
		return acme_accounts.Account{}, err
	}

	return convertedAccount, nil
}

func (storage *Storage) GetAccountPem(id int) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), storage.Timeout)
	defer cancel()

	query := `SELECT pk.pem
	FROM
		acme_accounts aa
		LEFT JOIN private_keys pk on (aa.private_key_id = pk.id)
	WHERE
		aa.id = $1
	`

	row := storage.Db.QueryRowContext(ctx, query, id)
	var pem sql.NullString
	err := row.Scan(&pem)

	if err != nil {
		return "", err
	}

	return pem.String, nil
}

func (storage *Storage) GetAccountKid(id int) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), storage.Timeout)
	defer cancel()

	query := `
		SELECT kid
		FROM
		  acme_accounts
		WHERE
			id = $1
	`

	row := storage.Db.QueryRowContext(ctx, query, id)
	var kid sql.NullString
	err := row.Scan(&kid)

	if err != nil {
		return "", err
	}

	return kid.String, nil
}
