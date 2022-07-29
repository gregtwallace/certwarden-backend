package orders

import (
	"legocerthub-backend/pkg/acme"
	"legocerthub-backend/pkg/domain/certificates"
	"legocerthub-backend/pkg/domain/private_keys"
)

// Order is a single order object
// Note: acme account is excluded even though stored in storage as it can be deduced through
// the Cert member. This is to allow storage to easily query for orders associated with an
// account.  Finalized key is included as the cert may change keys after an order is finalized.
type Order struct {
	ID             *int                      `json:"id,omitempty"`
	Certificate    *certificates.Certificate `json:"certificate,omitempty"`
	Location       *string                   `json:"location,omitempty"`
	Status         *string                   `json:"status,omitempty"`
	Error          *acme.Error               `json:"error,omitempty"`
	Expires        *int                      `json:"expires,omitempty"`
	DnsIdentifiers []string                  `json:"dns_identifiers,omitempty"`
	Authorizations []string                  `json:"authorizations,omitempty"`
	Finalize       *string                   `json:"finalize,omitempty"`
	FinalizedKey   *private_keys.Key         `json:"finalized_key,omitempty"`
	CertificateUrl *string                   `json:"certificate_url,omitempty"`
	CreatedAt      *int                      `json:"created_at,omitempty"`
	UpdatedAt      *int                      `json:"updated_at,omitempty"`
}
