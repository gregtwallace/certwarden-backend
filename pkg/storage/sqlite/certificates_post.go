package sqlite

import (
	"context"
	"legocerthub-backend/pkg/domain/certificates"
)

// PostNewAccount inserts a new cert into the db
func (store *Storage) PostNewCert(payload certificates.NewPayload) (certificates.Certificate, error) {
	// database update
	ctx, cancel := context.WithTimeout(context.Background(), store.timeout)
	defer cancel()

	// don't check for in use in storage. main app business logic should
	// take care of it

	// insert the new cert
	query := `
	INSERT INTO certificates (name, description, private_key_id, acme_account_id, subject, subject_alts, 
		csr_org, csr_ou, csr_country, csr_state, csr_city, created_at, updated_at, api_key, api_key_via_url)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	RETURNING id
	`

	id := -1
	err := store.db.QueryRowContext(ctx, query,
		payload.Name,
		payload.Description,
		payload.PrivateKeyID,
		payload.AcmeAccountID,
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
		return certificates.Certificate{}, err
	}

	// get updated to return
	newCert, err := store.GetOneCertById(id)
	if err != nil {
		return certificates.Certificate{}, err
	}

	return newCert, nil
}
