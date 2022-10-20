package sqlite

import (
	"legocerthub-backend/pkg/challenges"
	"legocerthub-backend/pkg/domain/certificates"
	"legocerthub-backend/pkg/domain/private_keys/key_crypto"
)

// certificateDb is a single certificate, as database table fields
// corresponds to certificates.Certificate
type certificateDb struct {
	id                   int
	name                 string
	description          string
	certificateKeyDb     certificateKeyDb
	certificateAccountDb certificateAccountDb
	subject              string
	subjectAltNames      commaJoinedStrings
}

type certificateKeyDb struct {
	id   int
	name string
}

type certificateAccountDb struct {
	id        int
	name      string
	isStaging bool
}

func (cert certificateDb) toCertificate() certificates.Certificate {
	return certificates.Certificate{
		ID:                 cert.id,
		Name:               cert.name,
		Description:        cert.description,
		CertificateKey:     cert.certificateKeyDb.toCertificateKey(),
		CertificateAccount: cert.certificateAccountDb.toCertificateAccount(),
		Subject:            cert.subject,
		SubjectAltNames:    cert.subjectAltNames.toSlice(),
	}
}

func (certKey certificateKeyDb) toCertificateKey() certificates.CertificateKey {
	return certificates.CertificateKey{
		ID:   certKey.id,
		Name: certKey.name,
	}
}

func (certAcct certificateAccountDb) toCertificateAccount() certificates.CertificateAccount {
	return certificates.CertificateAccount{
		ID:        certAcct.id,
		Name:      certAcct.name,
		IsStaging: certAcct.isStaging,
	}
}

// certificateDbExtended is a single certificate, as database table
// fields. corresponds to certificates.CertificateExtended
type certificateExtendedDb struct {
	certificateDb
	certificateKeyDb     certificateKeyExtendedDb
	challengeMethodValue string
	organization         string
	organizationalUnit   string
	country              string
	state                string
	city                 string
	createdAt            int
	updatedAt            int
	apiKey               string
	apiKeyViaUrl         bool
}

type certificateKeyExtendedDb struct {
	certificateKeyDb
	algorithmValue string
	pem            string
}

func (cert certificateExtendedDb) toCertificateExtended() certificates.CertificateExtended {
	return certificates.CertificateExtended{
		// regular fields
		Certificate: cert.toCertificate(),
		// extended fields
		CertificateKey: cert.certificateKeyDb.toCertificateKeyExtended(),

		ChallengeMethod:    challenges.MethodByValue(cert.challengeMethodValue),
		Organization:       cert.organization,
		OrganizationalUnit: cert.organizationalUnit,
		Country:            cert.country,
		State:              cert.state,
		City:               cert.city,
		CreatedAt:          cert.createdAt,
		UpdatedAt:          cert.updatedAt,
		ApiKey:             cert.apiKey,
		ApiKeyViaUrl:       cert.apiKeyViaUrl,
	}
}

func (certKey certificateKeyExtendedDb) toCertificateKeyExtended() certificates.CertificateKeyExtended {
	return certificates.CertificateKeyExtended{
		// regular
		CertificateKey: certKey.toCertificateKey(),
		// extended
		Algorithm: key_crypto.AlgorithmByValue(certKey.algorithmValue),
		Pem:       certKey.pem,
	}
}
