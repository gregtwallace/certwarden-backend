package orders

import (
	"errors"
	"legocerthub-backend/pkg/acme"
	"time"
)

// orderFromAcme launches a go routine to fulfill the specified order
func (service *Service) orderFromAcme(orderId int) (err error) {
	// add to working to indicate order is being worked
	err = service.working.add(orderId)
	// error indicates already being worked
	if err != nil {
		return err
	}

	// async go routine to try and fulfill order and then remove the order
	// from working after it is complete
	go func(orderId int, service *Service) {
		// remove id from working when goroutine is done
		defer func(orderId int, service *Service) {
			err := service.working.remove(orderId)
			if err != nil {
				service.logger.Error(err)
			}
			service.logger.Debugf("end of order fulfiller")
		}(orderId, service)

		// fetch the relevant order
		order, err := service.storage.GetOneOrder(orderId)
		if err != nil {
			service.logger.Error(err)
			return // done, failed
		}

		// fetch the certificate with sensitive data and update the order object
		*order.Certificate, err = service.storage.GetOneCertById(*order.Certificate.ID, true)
		if err != nil {
			service.logger.Error(err)
			return // done, failed
		}

		// get account key
		key, err := order.Certificate.AcmeAccount.AccountKey()
		if err != nil {
			service.logger.Error(err)
			return // done, failed
		}

		var acmeOrder acme.Order

		// use loop to retry order fulfillment as appropriate
		// set max tries to avoid infinite loop if something anomalous occurs
	fulfillLoop:
		for i := 1; i <= 3; i++ {
			// PaG the order (to get most recent status)
			if *order.Certificate.AcmeAccount.IsStaging {
				acmeOrder, err = service.acmeStaging.GetOrder(*order.Location, key)
			} else {
				acmeOrder, err = service.acmeProd.GetOrder(*order.Location, key)
			}
			if err != nil {
				service.logger.Error(err)
				return // done, failed
			}

			// action depends on order's current Status
			switch acmeOrder.Status {
			case "pending": // needs to be authed
				err = service.authorizations.FulfillAuths(acmeOrder.Authorizations, *order.Certificate.ChallengeMethod, key, *order.Certificate.AcmeAccount.IsStaging)
				if err != nil {
					service.logger.Debug(err)
					// TODO: Implement exponential backoff
					time.Sleep(time.Duration(i) * 30 * time.Second)
					break // break switch to try order again
				}

				fallthrough

			case "ready": // needs to be finalized
				// TODO: finalize
				// TODO: if err break switch to try again

				fallthrough

			case "valid": // can be downloaded
				// TODO: download? update related cert?
				// TODO: if err break switch to try again

				break fulfillLoop

			case "processing":
				// TODO: Implement exponential backoff
				time.Sleep(time.Duration(i) * 30 * time.Second)

			case "invalid": // break, irrecoverable
				if acmeOrder.Error != nil {
					service.logger.Error(acmeOrder.Error)
				} else {
					service.logger.Error(errors.New("order status invalid (no acme error provided)"))
				}
				break fulfillLoop

			// Note: there is no 'expired' Status case. If the order expires it simply moves to 'invalid'

			default:
				service.logger.Error(errors.New("order status unknown"))
				return // done, failed (don't update db since something anomalous happened)
			}
		}

		// update order in storage
		err = service.storage.UpdateOrderAcme(orderId, acmeOrder)
		if err != nil {
			service.logger.Error(err)
		}

		// TODO: Update cert (including updated timestamp)
		// Or should Get Cert just figure out the most recent valid order and act accordingly?

	}(orderId, service)

	return nil
}
