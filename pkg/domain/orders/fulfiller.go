package orders

import (
	"errors"
	"legocerthub-backend/pkg/acme"
	"sync"
	"time"
)

var ErrOrderAlreadyWorking = errors.New("order is already being processed")

// working holds a list of orders being worked
type working struct {
	ids []int
	mu  sync.Mutex
}

// newWorking creates the working struct for tracking orders being worked on
func newWorking() *working {
	return &working{
		ids: []int{},
	}
}

// add adds the specified order ID to the working slice (which tracks
// orders currently being worked)
func (working *working) add(orderId int) (err error) {
	working.mu.Lock()
	defer working.mu.Unlock()

	// check if already working
	for i := range working.ids {
		if working.ids[i] == orderId {
			return ErrOrderAlreadyWorking
		}
	}

	// if not, add
	working.ids = append(working.ids, orderId)
	return nil
}

// remove removes the specified order ID from the working slice.
// This is done after the fulfillment routine completes
func (working *working) remove(orderId int) (err error) {
	working.mu.Lock()
	defer working.mu.Unlock()

	// find index
	i := -2
	for i = range working.ids {
		if working.ids[i] == orderId {
			break
		}
	}

	// if id was not found, error
	if i == -2 {
		return errors.New("order fulfiller cannot remove non existent id")
	}

	// move last element to location of element being removed
	working.ids[i] = working.ids[len(working.ids)-1]
	working.ids = working.ids[:len(working.ids)-1]

	return nil
}

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
				err = service.authorizations.FulfillAuthz(acmeOrder.Authorizations, *order.Certificate.ChallengeMethod, key, *order.Certificate.AcmeAccount.IsStaging)
				if err != nil {
					service.logger.Debug(err)
					return // done, failed
				}

				fallthrough

			case "ready": // needs to be finalized
				// TODO: finalize

				fallthrough

			case "valid": // can be downloaded
				// TODO: download? update related cert?

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
