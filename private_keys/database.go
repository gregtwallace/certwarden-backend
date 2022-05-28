package private_keys

import (
	"context"
	"errors"
)

// dbGetAllPrivateKeys writes information about all private keys to json
func (keysApp *KeysApp) dbGetAllKeys() ([]Key, error) {
	ctx, cancel := context.WithTimeout(context.Background(), keysApp.DB.Timeout)
	defer cancel()

	query := `SELECT id, name, description, algorithm
	FROM private_keys ORDER BY id`

	rows, err := keysApp.DB.Database.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var allKeys []Key
	for rows.Next() {
		var oneKeyDb KeyDb
		err = rows.Scan(
			&oneKeyDb.ID,
			&oneKeyDb.Name,
			&oneKeyDb.Description,
			&oneKeyDb.AlgorithmValue,
		)
		if err != nil {
			return nil, err
		}

		convertedKey := oneKeyDb.KeyDbToKey()

		allKeys = append(allKeys, convertedKey)
	}

	return allKeys, nil
}

// dbGetOneKey returns a key from the db based on unique id
func (keysApp *KeysApp) dbGetOneKey(id int) (Key, error) {
	ctx, cancel := context.WithTimeout(context.Background(), keysApp.DB.Timeout)
	defer cancel()

	query := `SELECT id, name, description, algorithm, pem, api_key, created_at, updated_at
	FROM private_keys
	WHERE id = $1
	ORDER BY id`

	row := keysApp.DB.Database.QueryRowContext(ctx, query, id)

	var oneKeyDb KeyDb
	err := row.Scan(
		&oneKeyDb.ID,
		&oneKeyDb.Name,
		&oneKeyDb.Description,
		&oneKeyDb.AlgorithmValue,
		&oneKeyDb.Pem,
		&oneKeyDb.ApiKey,
		&oneKeyDb.CreatedAt,
		&oneKeyDb.UpdatedAt,
	)

	if err != nil {
		return Key{}, err
	}

	convertedKey := oneKeyDb.KeyDbToKey()

	return convertedKey, nil
}

// dbPutExistingKey sets an existing key equal to the PUT values (overwriting
//  old values)
func (keysApp *KeysApp) dbPutExistingKey(keyDb KeyDb) error {
	ctx, cancel := context.WithTimeout(context.Background(), keysApp.DB.Timeout)
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

	_, err := keysApp.DB.Database.ExecContext(ctx, query,
		keyDb.Name,
		keyDb.Description,
		keyDb.UpdatedAt,
		keyDb.ID)
	if err != nil {
		return err
	}

	// TODO: Handle 0 rows updated.

	return nil
}

// dbPostNewKey creates a new key based on what was POSTed
func (keysApp *KeysApp) dbPostNewKey(keyDb KeyDb) error {
	ctx, cancel := context.WithTimeout(context.Background(), keysApp.DB.Timeout)
	defer cancel()

	query := `
	INSERT INTO private_keys (name, description, algorithm, pem, api_key, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := keysApp.DB.Database.ExecContext(ctx, query,
		keyDb.Name,
		keyDb.Description,
		keyDb.AlgorithmValue,
		keyDb.Pem,
		keyDb.ApiKey,
		keyDb.CreatedAt,
		keyDb.UpdatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}

// delete a private key from the database
func (keysApp *KeysApp) dbDeleteKey(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), keysApp.DB.Timeout)
	defer cancel()

	query := `
	DELETE FROM
		private_keys
	WHERE
		id = $1
	`

	// TODO: Ensure can't delete a key that is in use on an account or certificate

	result, err := keysApp.DB.Database.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	resultRows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if resultRows == 0 {
		return errors.New("keys: Delete: failed to db delete -- 0 rows changed")
	}

	return nil
}
