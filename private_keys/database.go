package private_keys

import (
	"context"
)

func (privateKeysApp *PrivateKeysApp) dbGetAllPrivateKeys() ([]*privateKey, error) {
	ctx, cancel := context.WithTimeout(context.Background(), privateKeysApp.Timeout)
	defer cancel()

	query := `SELECT id, name, description, algorithm
	FROM private_keys ORDER BY id`

	rows, err := privateKeysApp.Database.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var allKeys []*privateKey
	for rows.Next() {
		var oneKey privateKeyDb
		err = rows.Scan(
			&oneKey.id,
			&oneKey.name,
			&oneKey.description,
			&oneKey.algorithmValue,
		)
		if err != nil {
			return nil, err
		}
		convertedKey, err := oneKey.privateKeyDbToPk()
		if err != nil {
			return nil, err
		}
		allKeys = append(allKeys, convertedKey)
	}

	return allKeys, nil
}

func (privateKeysApp *PrivateKeysApp) dbGetOnePrivateKey(id int) (*privateKey, error) {
	ctx, cancel := context.WithTimeout(context.Background(), privateKeysApp.Timeout)
	defer cancel()

	query := `SELECT id, name, description, algorithm, pem, api_key, created_at, updated_at
	FROM private_keys
	WHERE id = $1
	ORDER BY id`

	row := privateKeysApp.Database.QueryRowContext(ctx, query, id)

	var sqlKey privateKeyDb
	err := row.Scan(
		&sqlKey.id,
		&sqlKey.name,
		&sqlKey.description,
		&sqlKey.algorithmValue,
		&sqlKey.pem,
		&sqlKey.apiKey,
		&sqlKey.createdAt,
		&sqlKey.updatedAt,
	)

	if err != nil {
		return nil, err
	}

	convertedKey, err := sqlKey.privateKeyDbToPk()
	if err != nil {
		return nil, err
	}

	return convertedKey, nil
}

func (privateKeysApp *PrivateKeysApp) dbPutExistingPrivateKey(privateKey privateKeyDb) error {
	ctx, cancel := context.WithTimeout(context.Background(), privateKeysApp.Timeout)
	defer cancel()

	query := `
	UPDATE
		private_keys
	SET
		name = $1,
		description = $2,
		updated_at = $3
	WHERE
		id = $4`

	_, err := privateKeysApp.Database.ExecContext(ctx, query,
		privateKey.name,
		privateKey.description.String,
		privateKey.updatedAt,
		privateKey.id)
	if err != nil {
		return err
	}

	return nil
}

func (privateKeysApp *PrivateKeysApp) dbPostNewPrivateKey(privateKey privateKeyDb) error {
	ctx, cancel := context.WithTimeout(context.Background(), privateKeysApp.Timeout)
	defer cancel()

	query := `
	INSERT INTO private_keys (name, description, algorithm, pem, api_key, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := privateKeysApp.Database.ExecContext(ctx, query,
		privateKey.name,
		privateKey.description.String,
		privateKey.algorithmValue,
		privateKey.pem,
		privateKey.apiKey,
		privateKey.createdAt,
		privateKey.updatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}

// delete a private key from the database
func (privateKeysApp *PrivateKeysApp) dbDeletePrivateKey(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), privateKeysApp.Timeout)
	defer cancel()

	query := `
	DELETE FROM
		private_keys
	WHERE
		id = $1
	`

	// TODO: Ensure can't delete a key that is in use on an account or certificate

	_, err := privateKeysApp.Database.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	return nil
}
