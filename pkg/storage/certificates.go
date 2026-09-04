package storage

import (
	"certwarden-backend/pkg/domain/certificates"
	"encoding/json"
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
	subjectAltNames             []byte // json: []string
	organization                string
	organizationalUnit          string
	country                     string
	state                       string
	city                        string
	csrExtraExtensions          []byte // json: []certificates.CertExtension
	preferredRootCN             string
	lastAccess                  int64
	createdAt                   int64
	updatedAt                   int64
	apiKey                      string
	apiKeyNew                   string
	apiKeyViaUrl                bool
	postProcessingCommand       string
	postProcessingEnvironment   []byte // json: []string
	postProcessingClientAddress string
	postProcessingClientKeyB64  string // base64 raw url encoded AES 256 key
	profile                     string
}

func (cert *certificateDb) toCertificate() (*certificates.Certificate, error) {
	// slices
	certExts := []certificates.CertExtension{}
	err := json.Unmarshal(cert.csrExtraExtensions, &certExts)
	if err != nil {
		return nil, err
	}

	subjAlts := []string{}
	err = json.Unmarshal(cert.subjectAltNames, &subjAlts)
	if err != nil {
		return nil, err
	}

	postProcEnv := []string{}
	err = json.Unmarshal(cert.postProcessingEnvironment, &postProcEnv)
	if err != nil {
		return nil, err
	}

	return &certificates.Certificate{
		ID:                          cert.id,
		Name:                        cert.name,
		Description:                 cert.description,
		Key:                         *cert.certificateKeyDb.toKey(),
		Account:                     *cert.certificateAccountDb.toAccount(),
		Subject:                     cert.subject,
		SubjectAltNames:             subjAlts,
		Organization:                cert.organization,
		OrganizationalUnit:          cert.organizationalUnit,
		Country:                     cert.country,
		State:                       cert.state,
		City:                        cert.city,
		CSRExtraExtensions:          certExts,
		PreferredRootCN:             cert.preferredRootCN,
		LastAccess:                  time.Unix(cert.lastAccess, 0),
		CreatedAt:                   time.Unix(cert.createdAt, 0),
		UpdatedAt:                   time.Unix(cert.updatedAt, 0),
		ApiKey:                      cert.apiKey,
		ApiKeyNew:                   cert.apiKeyNew,
		ApiKeyViaUrl:                cert.apiKeyViaUrl,
		PostProcessingCommand:       cert.postProcessingCommand,
		PostProcessingEnvironment:   postProcEnv,
		PostProcessingClientAddress: cert.postProcessingClientAddress,
		PostProcessingClientKeyB64:  cert.postProcessingClientKeyB64,
		Profile:                     cert.profile,
	}, nil
}
