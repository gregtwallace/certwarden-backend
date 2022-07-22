package sqlite

import (
	"context"
	"database/sql"
	"legocerthub-backend/pkg/domain/acme_accounts"
	"legocerthub-backend/pkg/domain/certificates/challenges"
	"legocerthub-backend/pkg/domain/private_keys"
	"legocerthub-backend/pkg/storage"

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
	var challengeMethod = new(challenges.ChallengeMethod)
	if certDb.challengeMethodValue.Valid {
		*challengeMethod, err = challenges.ChallengeMethodByValue(certDb.challengeMethodValue.String)
		if err != nil {
			return certificates.Certificate{}, err
		}
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
		CommonName:         nullStringToString(certDb.commonName),
		Organization:       nullStringToString(certDb.organization),
		OrganizationalUnit: nullStringToString(certDb.organizationalUnit),
		Country:            nullStringToString(certDb.country),
		State:              nullStringToString(certDb.state),
		City:               nullStringToString(certDb.city),
		CreatedAt:          nullInt32ToInt(certDb.createdAt),
		UpdatedAt:          nullInt32ToInt(certDb.updatedAt),
		ApiKey:             nullStringToString(certDb.apiKey),
		Pem:                nullStringToString(certDb.pem),
		ValidFrom:          nullInt32ToInt(certDb.validFrom),
		ValidTo:            nullInt32ToInt(certDb.validTo),
	}, nil
}

func (store *Storage) GetAllCerts() (certs []certificates.Certificate, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), store.Timeout)
	defer cancel()

	query := `
	SELECT c.id, c.name, c.subject, c.subject_alts, c.description, pk.id, pk.name,
	aa.id, aa.name, aa.is_staging, c.valid_to
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
			&oneCert.validTo,
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
func (store *Storage) GetOneCertById(id int, withAcctPem bool) (cert certificates.Certificate, err error) {
	return store.getOneCert(id, "", withAcctPem)
}

// GetOneCertByName returns a Cert based on its unique name
func (store *Storage) GetOneCertByName(name string, withAcctPem bool) (cert certificates.Certificate, err error) {
	return store.getOneCert(-1, name, withAcctPem)
}

// getOneCert returns a Cert based on either its unique id or its unique name
func (store *Storage) getOneCert(id int, name string, withAcctPem bool) (cert certificates.Certificate, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), store.Timeout)
	defer cancel()

	query := `
	SELECT c.id, c.name, c.description, c.challenge_method, c.subject, c.subject_alts, c.csr_com_name, 
	c.csr_org, c.csr_country, c.csr_city, c.created_at, c.updated_at, c.api_key, c.pem, c.valid_from, c.valid_to,
	aa.id, aa.name, aa.is_staging, aa.kid,
	ak.id, ak.name, ak.algorithm, ak.pem,
	pk.id, pk.name, pk.algorithm
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
		&oneCert.commonName,
		&oneCert.organization,
		&oneCert.country,
		&oneCert.city,
		&oneCert.createdAt,
		&oneCert.updatedAt,
		&oneCert.apiKey,
		&oneCert.pem,
		&oneCert.validFrom,
		&oneCert.validTo,
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
	)

	if err != nil {
		// if no record exists
		if err == sql.ErrNoRows {
			err = storage.ErrNoRecord
		}
		return certificates.Certificate{}, err
	}

	// if not fetching account pem, invalidate it
	if !withAcctPem {
		oneCert.acmeAccount.privateKey.pem.Valid = false
	}

	cert, err = oneCert.certDbToCert()
	if err != nil {
		return certificates.Certificate{}, err
	}

	return cert, nil
}
