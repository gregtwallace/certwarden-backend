package orders

import (
	"errors"
	"legocerthub-backend/pkg/acme"
)

// fulfill launches a go routine to fulfill the specified order
func (service *Service) fulfill(orderId int) {
	go func(orderId int) {
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

			// TODO (to consider): use fallthrough to go from pending -> ready -> valid
			// Alternately, use a loop and after each status should be 'complete', PaG again and check new status

			case "pending": // needs to be authed
				err = service.authorizations.FulfillAuthz(acmeOrder.Authorizations, *order.Certificate.ChallengeMethod, key, *order.Certificate.AcmeAccount.IsStaging)
				if err != nil {
					service.logger.Error(err)
					return // done, failed
				}

				fallthrough

			case "ready": // needs to be finalized
				// TODO: finalize
				// TODO: download cert and store it

				fallthrough

			case "valid": // can be downloaded
				// TODO: do nothing? or redownload?
				break fulfillLoop

			case "processing":
				// TODO: wait and try again, ACME server is doing stuff

			case "invalid": // nothing to do, irrecoverable
				if acmeOrder.Error != nil {
					service.logger.Error(acmeOrder.Error)
				} else {
					service.logger.Error(errors.New("order status: invalid (no acme error provided)"))
				}
				break fulfillLoop

			default:
				service.logger.Error(errors.New("unknown order status"))
				break fulfillLoop
			}
		}

		// update order in storage
		err = service.storage.UpdateOrderAcme(orderId, acmeOrder)
		if err != nil {
			service.logger.Error(err)
		}

		// TODO: Update cert (including updated timestamp)

		service.logger.Debugf("end of order fulfiller")

	}(orderId)
}
