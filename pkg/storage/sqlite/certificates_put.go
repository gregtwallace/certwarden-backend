package sqlite

import (
	"context"
	"errors"
	"legocerthub-backend/pkg/domain/certificates"
)

// accountPayloadToDb turns the client payload into a db object
func certDetailsPayloadToDb(payload certificates.DetailsUpdatePayload) (certificateDb, error) {
	var certDb certificateDb

	// mandatory, error if somehow does not exist
	if payload.ID == nil {
		return certificateDb{}, errors.New("missing id")
	}
	certDb.id = intToNullInt32(payload.ID)

	certDb.name = stringToNullString(payload.Name)
	certDb.description = stringToNullString(payload.Description)
	certDb.challengeMethodValue = stringToNullString(payload.ChallengeMethodValue)

	certDb.commonName = stringToNullString(payload.CommonName)
	certDb.organization = stringToNullString(payload.Organization)
	certDb.organizationalUnit = stringToNullString(payload.OrganizationalUnit)
	certDb.country = stringToNullString(payload.Country)
	certDb.state = stringToNullString(payload.State)
	certDb.city = stringToNullString(payload.City)

	certDb.updatedAt = timeNow()

	return certDb, nil
}

// PutDetailsCert saves details about the cert that can be updated at any time. It only updates
// the details which are provided
func (store *Storage) PutDetailsCert(payload certificates.DetailsUpdatePayload) (err error) {
	// Load payload into db obj
	certDb, err := certDetailsPayloadToDb(payload)
	if err != nil {
		return err
	}

	// database update
	ctx, cancel := context.WithTimeout(context.Background(), store.Timeout)
	defer cancel()

	query := `
		UPDATE
			certificates
		SET
			name = case when $1 is null then name else $1 end,
			description = case when $2 is null then description else $2 end,
			challenge_method = case when $3 is null then challenge_method else $3 end,
			csr_com_name = case when $4 is null then csr_com_name else $4 end,
			csr_org = case when $5 is null then csr_org else $5 end,
			csr_ou = case when $6 is null then csr_ou else $6 end,
			csr_country = case when $7 is null then csr_country else $7 end,
			csr_state = case when $8 is null then csr_state else $8 end,
			csr_city = case when $9 is null then csr_city else $9 end,
			updated_at = $10
		WHERE
			id = $11
		`

	_, err = store.Db.ExecContext(ctx, query,
		certDb.name,
		certDb.description,
		certDb.challengeMethodValue,
		certDb.commonName,
		certDb.organization,
		certDb.organizationalUnit,
		certDb.country,
		certDb.state,
		certDb.city,
		certDb.updatedAt,
		certDb.id,
	)

	if err != nil {
		return err
	}

	return nil
}