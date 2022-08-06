package sqlite

import (
	"context"
	"database/sql"
	"legocerthub-backend/pkg/domain/private_keys"
	"legocerthub-backend/pkg/domain/private_keys/key_crypto"
	"legocerthub-backend/pkg/storage"
)

// KeyDbToKey translates the db object into the object the key service expects
func (keyDb *keyDb) keyDbToKey() (private_keys.Key, error) {
	var algorithm = new(key_crypto.Algorithm)
	var err error

	// if there is an algorithm value, specify the algorithm
	if keyDb.algorithmValue.Valid {
		*algorithm, err = key_crypto.AlgorithmByValue(keyDb.algorithmValue.String)
		if err != nil {
			return private_keys.Key{}, err
		}
	} else {
		algorithm = nil
	}

	return private_keys.Key{
		ID:          nullInt32ToInt(keyDb.id),
		Name:        nullStringToString(keyDb.name),
		Description: nullStringToString(keyDb.description),
		Algorithm:   algorithm,
		Pem:         nullStringToString(keyDb.pem),
		ApiKey:      nullStringToString(keyDb.apiKey),
		CreatedAt:   nullInt32ToInt(keyDb.createdAt),
		UpdatedAt:   nullInt32ToInt(keyDb.updatedAt),
	}, nil
}

// dbGetAllPrivateKeys writes information about all private keys to json
func (store Storage) GetAllKeys() ([]private_keys.Key, error) {
	ctx, cancel := context.WithTimeout(context.Background(), store.Timeout)
	defer cancel()

	query := `SELECT id, name, description, algorithm
	FROM private_keys ORDER BY name`

	rows, err := store.Db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var allKeys []private_keys.Key
	for rows.Next() {
		var oneKey keyDb
		err = rows.Scan(
			&oneKey.id,
			&oneKey.name,
			&oneKey.description,
			&oneKey.algorithmValue,
		)
		if err != nil {
			return nil, err
		}

		convertedKey, err := oneKey.keyDbToKey()
		if err != nil {
			return nil, err
		}

		allKeys = append(allKeys, convertedKey)
	}

	return allKeys, nil
}

// GetOneKeyById returns a key based on its unique id
func (store *Storage) GetOneKeyById(id int) (private_keys.Key, error) {
	return store.getOneKey(id, "")
}

// GetOneKeyByName returns a key based on its unique name
func (store *Storage) GetOneKeyByName(name string) (private_keys.Key, error) {
	return store.getOneKey(-1, name)
}

// dbGetOneKey returns a key from the db based on unique id or unique name
func (store Storage) getOneKey(id int, name string) (private_keys.Key, error) {
	ctx, cancel := context.WithTimeout(context.Background(), store.Timeout)
	defer cancel()

	query := `SELECT id, name, description, algorithm, pem, api_key, created_at, updated_at
	FROM private_keys
	WHERE id = $1 OR name = $2
	`

	row := store.Db.QueryRowContext(ctx, query, id, name)

	var oneKeyDb keyDb
	err := row.Scan(
		&oneKeyDb.id,
		&oneKeyDb.name,
		&oneKeyDb.description,
		&oneKeyDb.algorithmValue,
		&oneKeyDb.pem,
		&oneKeyDb.apiKey,
		&oneKeyDb.createdAt,
		&oneKeyDb.updatedAt,
	)

	if err != nil {
		// if no record exists
		if err == sql.ErrNoRows {
			err = storage.ErrNoRecord
		}
		return private_keys.Key{}, err
	}

	convertedKey, err := oneKeyDb.keyDbToKey()
	if err != nil {
		return private_keys.Key{}, err
	}

	return convertedKey, nil
}

// GetAvailableKeys returns a slice of private keys that exist but are not already associated
// with a known ACME account or certificate
func (store *Storage) GetAvailableKeys() ([]private_keys.Key, error) {
	ctx, cancel := context.WithTimeout(context.Background(), store.Timeout)
	defer cancel()

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
			AND
			NOT EXISTS(
				SELECT
					c.private_key_id
				FROM
					certificates c
				WHERE
					pk.id = c.private_key_id
			)
			AND
			NOT EXISTS(
				SELECT
					ao.certificate_id
				FROM
					acme_orders ao
				WHERE 
					status = "valid"
					AND
					ao.certificate_id not null
				GROUP BY
					ao.certificate_id
				HAVING
					MAX(ao.valid_to)
					AND
					pk.id = ao.finalized_key_id
			)
		ORDER BY name
	`

	rows, err := store.Db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var availableKeys []private_keys.Key
	for rows.Next() {
		var oneKey keyDb

		err = rows.Scan(
			&oneKey.id,
			&oneKey.name,
			&oneKey.description,
			&oneKey.algorithmValue,
		)
		if err != nil {
			return nil, err
		}

		convertedKey, err := oneKey.keyDbToKey()
		if err != nil {
			return nil, err
		}

		availableKeys = append(availableKeys, convertedKey)
	}

	return availableKeys, nil
}
