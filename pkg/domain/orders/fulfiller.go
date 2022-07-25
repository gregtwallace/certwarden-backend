package orders

import (
	"errors"
)

// FulfillOrder launches a go routine to fulfill the specified order
func (service *Service) FulfillOrder(order Order) {
	go func(order Order) {
		// get account key
		key, err := order.Certificate.AcmeAccount.AccountKey()
		if err != nil {
			service.logger.Error(err)
			return
		}

		// PaG the order (to get most recent status)
		if *order.Certificate.AcmeAccount.IsStaging {
			*order.Acme, err = service.acmeStaging.GetOrder(order.Acme.Location, key)
		} else {
			*order.Acme, err = service.acmeProd.GetOrder(order.Acme.Location, key)
		}
		if err != nil {
			service.logger.Error(err)
			return
		}

		// action depends on order's current Status
		switch order.Acme.Status {

		// TODO (to consider): use fallthrough to go from pending -> ready -> valid
		// Alternately, use a loop and after each status should be 'complete', PaG again and check new status

		case "pending": // needs to be authed
			service.logger.Debugf("auth array: %s", order.Acme.Authorizations)
			err = service.authorizations.FulfillAuthz(order.Acme.Authorizations, order.Certificate.ChallengeMethod.Type, key, *order.Certificate.AcmeAccount.IsStaging)
			fallthrough

		case "ready": // needs to be finalized
			// TODO: finalize
			fallthrough

		case "valid": // needs to be downloaded
			// TODO: download cert and store it

		case "processing":
			// TODO: wait and try again, ACME server is doing stuff

		case "invalid": // nothing to do, irrecoverable
			service.logger.Error(order.Acme.Error)
			return

		default:
			service.logger.Error(errors.New("unknown order status"))
			return
		}

		// TODO: Update database with new order status

	}(order)
}
