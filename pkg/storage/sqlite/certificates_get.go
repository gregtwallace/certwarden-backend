package sqlite

import (
	"context"
	"legocerthub-backend/pkg/domain/acme_accounts"
	"legocerthub-backend/pkg/domain/certificates/challenges"
	"legocerthub-backend/pkg/domain/private_keys"
	"strings"

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

	// if there is a challenge type value, specify the challenge type
	var challengeType = new(challenges.ChallengeType)
	if certDb.challengeTypeValue.Valid {
		*challengeType, err = challenges.ChallengeTypeByValue(certDb.challengeTypeValue.String)
		if err != nil {
			return certificates.Certificate{}, err
		}
	} else {
		challengeType = nil
	}

	// subject alts convert from strings separated by comma to slice
	var subjectAlts = new([]string)
	if certDb.subjectAlts.Valid {
		// if the string isn't empty, parse it
		if certDb.subjectAlts.String != "" {
			s := nullStringToString(certDb.subjectAlts)
			*subjectAlts = strings.Split(*s, ",")
		} else {
			// empty string = empty array
			*subjectAlts = []string{}
		}
	} else {
		subjectAlts = nil
	}

	return certificates.Certificate{
		ID:            nullInt32ToInt(certDb.id),
		Name:          nullStringToString(certDb.name),
		Description:   nullStringToString(certDb.description),
		PrivateKey:    privateKey,
		AcmeAccount:   acmeAccount,
		ChallengeType: challengeType,
		Subject:       nullStringToString(certDb.subject),
		SubjectAlts:   subjectAlts,
		CommonName:    nullStringToString(certDb.commonName),
		Organization:  nullStringToString(certDb.organization),
		Country:       nullStringToString(certDb.country),
		State:         nullStringToString(certDb.state),
		City:          nullStringToString(certDb.city),
		CreatedAt:     nullInt32ToInt(certDb.createdAt),
		UpdatedAt:     nullInt32ToInt(certDb.updatedAt),
		ApiKey:        nullStringToString(certDb.apiKey),
		Pem:           nullStringToString(certDb.pem),
		ValidFrom:     nullInt32ToInt(certDb.validFrom),
		ValidTo:       nullInt32ToInt(certDb.validTo),
	}, nil
}

func (store *Storage) GetAllCertificates() (certs []certificates.Certificate, err error) {
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

	var allCerts []certificates.Certificate
	for rows.Next() {
		var oneCert certificateDb
		// initialize keyDb & accountDb pointer (or nil deref)
		oneCert.privateKey = new(keyDb)
		oneCert.acmeAccount = new(accountDb)
		err = rows.Scan(
			&oneCert.id,
			&oneCert.name,
			&oneCert.subject,
			&oneCert.subjectAlts,
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

		allCerts = append(allCerts, convertedCert)
	}

	return allCerts, nil
}
