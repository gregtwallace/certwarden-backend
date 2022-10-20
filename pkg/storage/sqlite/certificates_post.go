package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"legocerthub-backend/pkg/domain/certificates"
)

// PostNewAccount inserts a new cert into the db
func (store *Storage) PostNewCert(payload certificates.NewPayload) (id int, err error) {
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
		csr_org, csr_ou, csr_country, csr_state, csr_city, created_at, updated_at, api_key, api_key_via_url)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	RETURNING id
	`

	err = tx.QueryRowContext(ctx, query,
		payload.Name,
		payload.Description,
		payload.PrivateKeyID,
		payload.AcmeAccountID,
		payload.ChallengeMethodValue,
		payload.Subject,
		makeCommaJoinedString(payload.SubjectAltNames),
		payload.Organization,
		payload.OrganizationalUnit,
		payload.Country,
		payload.State,
		payload.City,
		payload.CreatedAt,
		payload.UpdatedAt,
		payload.ApiKey,
		payload.ApiKeyViaUrl,
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

	row := tx.QueryRowContext(ctx, query, id)

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
