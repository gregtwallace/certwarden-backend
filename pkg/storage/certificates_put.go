package storage

import (
	"certwarden-backend/pkg/domain/certificates"
	"context"
	"time"
)

// PutCertUpdate saves details about the cert that can be updated at any time. It only updates
// the details which are provided
func (store *Storage) PutCertUpdate(payload *certificates.UpdatePayload) (certificates.Certificate, error) {
	// database update
	ctx, cancel := context.WithTimeout(store.shutdownContext, store.timeout)
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
			csr_extra_extensions = case when $10 is null then csr_extra_extensions else $10 end,
			preferred_root_cn = case when $11 is null then preferred_root_cn else $11 end,
			api_key = case when $12 is null then api_key else $12 end,
			api_key_new = case when $13 is null then api_key_new else $13 end,
			api_key_via_url = case when $14 is null then api_key_via_url else $14 end,
			post_processing_command = case when $15 is null then post_processing_command else $15 end,
			post_processing_environment = case when $16 is null then post_processing_environment else $16 end,
			post_processing_client_address = case when $17 is null then post_processing_client_address else $17 end,
			post_processing_client_key = case when $18 is null then post_processing_client_key else $18 end,
			profile = case when $19 is null then profile else $19 end,
			updated_at = $20
		WHERE
			id = $21
		`

	res, err := store.db.ExecContext(ctx, query,
		payload.Name,
		payload.Description,
		payload.PrivateKeyId,
		makeJsonStringSlice(payload.SubjectAltNames, true),
		payload.Organization,
		payload.OrganizationalUnit,
		payload.Country,
		payload.State,
		payload.City,
		makeJsonCertExtensionSlice(payload.CSRExtraExtensions, true),
		payload.PreferredRootCN,
		payload.ApiKey,
		payload.ApiKeyNew,
		payload.ApiKeyViaUrl,
		payload.PostProcessingCommand,
		makeJsonStringSlice(payload.PostProcessingEnvironment, true),
		payload.PostProcessingClientAddress,
		payload.PostProcessingClientKeyB64,
		payload.Profile,
		payload.UpdatedAt.Unix(),
		payload.ID,
	)
	if err != nil {
		return certificates.Certificate{}, err
	}

	// verify update actually happened
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return certificates.Certificate{}, err
	}
	if rowsAffected != 1 {
		return certificates.Certificate{}, errorWrongUpdateRowCount(1, rowsAffected)
	}

	// get updated to return
	updatedCert, err := store.GetOneCertById(payload.ID)
	if err != nil {
		return certificates.Certificate{}, err
	}

	return updatedCert, nil
}

// PutCertUpdatedAt sets the specified cert's updated_at
func (store *Storage) PutCertUpdatedAt(certId int, updatedAt time.Time) (err error) {
	// leverage main Put function
	payload := certificates.UpdatePayload{
		ID:        certId,
		UpdatedAt: updatedAt,
	}

	_, err = store.PutCertUpdate(&payload)
	return err
}

// PutCertApiKey sets a cert's api key and updates the updated at time
func (store *Storage) PutCertApiKey(certId int, apiKey string, updatedAt time.Time) (err error) {
	// leverage main Put function
	payload := certificates.UpdatePayload{
		ID:        certId,
		ApiKey:    &apiKey,
		UpdatedAt: updatedAt,
	}

	_, err = store.PutCertUpdate(&payload)
	return err
}

// PutCertApiKeyNew sets a cert's new api key and updates the updated at time
func (store *Storage) PutCertApiKeyNew(certId int, apiKeyNew string, updatedAt time.Time) (err error) {
	// leverage main Put function
	payload := certificates.UpdatePayload{
		ID:        certId,
		ApiKeyNew: &apiKeyNew,
		UpdatedAt: updatedAt,
	}

	_, err = store.PutCertUpdate(&payload)
	return err
}

// PutCertClientKey sets a cert's client key and updates the updated at time
func (store *Storage) PutCertClientKey(certId int, clientKeyB64 string, updatedAt time.Time) (err error) {
	// leverage main Put function
	payload := certificates.UpdatePayload{
		ID:                         certId,
		PostProcessingClientKeyB64: &clientKeyB64,
		UpdatedAt:                  updatedAt,
	}

	_, err = store.PutCertUpdate(&payload)
	return err
}

// PutCertLastAccess sets a cert's last access time
func (store *Storage) PutCertLastAccess(certId int, lastAccess time.Time) (err error) {
	// database action
	ctx, cancel := context.WithTimeout(store.shutdownContext, store.timeout)
	defer cancel()

	query := `
	UPDATE
		certificates
	SET
		last_access = $1
	WHERE
		id = $2
	`

	res, err := store.db.ExecContext(ctx, query,
		lastAccess.Unix(),
		certId,
	)
	if err != nil {
		return err
	}

	// verify update actually happened
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected != 1 {
		return errorWrongUpdateRowCount(1, rowsAffected)
	}

	return nil
}
