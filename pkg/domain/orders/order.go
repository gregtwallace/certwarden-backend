package orders

import (
	"bytes"
	"encoding/pem"
	"fmt"
	"legocerthub-backend/pkg/acme"
	"legocerthub-backend/pkg/domain/certificates"
	"legocerthub-backend/pkg/domain/private_keys"
	"time"
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
	FulfillmentWorker *int                            `json:"fulfillment_worker,omitempty"`
	ID                int                             `json:"id"`
	Certificate       orderCertificateSummaryResponse `json:"certificate"`
	Status            string                          `json:"status"`
	KnownRevoked      bool                            `json:"known_revoked"`
	Error             *acme.Error                     `json:"error"`
	DnsIdentifiers    []string                        `json:"dns_identifiers"`
	FinalizedKey      *orderKeySummaryResponse        `json:"finalized_key"`
	ValidFrom         *int                            `json:"valid_from"`
	ValidTo           *int                            `json:"valid_to"`
	CreatedAt         int                             `json:"created_at"`
	UpdatedAt         int                             `json:"updated_at"`
}

type orderCertificateSummaryResponse struct {
	ID                         int                                    `json:"id"`
	Name                       string                                 `json:"name"`
	CertificateAccount         orderCertificateAccountSummaryResponse `json:"acme_account"`
	Subject                    string                                 `json:"subject"`
	SubjectAltNames            []string                               `json:"subject_alts"`
	ApiKeyViaUrl               bool                                   `json:"api_key_via_url"`
	PostProcessingCommand      string                                 `json:"post_processing_command"`
	PostProcessingClientKeyB64 string                                 `json:"post_processing_client_key"`
}

type orderCertificateAccountSummaryResponse struct {
	ID                     int                                          `json:"id"`
	Name                   string                                       `json:"name"`
	OrderCertAccountServer orderCertificateAccountServerSummaryResponse `json:"acme_server"`
}

type orderCertificateAccountServerSummaryResponse struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	IsStaging bool   `json:"is_staging"`
}

type orderKeySummaryResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (order Order) summaryResponse(of *orderFulfiller) orderSummaryResponse {
	// depends on if FinalizedKey is set yet
	var finalKey *orderKeySummaryResponse
	if order.FinalizedKey != nil {
		finalKey = &orderKeySummaryResponse{
			ID:   order.FinalizedKey.ID,
			Name: order.FinalizedKey.Name,
		}
	}

	return orderSummaryResponse{
		FulfillmentWorker: of.checkForOrderId(order.ID),
		ID:                order.ID,
		Certificate: orderCertificateSummaryResponse{
			ID:   order.Certificate.ID,
			Name: order.Certificate.Name,
			CertificateAccount: orderCertificateAccountSummaryResponse{
				ID:   order.Certificate.CertificateAccount.ID,
				Name: order.Certificate.CertificateAccount.Name,
				OrderCertAccountServer: orderCertificateAccountServerSummaryResponse{
					ID:        order.Certificate.CertificateAccount.AcmeServer.ID,
					Name:      order.Certificate.CertificateAccount.AcmeServer.Name,
					IsStaging: order.Certificate.CertificateAccount.AcmeServer.IsStaging,
				},
			},
			Subject:                    order.Certificate.Subject,
			SubjectAltNames:            order.Certificate.SubjectAltNames,
			ApiKeyViaUrl:               order.Certificate.ApiKeyViaUrl,
			PostProcessingCommand:      order.Certificate.PostProcessingCommand,
			PostProcessingClientKeyB64: order.Certificate.PostProcessingClientKeyB64,
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

// Pem Output Methods

// PemFilename returns the filename that should be sent to the client when Order
// is sent to the client in Pem format
func (order Order) PemFilename() string {
	return fmt.Sprintf("%s.cert.pem", order.Certificate.Name)
}

// PemContent returns the actual Pem data of the order or an empty string
// if the Pem does not exist
func (order Order) PemContent() string {
	// if Pem is nil, return empty
	if order.Pem == nil {
		return ""
	}

	return *order.Pem
}

// PemModtime returns the more recent of the time the order was last updated or the
// order's certificate resource was last updated at.
// It is possible for an order updated time to move backward for a "newer" cert resource
// since a newer order could be revoked or a certificate could be renamed. Due to this
// also check the certificate's time stamp.
func (order Order) PemModtime() time.Time {
	orderModtime := time.Unix(int64(order.UpdatedAt), 0)
	certModtime := time.Unix(int64(order.Certificate.UpdatedAt), 0)

	// return later of the two
	if certModtime.After(orderModtime) {
		return certModtime
	}
	return orderModtime
}

// next two not required for output.Pem interface, but are used by `download` pkg

// PemContentNoChain returns the Pem data for the main certificate but discards the remainder
// of the certificate chain
func (order Order) PemContentNoChain() string {
	// decode first cert and drop the rest
	certBlock, _ := pem.Decode([]byte(order.PemContent()))

	// return re-encoded pem
	return string(pem.EncodeToMemory(certBlock))
}

// PemContentChainOnly returns the Pem data for the certificate chain, but not the actual
// main cert
func (order Order) PemContentChainOnly() string {
	// decode the first cert in the chain and discard it
	// this effectively leaves the root chain as the "rest"
	_, chain := pem.Decode([]byte(order.PemContent()))

	// remove any extraneouse chars before the first cert begins (spaces and such)
	beginIndex := bytes.Index(chain, []byte{45}) // ascii code for dash character

	// return pem content
	return string(chain[beginIndex:])
}

// end Pem Output Methods
