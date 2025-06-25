package orders

import (
	"bytes"
	"certwarden-backend/pkg/acme"
	"certwarden-backend/pkg/domain/certificates"
	"certwarden-backend/pkg/domain/private_keys"
	"certwarden-backend/pkg/pagination_sort"
	"certwarden-backend/pkg/storage"
	"crypto/x509"
	"encoding/pem"
	"errors"
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
	ValidFrom      *time.Time
	ValidTo        *time.Time
	ChainRootCN    *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Profile        *string
	RenewalInfo    *renewalInfo
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
	Profile           *string                         `json:"profile,omitempty"`
	RenewalInfo       *renewalInfo                    `json:"renewal_info"`
	CreatedAt         int64                           `json:"created_at"`
	UpdatedAt         int64                           `json:"updated_at"`
}

type orderCertificateSummaryResponse struct {
	ID                 int                                    `json:"id"`
	Name               string                                 `json:"name"`
	CertificateAccount orderCertificateAccountSummaryResponse `json:"acme_account"`
	Subject            string                                 `json:"subject"`
	SubjectAltNames    []string                               `json:"subject_alts"`
	ApiKeyViaUrl       bool                                   `json:"api_key_via_url"`
	LastAccess         int64                                  `json:"last_access"`
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

	var validFromUnix *int
	if order.ValidFrom != nil {
		validFromUnixVal := int(order.ValidFrom.Unix())
		validFromUnix = &validFromUnixVal
	}

	var validToUnix *int
	if order.ValidTo != nil {
		validToUnixVal := int(order.ValidTo.Unix())
		validToUnix = &validToUnixVal
	}

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
			Subject:         order.Certificate.Subject,
			SubjectAltNames: order.Certificate.SubjectAltNames,
			ApiKeyViaUrl:    order.Certificate.ApiKeyViaUrl,
			LastAccess:      order.Certificate.LastAccess.Unix(),
		},
		Status:         order.Status,
		KnownRevoked:   order.KnownRevoked,
		Error:          order.Error,
		DnsIdentifiers: order.DnsIdentifiers,
		FinalizedKey:   finalKey,
		ValidFrom:      validFromUnix,
		ValidTo:        validToUnix,
		ChainRootCN:    order.ChainRootCN,
		Profile:        order.Profile,
		RenewalInfo:    order.RenewalInfo,
		CreatedAt:      order.CreatedAt.Unix(),
		UpdatedAt:      order.UpdatedAt.Unix(),
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
	// only available if finalized, ensure nil check
	keyModtime := time.Time{}
	if order.FinalizedKey != nil {
		keyModtime = order.FinalizedKey.UpdatedAt
	}

	// return latest of the three
	if keyModtime.After(order.UpdatedAt) && keyModtime.After(order.Certificate.UpdatedAt) {
		return keyModtime
	}

	if order.Certificate.UpdatedAt.After(order.UpdatedAt) {
		return order.Certificate.UpdatedAt
	}

	return order.UpdatedAt
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
	if order.Certificate.PostProcessingClientAddress != "" && order.Certificate.PostProcessingClientKeyB64 != "" {
		return true
	}

	// post processing action = script or binary
	if order.Certificate.PostProcessingCommand != "" {
		return true
	}

	return false
}

// NewOrderPayload creates the appropriate newOrder payload for ACME
func (service *Service) NewOrderPayload(cert certificates.Certificate) acme.NewOrderPayload {
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

	// ACME Profile Extension: only send profile if it isn't blank
	p := &cert.Profile
	if cert.Profile == "" {
		p = nil
	}

	// ACME ARI Extension: try to include the `replaces` field
	replaces := func() *string {
		acmeServ, err := service.acmeServerService.AcmeService(cert.CertificateAccount.AcmeServer.ID)
		if err != nil {
			service.logger.Errorf("orders: new order cant populated `replaces`, failed to get acme service for cert %d (%s)", cert.ID, err)
			return nil
		}

		// ARI extension not supported
		if !acmeServ.SupportsARIExtension() {
			return nil
		}

		// get all orders from storage
		orders, _, err := service.storage.GetOrdersByCert(cert.ID, pagination_sort.Query{})
		if err != nil {
			// no records isn't an error worth logging
			if !errors.Is(err, storage.ErrNoRecord) {
				service.logger.Error(err)
			}
			return nil
		}

		// found orders
		var mostRecentValidFrom time.Time
		mostRecentPem := ""
		for i := range orders {
			if orders[i].Pem == nil {
				continue
			}

			if orders[i].ValidFrom != nil && orders[i].ValidFrom.After(mostRecentValidFrom) {
				mostRecentValidFrom = *orders[i].ValidFrom
				mostRecentPem = *orders[i].Pem
			}
		}

		// no order PEM to use for replaces
		if mostRecentPem == "" {
			return nil
		}

		// use the order's pem to create the unique ID for `replaces`
		certBlock, _ := pem.Decode([]byte(mostRecentPem))
		if certBlock == nil {
			service.logger.Errorf("orders: new order cant populated `replaces`, cert pem block of latest order of cert %d is nil", cert.ID)
			return nil
		}

		x509Cert, err := x509.ParseCertificate(certBlock.Bytes)
		if err != nil {
			service.logger.Errorf("orders: new order cant populated `replaces`, cert pem block of latest order of cert %d failed to parse (%s)", cert.ID, err)
			return nil
		}

		r := new(string)
		*r = acme.ACMERenewalInfoIdentifier(x509Cert)

		return r
	}()

	return acme.NewOrderPayload{
		Identifiers: identifiers,
		Profile:     p,
		Replaces:    replaces,
	}
}
