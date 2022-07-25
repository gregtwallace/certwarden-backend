package orders

import (
	"legocerthub-backend/pkg/acme"
	"legocerthub-backend/pkg/domain/certificates"
)

// a single order
type Order struct {
	Acme        *acme.Order
	ID          *int                      `json:"id"`
	Certificate *certificates.Certificate `json:"certificate,omitempty"`
	CreatedAt   *int                      `json:"created_at,omitempty"`
	UpdatedAt   *int                      `json:"updated_at,omitempty"`
}

// makeNewOrder creates an order from the specified cert and acme response
func makeNewOrder(cert *certificates.Certificate, acmeOrder *acme.Order) (order Order) {
	order.Acme = acmeOrder
	order.Certificate = cert

	return order
}
