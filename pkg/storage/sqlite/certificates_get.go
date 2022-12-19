package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"legocerthub-backend/pkg/pagination_sort"
	"legocerthub-backend/pkg/storage"
	"time"

	"legocerthub-backend/pkg/domain/certificates"
)

func (store *Storage) GetAllCerts(q pagination_sort.Query) (certs []certificates.Certificate, totalRowCount int, err error) {
	// validate and set sort
	sortField := q.SortField()

	switch sortField {
	// allow these
	case "id":
		sortField = "c.id"
	case "name":
		sortField = "c.name"
	case "subject":
		sortField = "c.subject"
	case "keyname":
		sortField = "pk.name"
	case "accountname":
		sortField = "aa.name"
	// default if not in allowed list
	default:
		sortField = "c.name"
	}

	sort := sortField + " " + q.SortDirection()

	// do query
	ctx, cancel := context.WithTimeout(context.Background(), store.Timeout)
	defer cancel()

	// WARNING: SQL Injection is possible if the variables are not properly
	// validated prior to this query being assembled!
	query := fmt.Sprintf(`
	SELECT 
		c.id, c.name, c.description, c.subject, c.subject_alts, c.challenge_method, 
		c.csr_org, c.csr_ou, c.csr_country, c.csr_state, c.csr_city, c.created_at, c.updated_at,
		c.api_key, c.api_key_via_url,
		
		pk.id, pk.name, pk.description, pk.algorithm, pk.pem, pk.api_key, pk.api_key_via_url,
		pk.created_at, pk.updated_at,

		aa.id, aa.name, aa.description, aa.status, aa.email, aa.accepted_tos, aa.is_staging,
		aa.created_at, aa.updated_at, aa.kid,

		ak.id, ak.name, ak.description, ak.algorithm, ak.pem, ak.api_key, ak.api_key_via_url,
		ak.created_at, ak.updated_at,

		count(*) OVER() AS full_count
	FROM
		certificates c
		LEFT JOIN private_keys pk on (c.private_key_id = pk.id)
		LEFT JOIN acme_accounts aa on (c.acme_account_id = aa.id)
		LEFT JOIN private_keys ak on (aa.private_key_id = ak.id)
	ORDER BY
		%s
	LIMIT
		$1
	OFFSET
		$2
	`, sort)

	rows, err := store.Db.QueryContext(ctx, query,
		q.Limit(),
		q.Offset(),
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	// for total row count
	var totalRows int

	for rows.Next() {
		var oneCert certificateDb

		err = rows.Scan(
			&oneCert.id,
			&oneCert.name,
			&oneCert.description,
			&oneCert.subject,
			&oneCert.subjectAltNames,
			&oneCert.challengeMethodValue,
			&oneCert.organization,
			&oneCert.organizationalUnit,
			&oneCert.country,
			&oneCert.state,
			&oneCert.city,
			&oneCert.createdAt,
			&oneCert.updatedAt,
			&oneCert.apiKey,
			&oneCert.apiKeyViaUrl,

			&oneCert.certificateKeyDb.id,
			&oneCert.certificateKeyDb.name,
			&oneCert.certificateKeyDb.description,
			&oneCert.certificateKeyDb.algorithmValue,
			&oneCert.certificateKeyDb.pem,
			&oneCert.certificateKeyDb.apiKey,
			&oneCert.certificateKeyDb.apiKeyViaUrl,
			&oneCert.certificateKeyDb.createdAt,
			&oneCert.certificateKeyDb.updatedAt,

			&oneCert.certificateAccountDb.id,
			&oneCert.certificateAccountDb.name,
			&oneCert.certificateAccountDb.description,
			&oneCert.certificateAccountDb.status,
			&oneCert.certificateAccountDb.email,
			&oneCert.certificateAccountDb.acceptedTos,
			&oneCert.certificateAccountDb.isStaging,
			&oneCert.certificateAccountDb.createdAt,
			&oneCert.certificateAccountDb.updatedAt,
			&oneCert.certificateAccountDb.kid,

			&oneCert.certificateAccountDb.accountKeyDb.id,
			&oneCert.certificateAccountDb.accountKeyDb.name,
			&oneCert.certificateAccountDb.accountKeyDb.description,
			&oneCert.certificateAccountDb.accountKeyDb.algorithmValue,
			&oneCert.certificateAccountDb.accountKeyDb.pem,
			&oneCert.certificateAccountDb.accountKeyDb.apiKey,
			&oneCert.certificateAccountDb.accountKeyDb.apiKeyViaUrl,
			&oneCert.certificateAccountDb.accountKeyDb.createdAt,
			&oneCert.certificateAccountDb.accountKeyDb.updatedAt,

			&totalRows,
		)
		if err != nil {
			return nil, 0, err
		}

		// convert and append
		certs = append(certs, oneCert.toCertificate(store))
	}

	return certs, totalRows, nil
}

// GetOneCertById returns a Cert based on its unique id
func (store *Storage) GetOneCertById(id int) (cert certificates.Certificate, err error) {
	return store.getOneCert(id, "")
}

// GetOneCertByName returns a Cert based on its unique name
func (store *Storage) GetOneCertByName(name string) (cert certificates.Certificate, err error) {
	return store.getOneCert(-1, name)
}

// getOneCert returns a Cert based on either its unique id or its unique name
func (store *Storage) getOneCert(id int, name string) (cert certificates.Certificate, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), store.Timeout)
	defer cancel()

	query := `
	SELECT
		c.id, c.name, c.description, c.subject, c.subject_alts, c.challenge_method, 
		c.csr_org, c.csr_ou, c.csr_country, c.csr_state, c.csr_city, c.created_at, c.updated_at,
		c.api_key, c.api_key_via_url,
		
		pk.id, pk.name, pk.description, pk.algorithm, pk.pem, pk.api_key, pk.api_key_via_url,
		pk.created_at, pk.updated_at,

		aa.id, aa.name, aa.description, aa.status, aa.email, aa.accepted_tos, aa.is_staging,
		aa.created_at, aa.updated_at, aa.kid,

		ak.id, ak.name, ak.description, ak.algorithm, ak.pem, ak.api_key, ak.api_key_via_url,
		ak.created_at, ak.updated_at
	FROM
		certificates c
		LEFT JOIN private_keys pk on (c.private_key_id = pk.id)
		LEFT JOIN acme_accounts aa on (c.acme_account_id = aa.id)
		LEFT JOIN private_keys ak on (aa.private_key_id = ak.id)
	WHERE 
		c.id = $1 OR c.name = $2
	ORDER BY c.name
	`

	row := store.Db.QueryRowContext(ctx, query, id, name)

	var oneCert certificateDb

	err = row.Scan(
		&oneCert.id,
		&oneCert.name,
		&oneCert.description,
		&oneCert.subject,
		&oneCert.subjectAltNames,
		&oneCert.challengeMethodValue,
		&oneCert.organization,
		&oneCert.organizationalUnit,
		&oneCert.country,
		&oneCert.state,
		&oneCert.city,
		&oneCert.createdAt,
		&oneCert.updatedAt,
		&oneCert.apiKey,
		&oneCert.apiKeyViaUrl,

		&oneCert.certificateKeyDb.id,
		&oneCert.certificateKeyDb.name,
		&oneCert.certificateKeyDb.description,
		&oneCert.certificateKeyDb.algorithmValue,
		&oneCert.certificateKeyDb.pem,
		&oneCert.certificateKeyDb.apiKey,
		&oneCert.certificateKeyDb.apiKeyViaUrl,
		&oneCert.certificateKeyDb.createdAt,
		&oneCert.certificateKeyDb.updatedAt,

		&oneCert.certificateAccountDb.id,
		&oneCert.certificateAccountDb.name,
		&oneCert.certificateAccountDb.description,
		&oneCert.certificateAccountDb.status,
		&oneCert.certificateAccountDb.email,
		&oneCert.certificateAccountDb.acceptedTos,
		&oneCert.certificateAccountDb.isStaging,
		&oneCert.certificateAccountDb.createdAt,
		&oneCert.certificateAccountDb.updatedAt,
		&oneCert.certificateAccountDb.kid,

		&oneCert.certificateAccountDb.accountKeyDb.id,
		&oneCert.certificateAccountDb.accountKeyDb.name,
		&oneCert.certificateAccountDb.accountKeyDb.description,
		&oneCert.certificateAccountDb.accountKeyDb.algorithmValue,
		&oneCert.certificateAccountDb.accountKeyDb.pem,
		&oneCert.certificateAccountDb.accountKeyDb.apiKey,
		&oneCert.certificateAccountDb.accountKeyDb.apiKeyViaUrl,
		&oneCert.certificateAccountDb.accountKeyDb.createdAt,
		&oneCert.certificateAccountDb.accountKeyDb.updatedAt,
	)

	if err != nil {
		// if no record exists
		if err == sql.ErrNoRows {
			err = storage.ErrNoRecord
		}
		return certificates.Certificate{}, err
	}

	// convert and return
	return oneCert.toCertificate(store), nil
}

