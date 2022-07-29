package orders

import (
	"legocerthub-backend/pkg/validation"
)

// isIdExistingMatch checks that the cert and order are both valid and in storeage
// and then also verifies the order belongs to the cert.  Returns an error if not
// valid, nil if valid
func (service *Service) isIdExistingMatch(certId int, orderId int) (err error) {
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

	// check the cert id on the order matches the cert
	if *cert.ID != *order.Certificate.ID {
		return validation.ErrOrderMismatch
	}

	return nil
}
