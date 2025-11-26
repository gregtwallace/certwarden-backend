package sqlite

import (
	"certwarden-backend/pkg/domain/certificates"
	"context"
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
		csr_org, csr_ou, csr_country, csr_state, csr_city, csr_extra_extensions, preferred_root_cn, 
		created_at, updated_at, api_key, api_key_via_url,
		post_processing_command, post_processing_environment, post_processing_client_address, 
		post_processing_client_key, profile)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22)
	RETURNING id
	`

	id := -1
	err := store.db.QueryRowContext(ctx, query,
		payload.Name,
		payload.Description,
		payload.PrivateKeyID,
		payload.AcmeAccountID,
		payload.Subject,
		makeJsonStringSlice(payload.SubjectAltNames),
		payload.Organization,
		payload.OrganizationalUnit,
		payload.Country,
		payload.State,
		payload.City,
		makeJsonCertExtensionSlice(payload.CSRExtraExtensions),
		payload.PreferredRootCN,
		payload.CreatedAt,
		payload.UpdatedAt,
		payload.ApiKey,
		payload.ApiKeyViaUrl,
		payload.PostProcessingCommand,
		makeJsonStringSlice(payload.PostProcessingEnvironment),
		payload.PostProcessingClientAddress,
		payload.PostProcessingClientKeyB64,
		payload.Profile,
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
