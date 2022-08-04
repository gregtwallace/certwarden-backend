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

		// acmeOrder to hold the Order responses and to later update storage
		var acmeOrder acme.Order

		// Use loop to retry order if "processing". Cap retries to avoid indefinite loop.
	fulfillLoop:
		for i := 1; i <= 5; i++ {
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
				var authStatus string
				authStatus, err = service.authorizations.FulfillAuths(acmeOrder.Authorizations, *order.Certificate.ChallengeMethod, key, *order.Certificate.AcmeAccount.IsStaging)
				if err != nil {
					service.logger.Error(err)
					return // done, failed
				}
				if authStatus != "valid" {
					break // PaG order again and to get final ('invalid' Status) version to save to storage
				}

				// auths were valid, fallthrough to "ready" (which order should now be in)
				fallthrough

			case "ready": // needs to be finalized
				// TODO: finalize
				// TODO: if err break switch to try again

				// loop around, "certificate" field isn't present until the order is valid so another PaG is needed
				break fulfillLoop // TODO REMOVE break fulfillLoop

			case "valid": // can be downloaded
				// TODO: download? update related cert?
				// TODO: if err break switch to try again

				break fulfillLoop

			case "processing":
				// TODO: Implement exponential backoff
				time.Sleep(time.Duration(i) * 30 * time.Second)

			case "invalid": // break, irrecoverable
				service.logger.Debugf("order status invalid; acme error: %s", acmeOrder.Error)
				break fulfillLoop

			// Note: there is no 'expired' Status case. If the order expires it simply moves to 'invalid'

			default:
				service.logger.Error(errors.New("order status unknown"))
				return // done, failed (don't update db since something anomalous happened)
			}
		}

		// TODO: TEMP: REMOVE
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
		// END TODO TEMP REMOVE

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
