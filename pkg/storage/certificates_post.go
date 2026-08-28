package storage

import (
	"certwarden-backend/pkg/domain/certificates"
	"context"
)

// PostNewAccount inserts a new cert into the db
func (store *Storage) PostNewCert(payload *certificates.NewPayload) (*certificates.Certificate, error) {
	// database update
	ctx, cancel := context.WithTimeout(store.shutdownContext, store.timeout)
	defer cancel()

	// don't check for in use in storage. main app business logic should
	// take care of it

	// slices
	certExts, err := sliceToJsonString(payload.CSRExtraExtensions, false)
	if err != nil {
		return nil, err
	}

	subjAlts, err := sliceToJsonString(payload.SubjectAltNames, false)
	if err != nil {
		return nil, err
	}

	postProcEnv, err := sliceToJsonString(payload.PostProcessingEnvironment, false)
	if err != nil {
		return nil, err
	}

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
	err = store.db.QueryRowContext(ctx, query,
		payload.Name,
		payload.Description,
		payload.PrivateKeyID,
		payload.AcmeAccountID,
		payload.Subject,
		subjAlts,
		payload.Organization,
		payload.OrganizationalUnit,
		payload.Country,
		payload.State,
		payload.City,
		certExts,
		payload.PreferredRootCN,
		payload.CreatedAt.Unix(),
		payload.UpdatedAt.Unix(),
		payload.ApiKey,
		payload.ApiKeyViaUrl,
		payload.PostProcessingCommand,
		postProcEnv,
		payload.PostProcessingClientAddress,
		payload.PostProcessingClientKeyB64,
		payload.Profile,
	).Scan(&id)

	if err != nil {
		return nil, err
	}

	// get updated to return
	newCert, err := store.GetOneCertById(id)
	if err != nil {
		return nil, err
	}

	return newCert, nil
}
