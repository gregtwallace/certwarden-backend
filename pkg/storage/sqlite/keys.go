package sqlite

import (
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
