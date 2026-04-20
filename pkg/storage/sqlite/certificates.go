package sqlite

import (
	"certwarden-backend/pkg/domain/certificates"
	"time"
)

// certificateDb is a single certificate, as database table fields
// corresponds to certificates.Certificate
type certificateDb struct {
	id                          int
	name                        string
	description                 string
	certificateKeyDb            keyDb
	certificateAccountDb        accountDb
	subject                     string
	subjectAltNames             jsonStringSlice // stored as json array
	organization                string
	organizationalUnit          string
	country                     string
	state                       string
	city                        string
	csrExtraExtensions          jsonCertExtensionSlice
	preferredRootCN             string
	lastAccess                  int64
	createdAt                   int64
	updatedAt                   int64
	apiKey                      string
	apiKeyNew                   string
	apiKeyViaUrl                bool
	postProcessingCommand       string
	postProcessingEnvironment   jsonStringSlice // stored as json array
	postProcessingClientAddress string
	postProcessingClientKeyB64  string // base64 raw url encoded AES 256 key
	profile                     string
	techPhone                   string
	techEmail                   string
}

func (cert certificateDb) toCertificate() (certificates.Certificate, error) {
	certExt, err := cert.csrExtraExtensions.toCertExtensionSlice()
	if err != nil {
		return certificates.Certificate{}, err
	}

	return certificates.Certificate{
		ID:                          cert.id,
		Name:                        cert.name,
		Description:                 cert.description,
		CertificateKey:              cert.certificateKeyDb.toKey(),
		CertificateAccount:          cert.certificateAccountDb.toAccount(),
		Subject:                     cert.subject,
		SubjectAltNames:             cert.subjectAltNames.toSlice(),
		Organization:                cert.organization,
		OrganizationalUnit:          cert.organizationalUnit,
		Country:                     cert.country,
		State:                       cert.state,
		City:                        cert.city,
		CSRExtraExtensions:          certExt,
		PreferredRootCN:             cert.preferredRootCN,
		LastAccess:                  time.Unix(cert.lastAccess, 0),
		CreatedAt:                   time.Unix(cert.createdAt, 0),
		UpdatedAt:                   time.Unix(cert.updatedAt, 0),
		ApiKey:                      cert.apiKey,
		ApiKeyNew:                   cert.apiKeyNew,
		ApiKeyViaUrl:                cert.apiKeyViaUrl,
		PostProcessingCommand:       cert.postProcessingCommand,
		PostProcessingEnvironment:   cert.postProcessingEnvironment.toSlice(),
		PostProcessingClientAddress: cert.postProcessingClientAddress,
		PostProcessingClientKeyB64:  cert.postProcessingClientKeyB64,
		Profile:                     cert.profile,
		TechPhone:                   cert.techPhone,
		TechEmail:                   cert.techEmail,
	}, nil
}
