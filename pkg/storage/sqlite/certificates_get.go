package sqlite

import (
	"context"
	"database/sql"
	"legocerthub-backend/pkg/challenges"
	"legocerthub-backend/pkg/domain/acme_accounts"
	"legocerthub-backend/pkg/domain/private_keys"
	"legocerthub-backend/pkg/storage"
	"time"

	"legocerthub-backend/pkg/domain/certificates"
)

// accountDbToAcc turns the database representation of a certificate into a Certificate
func (certDb *certificateDb) certDbToCert() (cert certificates.Certificate, err error) {
	// convert embedded private key db
	var privateKey = new(private_keys.Key)
	if certDb.privateKey != nil {
		*privateKey, err = certDb.privateKey.keyDbToKey()
		if err != nil {
			return certificates.Certificate{}, err
		}
	} else {
		privateKey = nil
	}

	// convert embedded account db
	var acmeAccount = new(acme_accounts.Account)
	if certDb.acmeAccount != nil {
		*acmeAccount, err = certDb.acmeAccount.accountDbToAcc()
		if err != nil {
			return certificates.Certificate{}, err
		}
	} else {
		acmeAccount = nil
	}

	// if there is a challenge type value, specify the challenge method
	var challengeMethod = new(challenges.Method)
	if certDb.challengeMethodValue.Valid {
		*challengeMethod = challenges.MethodByValue(certDb.challengeMethodValue.String)
	} else {
		challengeMethod = nil
	}

	return certificates.Certificate{
		ID:                 nullInt32ToInt(certDb.id),
		Name:               nullStringToString(certDb.name),
		Description:        nullStringToString(certDb.description),
		PrivateKey:         privateKey,
		AcmeAccount:        acmeAccount,
		ChallengeMethod:    challengeMethod,
		Subject:            nullStringToString(certDb.subject),
		SubjectAltNames:    commaNullStringToSlice(certDb.subjectAltNames),
		Organization:       nullStringToString(certDb.organization),
		OrganizationalUnit: nullStringToString(certDb.organizationalUnit),
		Country:            nullStringToString(certDb.country),
		State:              nullStringToString(certDb.state),
		City:               nullStringToString(certDb.city),
		CreatedAt:          nullInt32ToInt(certDb.createdAt),
		UpdatedAt:          nullInt32ToInt(certDb.updatedAt),
		ApiKey:             nullStringToString(certDb.apiKey),
	}, nil
}

func (store *Storage) GetAllCerts() (certs []certificates.Certificate, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), store.Timeout)
	defer cancel()

	query := `
	SELECT c.id, c.name, c.subject, c.subject_alts, c.description, pk.id, pk.name,
	aa.id, aa.name, aa.is_staging
	FROM
		certificates c
		LEFT JOIN private_keys pk on (c.private_key_id = pk.id)
		LEFT JOIN acme_accounts aa on (c.acme_account_id = aa.id)
	ORDER BY c.name
	`

	rows, err := store.Db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var oneCert certificateDb
		// initialize keyDb & accountDb pointer (or nil deref)
		oneCert.privateKey = new(keyDb)
		oneCert.acmeAccount = new(accountDb)
		err = rows.Scan(
			&oneCert.id,
			&oneCert.name,
			&oneCert.subject,
			&oneCert.subjectAltNames,
			&oneCert.description,
			&oneCert.privateKey.id,
			&oneCert.privateKey.name,
			&oneCert.acmeAccount.id,
			&oneCert.acmeAccount.name,
			&oneCert.acmeAccount.isStaging,
		)
		if err != nil {
			return nil, err
		}

		convertedCert, err := oneCert.certDbToCert()
		if err != nil {
			return nil, err
		}

		certs = append(certs, convertedCert)
	}

	return certs, nil
}

// GetOneCertById returns a Cert based on its unique id
func (store *Storage) GetOneCertById(id int, withKeyPems bool) (cert certificates.Certificate, err error) {
	return store.getOneCert(id, "", withKeyPems)
}

// GetOneCertByName returns a Cert based on its unique name
func (store *Storage) GetOneCertByName(name string, withKeyPems bool) (cert certificates.Certificate, err error) {
	return store.getOneCert(-1, name, withKeyPems)
}

// getOneCert returns a Cert based on either its unique id or its unique name
func (store *Storage) getOneCert(id int, name string, withKeyPems bool) (cert certificates.Certificate, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), store.Timeout)
	defer cancel()

	query := `
	SELECT c.id, c.name, c.description, c.challenge_method, c.subject, c.subject_alts, 
	c.csr_org, c.csr_ou, c.csr_country, c.csr_city, c.created_at, c.updated_at, c.api_key, 
	aa.id, aa.name, aa.is_staging, aa.kid,
	ak.id, ak.name, ak.algorithm, ak.pem,
	pk.id, pk.name, pk.algorithm, pk.pem
	FROM
		certificates c
		LEFT JOIN acme_accounts aa on (c.acme_account_id = aa.id)
		LEFT JOIN private_keys ak on (aa.private_key_id = ak.id)
		LEFT JOIN private_keys pk on (c.private_key_id = pk.id)
	WHERE c.id = $1 OR c.name = $2
	ORDER BY c.name
	`

	row := store.Db.QueryRowContext(ctx, query, id, name)

	var oneCert certificateDb
	// initialize accountDb, accountDb's keyDb, and keyDb pointer (or nil deref)
	oneCert.acmeAccount = new(accountDb)
	oneCert.acmeAccount.privateKey = new(keyDb)
	oneCert.privateKey = new(keyDb)

	err = row.Scan(
		&oneCert.id,
		&oneCert.name,
		&oneCert.description,
		&oneCert.challengeMethodValue,
		&oneCert.subject,
		&oneCert.subjectAltNames,
		&oneCert.organization,
		&oneCert.organizationalUnit,
		&oneCert.country,
		&oneCert.city,
		&oneCert.createdAt,
		&oneCert.updatedAt,
		&oneCert.apiKey,
		&oneCert.acmeAccount.id,
		&oneCert.acmeAccount.name,
		&oneCert.acmeAccount.isStaging,
		&oneCert.acmeAccount.kid,
		&oneCert.acmeAccount.privateKey.id,
		&oneCert.acmeAccount.privateKey.name,
		&oneCert.acmeAccount.privateKey.algorithmValue,
		&oneCert.acmeAccount.privateKey.pem,
		&oneCert.privateKey.id,
		&oneCert.privateKey.name,
		&oneCert.privateKey.algorithmValue,
		&oneCert.privateKey.pem,
	)

	if err != nil {
		// if no record exists
		if err == sql.ErrNoRows {
			err = storage.ErrNoRecord
		}
		return certificates.Certificate{}, err
	}

	// if not fetching pems, invalidate them
	if !withKeyPems {
		oneCert.acmeAccount.privateKey.pem.Valid = false
		oneCert.privateKey.pem.Valid = false
	}

	cert, err = oneCert.certDbToCert()
	if err != nil {
		return certificates.Certificate{}, err
	}

	return cert, nil
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
