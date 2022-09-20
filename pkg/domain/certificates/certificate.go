package certificates

import (
	"legocerthub-backend/pkg/acme"
	"legocerthub-backend/pkg/challenges"
	"legocerthub-backend/pkg/domain/acme_accounts"
	"legocerthub-backend/pkg/domain/private_keys"
)

// a single certificate
type Certificate struct {
	ID                 *int                      `json:"id"`
	Name               *string                   `json:"name"`
	Description        *string                   `json:"description"`
	PrivateKey         *private_keys.KeyExtended `json:"private_key,omitempty"`
	AcmeAccount        *acme_accounts.Account    `json:"acme_account,omitempty"`
	ChallengeMethod    *challenges.Method        `json:"challenge_method,omitempty"`
	Subject            *string                   `json:"subject,omitempty"`
	SubjectAltNames    *[]string                 `json:"subject_alts,omitempty"`
	Organization       *string                   `json:"organization,omitempty"`
	OrganizationalUnit *string                   `json:"organizational_unit,omitempty"`
	Country            *string                   `json:"country,omitempty"`
	State              *string                   `json:"state,omitempty"`
	City               *string                   `json:"city,omitempty"`
	CreatedAt          *int                      `json:"created_at,omitempty"`
	UpdatedAt          *int                      `json:"updated_at,omitempty"`
	ApiKey             *string                   `json:"api_key,omitempty"`
	ApiKeyViaUrl       bool                      `json:"api_key_via_url"`
}

// new account info
// used to return info about valid options when making a new account
type newCertOptions struct {
	AvailableKeys             []private_keys.Key      `json:"private_keys"`
	AvailableAccounts         []acme_accounts.Account `json:"acme_accounts"`
	AvailableChallengeMethods []challenges.Method     `json:"challenge_methods"`
}

// NewOrderPayload creates the appropriate newOrder payload for ACME
func (cert *Certificate) NewOrderPayload() acme.NewOrderPayload {
	var identifiers []acme.Identifier

	// subject is always required and should be first
	// dns is the only supported type and is hardcoded
	identifiers = append(identifiers, acme.Identifier{Type: "dns", Value: *cert.Subject})

	// add alt names if they exist
	if cert.SubjectAltNames != nil {
		for _, name := range *cert.SubjectAltNames {
			identifiers = append(identifiers, acme.Identifier{Type: "dns", Value: name})
		}
	}

	return acme.NewOrderPayload{
		Identifiers: identifiers,
	}
}
