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
			&oneKey.ID,
			&oneKey.Name,
			&oneKey.Description,
			&oneKey.Algorithm,
		)
		if err != nil {
			return nil, err
		}
		allKeys = append(allKeys, oneKey.sqlToPrivateKey())
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
		&sqlKey.ID,
		&sqlKey.Name,
		&sqlKey.Description,
		&sqlKey.Algorithm,
		&sqlKey.Pem,
		&sqlKey.ApiKey,
		&sqlKey.CreatedAt,
		&sqlKey.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return sqlKey.sqlToPrivateKey(), nil
}
