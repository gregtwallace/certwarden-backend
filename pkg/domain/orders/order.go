package orders

import (
	"bytes"
	"certwarden-backend/pkg/acme"
	"certwarden-backend/pkg/domain/certificates"
	"certwarden-backend/pkg/domain/private_keys"
	"encoding/pem"
	"fmt"
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
	ChainRootCN    *string
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
	ChainRootCN       *string                         `json:"chain_root_cn"`
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

func (order Order) summaryResponse(service *Service) orderSummaryResponse {
	// depends on if FinalizedKey is set yet
	var finalKey *orderKeySummaryResponse
	if order.FinalizedKey != nil {
		finalKey = &orderKeySummaryResponse{
			ID:   order.FinalizedKey.ID,
			Name: order.FinalizedKey.Name,
		}
	}

	// check if job is in queue (priority is irrelevant for checking if exists, so just use false)
	// should never error, so ignore err
	fulfillJob, _ := service.makeFulfillingJob(order.ID, false)
	fulfillingWorker := service.orderFulfilling.JobExists(fulfillJob)

	return orderSummaryResponse{
		FulfillmentWorker: fulfillingWorker,
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
		ChainRootCN:    order.ChainRootCN,
		CreatedAt:      order.CreatedAt,
		UpdatedAt:      order.UpdatedAt,
	}
}

// Output Methods

func (order Order) FilenameNoExt() string {
	return fmt.Sprintf("%s.cert", order.Certificate.Name)
}

func (order Order) PemContent() string {
	// if Pem is nil, return empty
	if order.Pem == nil {
		return ""
	}

	return *order.Pem
}

func (order Order) Modtime() time.Time {
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

// end Output Methods

// hasPostProcessingToDo returns if a given order object is configured in a way
// that involves one or more post processing actions
func (order *Order) hasPostProcessingToDo() bool {
	// post processing action = Client
	if order.Certificate.PostProcessingClientKeyB64 != "" {
		return true
	}

	// post processing action = script or binary
	if order.Certificate.PostProcessingCommand != "" {
		return true
	}

	return false
}
