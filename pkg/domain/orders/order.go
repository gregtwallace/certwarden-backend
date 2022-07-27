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
	FinalizedKey   *private_keys.Key         `json:"private_key,omitempty"`
	Location       *string                   `json:"location,omitempty"`
	Status         *string                   `json:"status,omitempty"`
	Error          *acme.Error               `json:"error,omitempty"`
	Expires        *int                      `json:"expires,omitempty"`
	DnsIdentifiers []string                  `json:"dns_identifiers,omitempty"`
	Authorizations []string                  `json:"authorizations,omitempty"`
	Finalize       *string                   `json:"finalize,omitempty"`
	CertificateUrl *string                   `json:"certificate_url,omitempty"`
	CreatedAt      *int                      `json:"created_at,omitempty"`
	UpdatedAt      *int                      `json:"updated_at,omitempty"`
}

// Finalize       string          `json:"finalize"`
// Certificate    string          `json:"certificate,omitempty"`
// NotBefore      acmeTimeString  `json:"notBefore,omitempty"`
// NotAfter       acmeTimeString  `json:"notAfter,omitempty"`

// makeNewOrder creates an order from the specified cert and acme response
func makeNewOrder(cert *certificates.Certificate, acmeOrder *acme.Order) (order Order) {
	order.Certificate = cert

	order.Location = &acmeOrder.Location
	order.Status = &acmeOrder.Status
	order.Error = acmeOrder.Error
	order.Expires = acmeOrder.Expires.ToUnixTime()
	order.DnsIdentifiers = acmeOrder.Identifiers.DnsIdentifiers()
	order.Authorizations = acmeOrder.Authorizations
	order.Finalize = &acmeOrder.Finalize
	order.CertificateUrl = &acmeOrder.Certificate

	return order
}
