package orders

import (
	"certwarden-backend/pkg/acme"
	"certwarden-backend/pkg/domain/certificates"
	"errors"
	"time"
)

var ErrOrderExists = errors.New("order location already in storage")

// NewOrderAcmePayload is a struct for posting new and updated Order info
// to storage. This is used when ACME returns info about an Order.
type NewOrderAcmePayload struct {
	CertId         int
	Status         string
	KnownRevoked   bool
	Expires        *time.Time
	DnsIds         []string
	Error          *acme.Error
	Authorizations []string
	Finalize       string
	Profile        *string
	Location       string
	CreatedAt      int
	UpdatedAt      int
}

// newOrderAcmePayload makes a OrderAcmePayload using the specified certificate
// and acme.Response
func makeNewOrderAcmePayload(cert *certificates.Certificate, acmeResponse *acme.Order) NewOrderAcmePayload {
	return NewOrderAcmePayload{
		CertId:         cert.ID,
		Status:         acmeResponse.Status,
		KnownRevoked:   false,
		Expires:        acmeResponse.Expires,
		DnsIds:         acmeResponse.Identifiers.DnsIdentifiers(),
		Error:          acmeResponse.Error,
		Authorizations: acmeResponse.Authorizations,
		Finalize:       acmeResponse.Finalize,
		Profile:        acmeResponse.Profile,
		Location:       acmeResponse.Location,
		CreatedAt:      int(time.Now().Unix()),
		UpdatedAt:      int(time.Now().Unix()),
	}
}

// UpdateAcmeOrderPayload is the payload to update storage regarding an existing ACME order
type UpdateAcmeOrderPayload struct {
	OrderID        int
	Status         string
	Expires        *time.Time
	DnsIds         []string
	Error          *acme.Error
	Authorizations []string
	Finalize       string
	Profile        *string
	CertificateUrl *string
	UpdatedAt      time.Time
}

// makeUpdateOrderAcmePayload makes the UpdateAcmeOrderPayload using a new payload and the orderId
func makeUpdateOrderAcmePayload(orderID int, acmeResponse *acme.Order) *UpdateAcmeOrderPayload {
	return &UpdateAcmeOrderPayload{
		OrderID:        orderID,
		Status:         acmeResponse.Status,
		Expires:        acmeResponse.Expires,
		DnsIds:         acmeResponse.Identifiers.DnsIdentifiers(),
		Error:          acmeResponse.Error,
		Authorizations: acmeResponse.Authorizations,
		Finalize:       acmeResponse.Finalize,
		Profile:        acmeResponse.Profile,
		CertificateUrl: acmeResponse.Certificate,
		UpdatedAt:      time.Now(),
	}
}
