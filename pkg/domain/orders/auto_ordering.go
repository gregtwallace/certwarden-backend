package orders

import (
	"certwarden-backend/pkg/datatypes/job_manager"
	"certwarden-backend/pkg/randomness"
	"context"
	"errors"
	"sync"
	"time"
)

var autoOrderRunInterval = 2 * time.Hour

// startAutoOrderService starts a go routine that manages certificate renewals. It both completes existing orders
// that are not yet in a 'valid' or 'invalid' state and also places new orders for expiring certs. A job is run
// hourly (with some jitter) to check for work to do. This service also polls and saves ARI information for orders
// on ACME Servers that support ARI.
func (service *Service) startAutoOrderService(ctx context.Context, wg *sync.WaitGroup) {
	// log start and update wg
	service.logger.Infof("orders: starting automatic certificate ordering service; short-lived certificate definition: %d days of validity or less; "+
		"valid remaining threshold: %.01f%%; short-lived validity threshold: %.01f%%; ACME Servers supporting ARI can override these values",
		shortLivedValidityThreshold/(24*time.Hour), expiringRemainingValidFraction*100, expiringShortLivedRemainingValidFraction*100)

	// service routine
	wg.Add(1)
	go func() {
		defer wg.Done()

		// do initial run after app loads and settles
		nextRunTime := time.Now().Add(1 * time.Minute)

		// indefinite service loop
		for {
			select {
			case <-ctx.Done():
				// close routine
				service.logger.Info("orders: automatic certificate ordering service shutdown complete")
				return

			case <-time.After(time.Until(nextRunTime)):
				// proceed to next run
			}

			// debug log activity (todo: comment out?)
			service.logger.Debugf("orders: running auto order tasks")

			// complete existing orders that are not 'valid' or 'invalid' (i.e. not completed)
			service.retryIncompleteOrders()

			// order expiring certificates
			service.orderExpiringCerts()

			// next run time (add autoOrderRunInterval and some jitter)
			// add random second to runtime, as preferred by Let's Encrypt
			// see: https://letsencrypt.org/docs/integration-guide/#when-to-renew
			nextRunTime = time.Now().Add(autoOrderRunInterval)
			nextRunTime = nextRunTime.Add(time.Duration(randomness.GenerateInsecureInt(60)) * time.Second)
		}
	}()
}

// retryIncompleteOrders retries all incomplete orders within storage. this should
// move all orders to valid or invalid state.
func (service *Service) retryIncompleteOrders() {
	// get all incomplete order ids from storage
	incompleteOrderIds, err := service.storage.GetAllIncompleteOrderIds()
	if err != nil {
		service.logger.Errorf("orders: error attempting retry of incomplete orders (%s)", err)
		return
	}

	// nothing to do?
	if len(incompleteOrderIds) <= 0 {
		return
	}

	// add all incompletes to the low priority order queue
	addedCount := 0
	for _, orderId := range incompleteOrderIds {
		err = service.fulfillOrder(orderId, false)
		if err != nil {
			if errors.Is(err, job_manager.ErrAddDuplicateJob) {
				service.logger.Debugf("orders: failed to add order %d to processing queue (order is already in queue)", orderId)
			} else {
				// log error, but keep going through remaining range
				service.logger.Errorf("orders: failed to add order %d to processing queue (%s)", orderId, err)
			}
		} else {
			addedCount++
		}
	}
	service.logger.Infof("orders: %d incomplete orders added to order queue", addedCount)
}
