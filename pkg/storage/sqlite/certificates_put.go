package sqlite

import (
	"certwarden-backend/pkg/domain/certificates"
	"context"
	"time"
)

// PutDetailsCert saves details about the cert that can be updated at any time. It only updates
// the details which are provided
func (store *Storage) PutDetailsCert(payload certificates.DetailsUpdatePayload) (certificates.Certificate, error) {
	// database update
	ctx, cancel := context.WithTimeout(context.Background(), store.timeout)
	defer cancel()

	query := `
		UPDATE
			certificates
		SET
			name = case when $1 is null then name else $1 end,
			description = case when $2 is null then description else $2 end,
			private_key_id = case when $3 is null then private_key_id else $3 end,
			subject_alts = case when $4 is null then subject_alts else $4 end,
			csr_org = case when $5 is null then csr_org else $5 end,
			csr_ou = case when $6 is null then csr_ou else $6 end,
			csr_country = case when $7 is null then csr_country else $7 end,
			csr_state = case when $8 is null then csr_state else $8 end,
			csr_city = case when $9 is null then csr_city else $9 end,
			csr_extra_extensions = case when $10 is null then csr_city else $10 end,
			preferred_root_cn = case when $11 is null then preferred_root_cn else $11 end,
			api_key = case when $12 is null then api_key else $12 end,
			api_key_new = case when $13 is null then api_key_new else $13 end,
			api_key_via_url = case when $14 is null then api_key_via_url else $14 end,
			post_processing_command = case when $15 is null then post_processing_command else $15 end,
			post_processing_environment = case when $16 is null then post_processing_environment else $16 end,
			updated_at = $17
		WHERE
			id = $18
		`

	_, err := store.db.ExecContext(ctx, query,
		payload.Name,
		payload.Description,
		payload.PrivateKeyId,
		makeJsonStringSlice(payload.SubjectAltNames),
		payload.Organization,
		payload.OrganizationalUnit,
		payload.Country,
		payload.State,
		payload.City,
		makeJsonCertExtensionSlice(payload.CSRExtraExtensions),
		payload.PreferredRootCN,
		payload.ApiKey,
		payload.ApiKeyNew,
		payload.ApiKeyViaUrl,
		payload.PostProcessingCommand,
		makeJsonStringSlice(payload.PostProcessingEnvironment),
		payload.UpdatedAt,
		payload.ID,
	)

	if err != nil {
		return certificates.Certificate{}, err
	}

	// get updated to return
	updatedCert, err := store.GetOneCertById(payload.ID)
	if err != nil {
		return certificates.Certificate{}, err
	}

	return updatedCert, nil
}

// UpdateCertUpdatedTime sets the specified order's updated_at to now
func (store *Storage) UpdateCertUpdatedTime(certId int) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), store.timeout)
	defer cancel()

	query := `
	UPDATE
		certificates
	SET
		updated_at = $1
	WHERE
		id = $2
	`

	_, err = store.db.ExecContext(ctx, query,
		time.Now().Unix(),
		certId,
	)
	if err != nil {
		return err
	}

	// TODO: Handle 0 rows updated.

	return nil
}

// PutCertNewApiKey sets a cert's new api key and updates the updated at time
func (store *Storage) PutCertNewApiKey(certId int, newApiKey string, updateTimeUnix int) (err error) {
	// database action
	ctx, cancel := context.WithTimeout(context.Background(), store.timeout)
	defer cancel()

	query := `
	UPDATE
		certificates
	SET
		api_key_new = $1,
		updated_at = $2
	WHERE
		id = $3
	`

	_, err = store.db.ExecContext(ctx, query,
		newApiKey,
		updateTimeUnix,
		certId,
	)

	if err != nil {
		return err
	}

	return nil
}

// PutCertApiKey sets a cert's api key and updates the updated at time
func (store *Storage) PutCertApiKey(certId int, apiKey string, updateTimeUnix int) (err error) {
	// database action
	ctx, cancel := context.WithTimeout(context.Background(), store.timeout)
	defer cancel()

	query := `
	UPDATE
		certificates
	SET
		api_key = $1,
		updated_at = $2
	WHERE
		id = $3
	`

	_, err = store.db.ExecContext(ctx, query,
		apiKey,
		updateTimeUnix,
		certId,
	)

	if err != nil {
		return err
	}

	return nil
}

// PutCertClientKey sets a cert's client key and updates the updated at time
func (store *Storage) PutCertClientKey(certId int, newClientKeyB64 string, updateTimeUnix int) (err error) {
	// database action
	ctx, cancel := context.WithTimeout(context.Background(), store.timeout)
	defer cancel()

	query := `
	UPDATE
		certificates
	SET
		post_processing_client_key = $1,
		updated_at = $2
	WHERE
		id = $3
	`

	_, err = store.db.ExecContext(ctx, query,
		newClientKeyB64,
		updateTimeUnix,
		certId,
	)

	if err != nil {
		return err
	}

	return nil
}

// PutCertLastAccess sets a cert's last access time
func (store *Storage) PutCertLastAccess(certId int, unixLastAccessTime int64) (err error) {
	// database action
	ctx, cancel := context.WithTimeout(context.Background(), store.timeout)
	defer cancel()

	query := `
	UPDATE
		certificates
	SET
		last_access = $1
	WHERE
		id = $2
	`

	_, err = store.db.ExecContext(ctx, query,
		unixLastAccessTime,
		certId,
	)

	if err != nil {
		return err
	}

	return nil
}
