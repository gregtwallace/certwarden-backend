package orders

import (
	"legocerthub-backend/pkg/acme"
	"legocerthub-backend/pkg/domain/certificates"
	"legocerthub-backend/pkg/domain/private_keys"
)

// Order is a single ACME order object
// Finalized key is included as the cert may change keys after an order is finalized.
type Order struct {
	ID             int
	Certificate    certificates.Certificate
	Location       string
	Status         string
	KnownRevoked   bool
	Error          *acme.Error
	Expires        *int
	DnsIdentifiers []string
	Authorizations []string
	Finalize       string
	FinalizedKey   *private_keys.Key
	CertificateUrl *string
	Pem            *string
	ValidFrom      *int
	ValidTo        *int
	CreatedAt      int
	UpdatedAt      int
}

// orderSummaryResponse is a JSON response containing only
// fields desired for the summary
type orderSummaryResponse struct {
	ID             int                             `json:"id"`
	Certificate    orderCertificateSummaryResponse `json:"certificate"`
	Status         string                          `json:"status"`
	KnownRevoked   bool                            `json:"known_revoked"`
	Error          *acme.Error                     `json:"error"`
	DnsIdentifiers []string                        `json:"dns_identifiers"`
	FinalizedKey   *orderKeySummaryResponse        `json:"finalized_key"`
	ValidFrom      *int                            `json:"valid_from"`
	ValidTo        *int                            `json:"valid_to"`
	CreatedAt      int                             `json:"created_at"`
	UpdatedAt      int                             `json:"updated_at"`
}

type orderCertificateSummaryResponse struct {
	ID                 int                                    `json:"id"`
	Name               string                                 `json:"name"`
	CertificateAccount orderCertificateAccountSummaryResponse `json:"acme_account"`
	Subject            string                                 `json:"subject"`
	SubjectAltNames    []string                               `json:"subject_alts"`
}

type orderCertificateAccountSummaryResponse struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	IsStaging bool   `json:"is_staging"`
}

type orderKeySummaryResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (order Order) summaryResponse() orderSummaryResponse {
	// depends on if FinalizedKey is set yet
	var finalKey *orderKeySummaryResponse
	if order.FinalizedKey != nil {
		finalKey = &orderKeySummaryResponse{
			ID:   order.FinalizedKey.ID,
			Name: order.FinalizedKey.Name,
		}
	}

	return orderSummaryResponse{
		ID: order.ID,
		Certificate: orderCertificateSummaryResponse{
			ID:   order.Certificate.ID,
			Name: order.Certificate.Name,
			CertificateAccount: orderCertificateAccountSummaryResponse{
				ID:        order.Certificate.CertificateAccount.ID,
				Name:      order.Certificate.CertificateAccount.Name,
				IsStaging: order.Certificate.CertificateAccount.IsStaging,
			},
			Subject:         order.Certificate.Subject,
			SubjectAltNames: order.Certificate.SubjectAltNames,
		},
		Status:         order.Status,
		KnownRevoked:   order.KnownRevoked,
		Error:          order.Error,
		DnsIdentifiers: order.DnsIdentifiers,
		FinalizedKey:   finalKey,
		ValidFrom:      order.ValidFrom,
		ValidTo:        order.ValidTo,
		CreatedAt:      order.CreatedAt,
		UpdatedAt:      order.UpdatedAt,
	}
}
