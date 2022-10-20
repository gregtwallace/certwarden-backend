package certificates

import (
	"legocerthub-backend/pkg/acme"
	"legocerthub-backend/pkg/challenges"
	"legocerthub-backend/pkg/domain/acme_accounts"
	"legocerthub-backend/pkg/domain/private_keys"
	"legocerthub-backend/pkg/domain/private_keys/key_crypto"
)

// Certificate is a single certificate (summary)
type Certificate struct {
	ID                 int                `json:"id"`
	Name               string             `json:"name"`
	Description        string             `json:"description"`
	CertificateKey     CertificateKey     `json:"private_key"`
	CertificateAccount CertificateAccount `json:"acme_account"`
	Subject            string             `json:"subject"`
	SubjectAltNames    []string           `json:"subject_alts"`
}

type CertificateKey struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type CertificateAccount struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	IsStaging bool   `json:"is_staging"`
}

// temp workaround until further refactor
type acctStore interface {
	GetOneAccountById(id int) (acme_accounts.AccountExtended, error)
}

func (certAcct CertificateAccount) AcmeAccountKey(store acctStore) (acmeKey acme.AccountKey, err error) {
	extendedAcct, err := store.GetOneAccountById(certAcct.ID)
	if err != nil {
		return acme.AccountKey{}, err
	}

	acmeKey, err = extendedAcct.AcmeAccountKey()
	if err != nil {
		return acme.AccountKey{}, err
	}

	return acmeKey, nil
}

// CertificateExtended is a single certificate with all of the
// pertinent details
type CertificateExtended struct {
	Certificate
	CertificateKey     CertificateKeyExtended `json:"private_key"`
	ChallengeMethod    challenges.Method      `json:"challenge_method"`
	Organization       string                 `json:"organization"`
	OrganizationalUnit string                 `json:"organizational_unit"`
	Country            string                 `json:"country"`
	State              string                 `json:"state"`
	City               string                 `json:"city"`
	CreatedAt          int                    `json:"created_at"`
	UpdatedAt          int                    `json:"updated_at"`
	ApiKey             string                 `json:"api_key"`
	ApiKeyViaUrl       bool                   `json:"api_key_via_url"`
}

type CertificateKeyExtended struct {
	CertificateKey
	Algorithm key_crypto.Algorithm `json:"-"`
	Pem       string               `json:"-"`
}

// NewOrderPayload creates the appropriate newOrder payload for ACME
func (cert *Certificate) NewOrderPayload() acme.NewOrderPayload {
	var identifiers []acme.Identifier

	// subject is always required and should be first
	// dns is the only supported type and is hardcoded
	identifiers = append(identifiers, acme.Identifier{Type: "dns", Value: cert.Subject})

	// add alt names if they exist
	if cert.SubjectAltNames != nil {
		for _, name := range cert.SubjectAltNames {
			identifiers = append(identifiers, acme.Identifier{Type: "dns", Value: name})
		}
	}

	return acme.NewOrderPayload{
		Identifiers: identifiers,
	}
}

// new account info
// used to return info about valid options when making a new account
type newCertOptions struct {
	AvailableKeys             []private_keys.Key      `json:"private_keys"`
	UsableAccounts            []acme_accounts.Account `json:"acme_accounts"`
	AvailableChallengeMethods []challenges.Method     `json:"challenge_methods"`
}
