package sqlite

import (
	"context"
	"database/sql"
	"legocerthub-backend/pkg/domain/private_keys"
	"legocerthub-backend/pkg/storage"
)

// GetAllKeys returns a slice of all Keys in the db
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
		var oneKeyDb keyDb
		err = rows.Scan(
			&oneKeyDb.id,
			&oneKeyDb.name,
			&oneKeyDb.description,
			&oneKeyDb.algorithmValue,
		)
		if err != nil {
			return nil, err
		}

		convertedKey := oneKeyDb.toKey()

		allKeys = append(allKeys, convertedKey)
	}

	return allKeys, nil
}

// GetOneKeyById returns a KeyExtended based on unique id
func (store *Storage) GetOneKeyById(id int, withPem bool) (private_keys.KeyExtended, error) {
	return store.getOneKey(id, "", withPem)
}

// GetOneKeyByName returns a KeyExtended based on unique name
func (store *Storage) GetOneKeyByName(name string, withPem bool) (private_keys.KeyExtended, error) {
	return store.getOneKey(-1, name, withPem)
}

// dbGetOneKey returns a KeyExtended based on unique id or unique name
func (store Storage) getOneKey(id int, name string, withPem bool) (private_keys.KeyExtended, error) {
	ctx, cancel := context.WithTimeout(context.Background(), store.Timeout)
	defer cancel()

	query := `
	SELECT
		id, name, description, algorithm, pem, api_key, api_key_via_url, created_at, updated_at
	FROM
		private_keys
	WHERE
		id = $1
		OR
		name = $2
	`

	row := store.Db.QueryRowContext(ctx, query, id, name)

	var oneKeyDb keyDbExtended
	err := row.Scan(
		&oneKeyDb.id,
		&oneKeyDb.name,
		&oneKeyDb.description,
		&oneKeyDb.algorithmValue,
		&oneKeyDb.pem,
		&oneKeyDb.apiKey,
		&oneKeyDb.apiKeyViaUrl,
		&oneKeyDb.createdAt,
		&oneKeyDb.updatedAt,
	)

	if err != nil {
		// if no record exists
		if err == sql.ErrNoRows {
			err = storage.ErrNoRecord
		}
		return private_keys.KeyExtended{}, err
	}

	// convert to KeyExtended
	oneKeyExt := oneKeyDb.toKeyExtended(withPem)
	if err != nil {
		return private_keys.KeyExtended{}, err
	}

	return oneKeyExt, nil
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
		var oneKeyDb keyDb

		err = rows.Scan(
			&oneKeyDb.id,
			&oneKeyDb.name,
			&oneKeyDb.description,
			&oneKeyDb.algorithmValue,
		)
		if err != nil {
			return nil, err
		}

		oneKeyExt := oneKeyDb.toKey()

		availableKeys = append(availableKeys, oneKeyExt)
	}

	return availableKeys, nil
}

// // GetKeyPemById returns the pem for the specified key id
// func (store *Storage) GetKeyPemById(id int) (pem string, err error) {
// 	return store.getKeyPem(id, "")
// }

// GetKeyPemByName returns the pem for the specified key name
func (store *Storage) GetKeyPemByName(name string) (pem string, err error) {
	return store.getKeyPem(-1, name)
}

// dbGetOneKey returns a key from the db based on unique id or unique name
func (store Storage) getKeyPem(id int, name string) (pem string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), store.Timeout)
	defer cancel()

	query := `
	SELECT
		pem
	FROM
		private_keys
	WHERE
		id = $1
		OR
		name = $2
	`

	// query
	row := store.Db.QueryRowContext(ctx, query,
		id,
		name,
	)

	// scan
	err = row.Scan(&pem)
	if err != nil {
		// if no record exists
		if err == sql.ErrNoRows {
			err = storage.ErrNoRecord
		}
		return "", err
	}

	return pem, nil
}
