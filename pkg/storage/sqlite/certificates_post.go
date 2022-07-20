package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"legocerthub-backend/pkg/domain/certificates"
	"legocerthub-backend/pkg/utils"
)

// accountPayloadToDb turns the client payload into a db object
func newCertPayloadToDb(payload certificates.NewPayload) (certDb certificateDb, err error) {
	// nil check mandatory components
	if payload.Name == nil ||
		payload.PrivateKeyID == nil ||
		payload.AcmeAccountID == nil ||
		payload.ChallengeMethodValue == nil ||
		payload.Subject == nil {

		return certificateDb{}, errors.New("missing mandatory payload component")
	}

	// initialize to avoid nil pointer
	certDb.privateKey = new(keyDb)
	certDb.acmeAccount = new(accountDb)

	certDb.name = stringToNullString(payload.Name)
	certDb.description = stringToNullString(payload.Description)
	certDb.privateKey.id = intToNullInt32(payload.PrivateKeyID)
	certDb.acmeAccount.id = intToNullInt32(payload.AcmeAccountID)
	certDb.challengeMethodValue = stringToNullString(payload.ChallengeMethodValue)
	certDb.subject = stringToNullString(payload.Subject)
	certDb.subjectAltNames = sliceToCommaNullString(payload.SubjectAltNames)

	// csr
	certDb.commonName = stringToNullString(payload.CommonName)
	certDb.organization = stringToNullString(payload.Organization)
	certDb.organizationalUnit = stringToNullString(payload.OrganizationalUnit)
	certDb.country = stringToNullString(payload.Country)
	certDb.state = stringToNullString(payload.State)
	certDb.city = stringToNullString(payload.City)

	// not in payload
	certDb.createdAt = timeNow()
	certDb.updatedAt = certDb.createdAt

	apiKey, err := utils.GenerateApiKey()
	certDb.apiKey = stringToNullString(&apiKey)
	if err != nil {
		return certificateDb{}, err
	}

	return certDb, nil
}

// PostNewAccount inserts a new cert into the db
func (store *Storage) PostNewCert(payload certificates.NewPayload) (id int, err error) {
	// Load payload into db obj
	certDb, err := newCertPayloadToDb(payload)
	if err != nil {
		return -2, err
	}

	// database update
	ctx, cancel := context.WithTimeout(context.Background(), store.Timeout)
	defer cancel()

	tx, err := store.Db.BeginTx(ctx, nil)
	if err != nil {
		return -2, err
	}
	defer tx.Rollback()

	// insert the new cert
	query := `
	INSERT INTO certificates (name, description, private_key_id, acme_account_id, challenge_method, subject, subject_alts, 
		csr_com_name, csr_org, csr_ou, csr_country, csr_state, csr_city, created_at, updated_at, api_key)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	RETURNING id
	`

	err = tx.QueryRowContext(ctx, query,
		certDb.name,
		certDb.description,
		certDb.privateKey.id,
		certDb.acmeAccount.id,
		certDb.challengeMethodValue,
		certDb.subject,
		certDb.subjectAltNames,
		certDb.commonName,
		certDb.organization,
		certDb.organizationalUnit,
		certDb.country,
		certDb.state,
		certDb.city,
		certDb.createdAt,
		certDb.updatedAt,
		certDb.apiKey,
	).Scan(&id)

	if err != nil {
		return -2, err
	}

	// table already enforces unique private_key_id, so no need to check against other certs
	// however, verify there is not an acme account with the same key
	query = `
		SELECT private_key_id
		FROM
		  acme_accounts
		WHERE
			private_key_id = $1
	`

	row := tx.QueryRowContext(ctx, query, certDb.privateKey.id)

	var exists bool
	err = row.Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		return -2, err
	} else if exists {
		return -2, errors.New("private key in use by acme account")
	}

	err = tx.Commit()
	if err != nil {
		return -2, err
	}

	return id, nil
}
