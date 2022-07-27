package orders

import (
	"errors"
	"legocerthub-backend/pkg/acme"
)

// FulfillOrder launches a go routine to fulfill the specified order
func (service *Service) FulfillOrder(order Order) {
	go func() {
		// get account key
		key, err := order.Certificate.AcmeAccount.AccountKey()
		if err != nil {
			service.logger.Error(err)
			return
		}

		// PaG the order (to get most recent status)
		var acmeOrder acme.Order
		if *order.Certificate.AcmeAccount.IsStaging {
			acmeOrder, err = service.acmeStaging.GetOrder(*order.Location, key)
		} else {
			acmeOrder, err = service.acmeProd.GetOrder(*order.Location, key)
		}
		if err != nil {
			service.logger.Error(err)
			return
		}

		// action depends on order's current Status
		switch acmeOrder.Status {

		// TODO (to consider): use fallthrough to go from pending -> ready -> valid
		// Alternately, use a loop and after each status should be 'complete', PaG again and check new status

		case "pending": // needs to be authed
			service.logger.Debugf("auth array: %s", acmeOrder.Authorizations)
			err = service.authorizations.FulfillAuthz(acmeOrder.Authorizations, order.Certificate.ChallengeMethod.Type, key, *order.Certificate.AcmeAccount.IsStaging)
			fallthrough

		case "ready": // needs to be finalized
			// TODO: finalize
			fallthrough

		case "valid": // needs to be downloaded
			// TODO: download cert and store it

		case "processing":
			// TODO: wait and try again, ACME server is doing stuff

		case "invalid": // nothing to do, irrecoverable
			service.logger.Error(acmeOrder.Error)
			return

		default:
			service.logger.Error(errors.New("unknown order status"))
			return
		}

		// TODO: Update database with new order status
		// TODO: Update cert (including updated timestamp)

		service.logger.Debugf("end of order fulfiller")

	}()
}
