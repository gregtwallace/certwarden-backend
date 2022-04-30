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
		var oneKey sqlPrivateKey
		err = rows.Scan(
			&oneKey.id,
			&oneKey.name,
			&oneKey.description,
			&oneKey.algorithmValue,
		)
		if err != nil {
			return nil, err
		}
		convertedKey, err := oneKey.sqlToPrivateKey()
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

	var sqlKey sqlPrivateKey
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

	convertedKey, err := sqlKey.sqlToPrivateKey()
	if err != nil {
		return nil, err
	}

	return convertedKey, nil
}