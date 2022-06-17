package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"legocerthub-backend/pkg/private_keys"
	"legocerthub-backend/pkg/utils"
	"time"
)

// a single private key, as database table fields
type keyDb struct {
	id             int
	name           sql.NullString
	description    sql.NullString
	algorithmValue sql.NullString
	pem            sql.NullString
	apiKey         string
	createdAt      int
	updatedAt      int
}

// KeyDbToKey translates the db object into the object the key service expects
func (keyDb *keyDb) keyDbToKey() private_keys.Key {
	return private_keys.Key{
		ID:          keyDb.id,
		Name:        keyDb.name.String,
		Description: keyDb.description.String,
		Algorithm:   utils.AlgorithmByValue(keyDb.algorithmValue.String),
		Pem:         keyDb.pem.String,
		ApiKey:      keyDb.apiKey,
		CreatedAt:   keyDb.createdAt,
		UpdatedAt:   keyDb.updatedAt,
	}
}

// payloadToDb translates a client payload into the db object
func payloadToDb(payload private_keys.KeyPayload) (keyDb, error) {
	var dbObj keyDb
	var err error

	// payload ID should never be missing at this point, regardless error if it somehow
	//  is to avoid nil pointer dereference
	if payload.ID == nil {
		err = errors.New("id missing in payload")
		return keyDb{}, err
	}
	dbObj.id = *payload.ID

	dbObj.name = stringToNullString(payload.Name)

	dbObj.description = stringToNullString(payload.Description)

	dbObj.algorithmValue = stringToNullString(payload.AlgorithmValue)

	dbObj.pem = stringToNullString(payload.PemContent)

	// CreatedAt is always populated but only sometimes used
	dbObj.createdAt = int(time.Now().Unix())

	dbObj.updatedAt = dbObj.createdAt

	return dbObj, nil
}

// dbGetAllPrivateKeys writes information about all private keys to json
func (storage Storage) GetAllKeys() ([]private_keys.Key, error) {
	ctx, cancel := context.WithTimeout(context.Background(), storage.Timeout)
	defer cancel()

	query := `SELECT id, name, description, algorithm
	FROM private_keys ORDER BY id`

	rows, err := storage.Db.QueryContext(ctx, query)
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

		convertedKey := oneKeyDb.keyDbToKey()

		allKeys = append(allKeys, convertedKey)
	}

	return allKeys, nil
}

// dbGetOneKey returns a key from the db based on unique id
func (storage Storage) GetOneKey(id int) (private_keys.Key, error) {
	ctx, cancel := context.WithTimeout(context.Background(), storage.Timeout)
	defer cancel()

	query := `SELECT id, name, description, algorithm, pem, api_key, created_at, updated_at
	FROM private_keys
	WHERE id = $1
	ORDER BY id`

	row := storage.Db.QueryRowContext(ctx, query, id)

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
		return private_keys.Key{}, err
	}

	convertedKey := oneKeyDb.keyDbToKey()

	return convertedKey, nil
}

// dbPutExistingKey sets an existing key equal to the PUT values (overwriting
//  old values)
func (storage *Storage) PutExistingKey(payload private_keys.KeyPayload) error {
	// load payload fields into db struct
	key, err := payloadToDb(payload)
	if err != nil {
		return err
	}

	// database action
	ctx, cancel := context.WithTimeout(context.Background(), storage.Timeout)
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

	_, err = storage.Db.ExecContext(ctx, query,
		key.name,
		key.description,
		key.updatedAt,
		key.id)
	if err != nil {
		return err
	}

	// TODO: Handle 0 rows updated.

	return nil
}

// dbPostNewKey creates a new key based on what was POSTed
func (storage *Storage) PostNewKey(payload private_keys.KeyPayload) error {
	// load payload fields into db struct
	key, err := payloadToDb(payload)
	if err != nil {
		return err
	}

	// generate api key
	key.apiKey, err = utils.GenerateApiKey()
	if err != nil {
		return err
	}

	// database action
	ctx, cancel := context.WithTimeout(context.Background(), storage.Timeout)
	defer cancel()

	query := `
	INSERT INTO private_keys (name, description, algorithm, pem, api_key, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err = storage.Db.ExecContext(ctx, query,
		key.name,
		key.description,
		key.algorithmValue,
		key.pem,
		key.apiKey,
		key.createdAt,
		key.updatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}

// delete a private key from the database
func (storage *Storage) DeleteKey(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), storage.Timeout)
	defer cancel()

	query := `
	DELETE FROM
		private_keys
	WHERE
		id = $1
	`

	// TODO: Ensure can't delete a key that is in use on an account or certificate

	result, err := storage.Db.ExecContext(ctx, query, id)
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

// dbGetAvailableKeys returns a slice of private keys that exist but are not already associated
//  with a known ACME account or certificate
func (storage *Storage) GetAvailableKeys() ([]private_keys.Key, error) {
	ctx, cancel := context.WithTimeout(context.Background(), storage.Timeout)
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

	rows, err := storage.Db.QueryContext(ctx, query)
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

		convertedKey := oneKey.keyDbToKey()

		availableKeys = append(availableKeys, convertedKey)
	}

	return availableKeys, nil
}
