package orders

import (
	"legocerthub-backend/pkg/validation"
	"time"
)

// isCertOrderMatch returns true if the order and cert are both valid and the order
// belongs to the specified cert. It also returns the order if valid.
func (service *Service) isCertOrderMatch(certId int, orderId int) (bool, Order) {
	// basic check
	if !validation.IsIdExisting(certId) {
		return false, Order{}
	}
	if !validation.IsIdExisting(orderId) {
		return false, Order{}
	}

	// check that cert exists in storage
	cert, err := service.storage.GetOneCertById(certId)
	if err != nil {
		return false, Order{}
	}

	// check that order exists in storage
	order, err := service.storage.GetOneOrder(orderId)
	if err != nil {
		return false, Order{}
	}

	// check the cert id on the order matches the cert
	if cert.ID != order.Certificate.ID {
		return false, Order{}
	}

	return true, order
}

// isOrderRetryable returns true if the cert and order are both valid and in storage,
// the order belongs to the cert, and the order is in a state that can be retried.
func (service *Service) isOrderRetryable(certId int, orderId int) bool {
	match, order := service.isCertOrderMatch(certId, orderId)
	if !match {
		return false
	}

	// check order is in a state that can be retried
	return !(order.Status == "valid" || order.Status == "invalid")
}

// isOrderRevocable verifies order belongs to cert and confirms the order
// is in a state that can be revoked ('valid' and 'valid_to' < current time)
func (service *Service) isOrderRevocable(certId, orderId int) bool {
	match, order := service.isCertOrderMatch(certId, orderId)
	if !match {
		return false
	}

	// check order is in a state that can be revoked
	// nil check
	if order.ValidTo == nil {
		return false
	}

	// must be valid and not expired
	return (order.Status == "valid" && !order.KnownRevoked && int(time.Now().Unix()) < *order.ValidTo)
}
