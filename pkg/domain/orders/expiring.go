package orders

import (
	"database/sql"
	"legocerthub-backend/pkg/acme"
	"time"
)

// reorderTime: If less than this duration of time remaining on cert,
// LeGo Certhub will try to obtain a newer cert.
// TODO: Move to customizable setting in config.yaml and frontent->settings
const reorderTime = 40 * (24 * time.Hour)

// orderExpiringCerts automatically orders any certficates that are within
// the specified expiration window.
func (service *Service) orderExpiringCerts() (err error) {
	// get orders relating to all currently valid cers
	orders, err := service.storage.GetAllValidCurrentOrders()
	if err != nil {
		return err
	}

	// for each order, check expiration
	for i := range orders {
		// calculate remaining time on order's cert
		expUnix := time.Unix(int64(*orders[i].ValidTo), 0)
		// if less than reorderTime, order a new one
		if time.Until(expUnix) < reorderTime {
			// refresh
			err = service.refreshCert(*orders[i].Certificate.ID)
			if err != nil {
				// log error, but keep going through remaining range
				service.logger.Errorf("failed to refresh cert (%d): %s", *orders[i].Certificate.ID, err)
			}
			// sleep a little so slew of new orders doesn't hit ACME all at once
			time.Sleep(15 * time.Second)
		}
	}

	return nil
}

// refreshCert retries an existing pending order for the specified
// cert or if there is no pending new order it places a new order
// for the specified cert
func (service *Service) refreshCert(certId int) (err error) {
	// check for an existing incomplete order
	orderId, err := service.storage.GetNewestIncompleteCertOrderId(certId)
	if err == sql.ErrNoRows {
		// no-op, skip down to new order
	} else if err != nil {
		// any other error
		return err
	} else {
		// no error, retry existing order
		service.logger.Debugf("refreshing cert (%d): retrying order %d", certId, orderId)
		// kickoff order fulfillment (low priority) (async)
		err = service.orderFromAcme(orderId, false)
		if err != nil {
			return err
		}
		// done
		return nil
	}

	// if there was no existing incomplete order, place a new order
	service.logger.Debugf("refreshing cert (%d): placing new order", certId)

	// fetch the relevant cert
	cert, err := service.storage.GetOneCertById(certId, true)
	if err != nil {
		return err
	}

	// get account key
	key, err := cert.AcmeAccount.AccountKey()
	if err != nil {
		return err
	}

	// send the new-order to ACME
	var acmeResponse acme.Order
	if *cert.AcmeAccount.IsStaging {
		acmeResponse, err = service.acmeStaging.NewOrder(cert.NewOrderPayload(), key)
	} else {
		acmeResponse, err = service.acmeProd.NewOrder(cert.NewOrderPayload(), key)
	}
	if err != nil {
		return err
	}
	service.logger.Debugf("new order location: %s", acmeResponse.Location)

	// save ACME response to order storage
	newOrderId, err := service.storage.PostNewOrder(cert, acmeResponse)
	if err != nil {
		return err
	}

	// update certificate timestamp
	err = service.storage.UpdateCertUpdatedTime(certId)
	if err != nil {
		service.logger.Error(err)
	}

	// kickoff order fulfillment (low priority) (async)
	err = service.orderFromAcme(newOrderId, false)
	if err != nil {
		service.logger.Error(err)
	}

	return nil
}
