package orders

import (
	"errors"
	"legocerthub-backend/pkg/output"
)

// placeNewOrderAndFulfill creates a new ACME order for the specified Certificate ID,
// and prioritizes the order as specified. It returns the new orderId.
func (service *Service) placeNewOrderAndFulfill(certId int, highPriority bool) (orderId int, err error) {
	// get cert
	cert, err := service.certificates.GetCertificate(certId)
	if err != nil {
		return -2, err
	}

	// get account key
	key, err := cert.CertificateAccount.AcmeAccountKey()
	if err != nil {
		service.logger.Error(err)
		return -2, output.ErrInternal
	}

	// send the new-order to ACME
	acmeService, err := service.acmeServerService.AcmeService(cert.CertificateAccount.AcmeServer.ID)
	if err != nil {
		service.logger.Error(err)
		return -2, output.ErrInternal
	}

	acmeResponse, err := acmeService.NewOrder(cert.NewOrderPayload(), key)
	if err != nil {
		service.logger.Error(err)
		return -2, output.ErrInternal
	}
	service.logger.Debugf("new order location: %s", acmeResponse.Location)

	// populate new order payload
	payload := makeNewOrderAcmePayload(cert, acmeResponse)

	// save ACME response to order storage
	orderId, err = service.storage.PostNewOrder(payload)
	// if exists error, try to update an existing order
	if errors.Is(err, ErrOrderExists) {
		err = service.storage.PutOrderAcme(makeUpdateOrderAcmePayload(orderId, acmeResponse))
		if err != nil {
			service.logger.Error(err)
			return -2, output.ErrStorageGeneric
		}
	} else if err != nil {
		service.logger.Error(err)
		return -2, output.ErrStorageGeneric
	}

	// update certificate timestamp
	err = service.storage.UpdateCertUpdatedTime(cert.ID)
	if err != nil {
		service.logger.Error(err)
		// no return
	}

	// kickoff order fulfillment (async)
	err = service.orderFromAcme(orderId, highPriority)
	// log error if something strange happened
	if err != nil {
		service.logger.Error(err)
		// no return
	}

	return orderId, nil
}
