package orders

import (
	"certwarden-backend/pkg/output"
	"database/sql"
	"errors"
	"time"
)

// placeNewOrderAndFulfill creates a new ACME order for the specified Certificate ID,
// and prioritizes the order as specified. It returns the new orderId.
func (service *Service) placeNewOrderAndFulfill(certId int, highPriority bool) (Order, *output.JsonError) {
	// dont allow new order if a pending order exists
	orderId, err := service.storage.GetNewestIncompleteCertOrderId(certId)
	//nolint:gocritic // TODO: This code needs a refactor anyway, so leave the logic as-is for now
	if errors.Is(err, sql.ErrNoRows) {
		// no existing incomplete order, make a new one

		// get cert
		cert, outErr := service.certificates.GetCertificate(certId)
		if outErr != nil {
			return Order{}, outErr
		}

		// get account key
		key, err := cert.Account.AcmeAccountKey()
		if err != nil {
			service.logger.Error(err)
			return Order{}, output.ErrorJsonErrInternal(err)
		}

		// send the new-order to ACME
		acmeService, err := service.acmeServerService.AcmeService(cert.Account.AcmeServer.ID)
		if err != nil {
			service.logger.Error(err)
			return Order{}, output.ErrorJsonErrInternal(err)
		}

		acmeResponse, err := acmeService.NewOrder(service.NewOrderPayload(cert), key)
		if err != nil {
			service.logger.Error(err)
			return Order{}, output.ErrorJsonErrInternal(err)
		}
		service.logger.Debugf("orders: new order location: %s", acmeResponse.Location)

		// populate new order payload
		payload := makeNewOrderAcmePayload(cert, &acmeResponse)

		// save ACME response to order storage
		orderId, err = service.storage.PostNewOrder(&payload)
		// if exists error, try to update an existing order
		if errors.Is(err, ErrOrderExists) {
			err = service.storage.PutOrderACME(makeUpdateOrderAcmePayload(orderId, &acmeResponse))
			if err != nil {
				service.logger.Error(err)
				return Order{}, output.ErrorJsonErrStorageGeneric(err)
			}
		} else if err != nil {
			service.logger.Error(err)
			return Order{}, output.ErrorJsonErrStorageGeneric(err)
		}

		// update certificate timestamp
		err = service.storage.PutCertUpdatedAt(cert.ID, time.Now())
		if err != nil {
			service.logger.Error(err)
			// no return
		}
	} else if err != nil {
		// some other unexpected storage error
		service.logger.Error(err)
		return Order{}, output.ErrorJsonErrStorageGeneric(err)
	} else {
		// continue with existing order
		service.logger.Debugf("orders: create new order is retrying existing pending order instead of creating new")
	}

	// kickoff order fulfillment (async)
	err = service.fulfillOrder(orderId, highPriority)
	// log error if something strange happened
	if err != nil {
		service.logger.Error(err)
		// no return
	}

	// get new order from db to return
	newOrder, outErr := service.getOrder(certId, orderId)
	if outErr != nil {
		return Order{}, outErr
	}

	return newOrder, nil
}
