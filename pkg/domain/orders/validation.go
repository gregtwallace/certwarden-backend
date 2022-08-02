package orders

import (
	"legocerthub-backend/pkg/validation"
)

// isIdExistingMatch checks that the cert and order are both valid and in storeage
// and then also verifies the order belongs to the cert.  Returns an error if not
// valid, nil if valid
func (service *Service) isOrderRetryable(certId int, orderId int) (err error) {
	// basic check
	err = validation.IsIdExisting(&certId)
	if err != nil {
		return err
	}
	err = validation.IsIdExisting(&orderId)
	if err != nil {
		return err
	}

	// check that cert exists in storage
	cert, err := service.storage.GetOneCertById(certId, false)
	if err != nil {
		return err
	}

	// check that order exists in storage
	order, err := service.storage.GetOneOrder(orderId)
	if err != nil {
		return err
	}

	// check for nil pointers (in the event a deletion has NULLed an id)
	if cert.ID == nil || order.Certificate == nil || order.Certificate.ID == nil {
		return validation.ErrOrderMismatch
	}

	// check the cert id on the order matches the cert
	if *cert.ID != *order.Certificate.ID {
		return validation.ErrOrderMismatch
	}

	// check order is in a state that can be retried
	if *order.Status == "invalid" {
		return validation.ErrOrderInvalid
	}

	return nil
}
