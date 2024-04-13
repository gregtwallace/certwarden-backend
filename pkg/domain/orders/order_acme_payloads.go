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
	AccountId      int
	Status         string
	KnownRevoked   bool
	Expires        int
	DnsIds         []string
	Error          *string
	Authorizations []string
	Finalize       string
	Location       string
	CreatedAt      int
	UpdatedAt      int
}

// newOrderAcmePayload makes a OrderAcmePayload using the specified certificate
// and acme.Response
func makeNewOrderAcmePayload(cert certificates.Certificate, acmeResponse acme.Order) NewOrderAcmePayload {
	acmeErr, err := acmeResponse.Error.MarshalledString()
	if err != nil {
		acmeErr = nil
	}

	payload := NewOrderAcmePayload{
		CertId:         cert.ID,
		AccountId:      cert.CertificateAccount.ID,
		Status:         acmeResponse.Status,
		KnownRevoked:   false,
		Expires:        acmeResponse.Expires.ToUnixTime(),
		DnsIds:         acmeResponse.Identifiers.DnsIdentifiers(),
		Error:          acmeErr,
		Authorizations: acmeResponse.Authorizations,
		Finalize:       acmeResponse.Finalize,
		Location:       acmeResponse.Location,
		CreatedAt:      int(time.Now().Unix()),
		UpdatedAt:      int(time.Now().Unix()),
	}

	return payload
}

// UpdateAcmeOrderPayload is the payload to update storage regarding an existing ACME order
type UpdateAcmeOrderPayload struct {
	Status         string
	Expires        *int
	DnsIds         []string
	Error          *string
	Authorizations []string
	Finalize       string
	CertificateUrl *string
	UpdatedAt      int
	OrderId        int
}

// makeUpdateOrderAcmePayload makes the UpdateAcmeOrderPayload using a new payload and the orderId
func makeUpdateOrderAcmePayload(orderId int, acmeResponse acme.Order) UpdateAcmeOrderPayload {
	acmeErr, err := acmeResponse.Error.MarshalledString()
	if err != nil {
		acmeErr = nil
	}

	return UpdateAcmeOrderPayload{
		Status:         acmeResponse.Status,
		DnsIds:         acmeResponse.Identifiers.DnsIdentifiers(),
		Error:          acmeErr,
		Authorizations: acmeResponse.Authorizations,
		UpdatedAt:      int(time.Now().Unix()),
		OrderId:        orderId,
		CertificateUrl: acmeResponse.Certificate,
	}
}
