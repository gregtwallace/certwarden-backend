package certificates

import (
	"legocerthub-backend/pkg/domain/acme_accounts"
	"legocerthub-backend/pkg/domain/certificates/challenges"
	"legocerthub-backend/pkg/domain/private_keys"
)

// a single certificate
type Certificate struct {
	ID                 *int                        `json:"id"`
	Name               *string                     `json:"name"`
	Description        *string                     `json:"description"`
	PrivateKey         *private_keys.Key           `json:"private_key,omitempty"`
	AcmeAccount        *acme_accounts.Account      `json:"acme_account,omitempty"`
	ChallengeType      *challenges.ChallengeMethod `json:"challenge_type,omitempty"`
	Subject            *string                     `json:"subject,omitempty"`
	SubjectAltNames    *[]string                   `json:"subject_alts,omitempty"`
	CommonName         *string                     `json:"common_name,omitempty"`
	Organization       *string                     `json:"organization,omitempty"`
	OrganizationalUnit *string                     `json:"organizational_unit,omitempty"`
	Country            *string                     `json:"country,omitempty"`
	State              *string                     `json:"state,omitempty"`
	City               *string                     `json:"city,omitempty"`
	CreatedAt          *int                        `json:"created_at,omitempty"`
	UpdatedAt          *int                        `json:"updated_at,omitempty"`
	ApiKey             *string                     `json:"api_key,omitempty"`
	Pem                *string                     `json:"pem,omitempty"`
	ValidFrom          *int                        `json:"valid_from,omitempty"`
	ValidTo            *int                        `json:"valid_to,omitempty"`
}

// new account info
// used to return info about valid options when making a new account
type newCertOptions struct {
	AvailableKeys     []private_keys.Key      `json:"private_keys"`
	AvailableAccounts []acme_accounts.Account `json:"acme_accounts"`
}
