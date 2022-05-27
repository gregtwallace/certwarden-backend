package private_keys

import (
	"context"
	"errors"
)

func (privateKeysApp *PrivateKeysApp) dbGetAllPrivateKeys() ([]*PrivateKey, error) {
	ctx, cancel := context.WithTimeout(context.Background(), privateKeysApp.Timeout)
	defer cancel()

	query := `SELECT id, name, description, algorithm
	FROM private_keys ORDER BY id`

	rows, err := privateKeysApp.Database.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var allKeys []*PrivateKey
	for rows.Next() {
		var oneKey PrivateKeyDb
		err = rows.Scan(
			&oneKey.ID,
			&oneKey.Name,
			&oneKey.Description,
			&oneKey.AlgorithmValue,
		)
		if err != nil {
			return nil, err
		}

		convertedKey := oneKey.PrivateKeyDbToPk()

		allKeys = append(allKeys, convertedKey)
	}

	return allKeys, nil
}

func (privateKeysApp *PrivateKeysApp) dbGetOnePrivateKey(id int) (*PrivateKey, error) {
	ctx, cancel := context.WithTimeout(context.Background(), privateKeysApp.Timeout)
	defer cancel()

	query := `SELECT id, name, description, algorithm, pem, api_key, created_at, updated_at
	FROM private_keys
	WHERE id = $1
	ORDER BY id`

	row := privateKeysApp.Database.QueryRowContext(ctx, query, id)

	var sqlKey PrivateKeyDb
	err := row.Scan(
		&sqlKey.ID,
		&sqlKey.Name,
		&sqlKey.Description,
		&sqlKey.AlgorithmValue,
		&sqlKey.Pem,
		&sqlKey.ApiKey,
		&sqlKey.CreatedAt,
		&sqlKey.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	convertedKey := sqlKey.PrivateKeyDbToPk()

	return convertedKey, nil
}

func (privateKeysApp *PrivateKeysApp) dbPutExistingPrivateKey(privateKey PrivateKeyDb) error {
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
		privateKey.Name,
		privateKey.Description,
		privateKey.UpdatedAt,
		privateKey.ID)
	if err != nil {
		return err
	}

	// TODO: Handle 0 rows updated.

	return nil
}

func (privateKeysApp *PrivateKeysApp) dbPostNewPrivateKey(privateKey PrivateKeyDb) error {
	ctx, cancel := context.WithTimeout(context.Background(), privateKeysApp.Timeout)
	defer cancel()

	query := `
	INSERT INTO private_keys (name, description, algorithm, pem, api_key, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := privateKeysApp.Database.ExecContext(ctx, query,
		privateKey.Name,
		privateKey.Description,
		privateKey.AlgorithmValue,
		privateKey.Pem,
		privateKey.ApiKey,
		privateKey.CreatedAt,
		privateKey.UpdatedAt,
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

	result, err := privateKeysApp.Database.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	resultRows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if resultRows == 0 {
		return errors.New("privatekeys: Delete: failed to db delete -- 0 rows changed")
	}

	return nil
}