// GetCertPemById returns a the pem from the most recent valid order for the specified
// cert id
func (store *Storage) GetCertPemById(id int) (pem string, err error) {
	return store.getCertPem(id, "")
}

// GetCertPemByName returns a the pem from the most recent valid order for the specified
// cert name
func (store *Storage) GetCertPemByName(name string) (pem string, err error) {
	return store.getCertPem(-1, name)
}

// GetCertPem returns the pem for the most recent valid order of the specified
// cert (id or name)
func (store *Storage) getCertPem(certId int, certName string) (pem string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), store.Timeout)
	defer cancel()

	query := `
	SELECT
		pem
	FROM
		acme_orders ao
		LEFT JOIN certificates c on (ao.certificate_id = c.id)
	WHERE 
		ao.status = "valid"
		AND
		ao.known_revoked = 0
		AND
		ao.valid_to > $1
		AND
		ao.pem NOT NULL
		AND
		(
			ao.certificate_id = $2
			OR
			c.name = $3
		)
	GROUP BY
		certificate_id
	HAVING
		MAX(valid_to)
	`

	row := store.Db.QueryRowContext(ctx, query,
		time.Now().Unix(),
		certId,
		certName,
	)

	err = row.Scan(&pem)
	if err != nil {
		return "", err
	}

	return pem, nil
}
