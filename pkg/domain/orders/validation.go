package orders

import (
	"legocerthub-backend/pkg/validation"
	"time"
)

// isCertOrderMatch confirms the order and cert are both valid and that
// the order belongs to the specified cert. It also returns the order
// in case additional validation is needed.
func (service *Service) isCertOrderMatch(certId int, orderId int) (order Order, err error) {
	// basic check
	err = validation.IsIdExisting(&certId)
	if err != nil {
		return Order{}, err
	}
	err = validation.IsIdExisting(&orderId)
	if err != nil {
		return Order{}, err
	}

	// check that cert exists in storage
	cert, err := service.storage.GetOneCertById(certId, false)
	if err != nil {
		return Order{}, err
	}

	// check that order exists in storage
	order, err = service.storage.GetOneOrder(orderId)
	if err != nil {
		return Order{}, err
	}

	// check for nil pointers (in the event a deletion has NULLed an id)
	if cert.ID == nil || order.Certificate == nil || order.Certificate.ID == nil {
		return Order{}, validation.ErrOrderMismatch
	}

	// check the cert id on the order matches the cert
	if *cert.ID != *order.Certificate.ID {
		return Order{}, validation.ErrOrderMismatch
	}

	return order, nil
}

// isOrderRetryable checks that the cert and order are both valid and in storage
// and then also verifies the order belongs to the cert.  Returns an error if not
// valid, nil if valid
func (service *Service) isOrderRetryable(certId int, orderId int) (err error) {
	order, err := service.isCertOrderMatch(certId, orderId)
	if err != nil {
		return err
	}

	// check order is in a state that can be retried
	// if pem is missing, allow proceeding so pem can be acquired
	// by order fulfiller
	validPem := false
	if order.Pem != nil && *order.Pem != "" {
		validPem = true
	}

	if *order.Status == "valid" && validPem {
		return validation.ErrOrderValid
	} else if *order.Status == "invalid" {
		return validation.ErrOrderInvalid
	}

	return nil
}

// isOrderRevocable verifies order belongs to cert and confirms the order
// is in a state that can be revoked ('valid' and 'valid_to' < current time)
func (service *Service) isOrderRevocable(certId, orderId int) (err error) {
	order, err := service.isCertOrderMatch(certId, orderId)
	if err != nil {
		return err
	}

	// check order is in a state that can be revoked
	// must be valid and not expired
	if *order.Status == "valid" && !*order.KnownRevoked && int(time.Now().Unix()) < *order.ValidTo {
		return nil
	}

	return validation.ErrOrderNotRevocable
}
