package orders

import (
	"errors"
	"legocerthub-backend/pkg/acme"
	"legocerthub-backend/pkg/output"
	"log"
)

func (service *Service) placeNewOrderAndFulfill(certId int, highPriority bool) (orderId int, err error) {
	// fetch the relevant cert
	cert, err := service.storage.GetOneCertById(certId)
	if err != nil {
		service.logger.Error(err)
		return -2, output.ErrStorageGeneric
	}

	// no need to validate, can try to order any cert in storage

	// get account key
	key, err := cert.CertificateAccount.AcmeAccountKey()
	if err != nil {
		service.logger.Error(err)
		return -2, output.ErrInternal
	}

	// send the new-order to ACME
	var acmeResponse acme.Order
	if cert.CertificateAccount.IsStaging {
		acmeResponse, err = service.acmeStaging.NewOrder(cert.NewOrderPayload(), key)
	} else {
		acmeResponse, err = service.acmeProd.NewOrder(cert.NewOrderPayload(), key)
	}
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
	log.Println(err)

	// update certificate timestamp
	err = service.storage.UpdateCertUpdatedTime(certId)
	if err != nil {
		service.logger.Error(err)
	}

	// kickoff order fulfillment (async)
	err = service.orderFromAcme(orderId, highPriority)
	// log error if something strange happened
	if err != nil {
		service.logger.Error(err)
	}

	return orderId, nil
}
