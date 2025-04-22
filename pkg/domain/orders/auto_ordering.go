package orders

import (
	"certwarden-backend/pkg/pagination_sort"
	"certwarden-backend/pkg/randomness"
	"context"
	"database/sql"
	"sync"
	"time"
)

// expiringAfterElapsedRatio is the ratio of a certificate's elapsed validity / total validity
// after which the certificate should be considered expiring
// KEEP IN SYNC with frontend `FlagExpireDays.tsx` consts
const (
	// expiringRemainingValidFraction is the fraction of valid time remaining on a certificate below
	// which the certificate is considered expiring
	expiringRemainingValidFraction = 0.333
	// expiringMinDaysRemaining is the number of days of valid time remaining (regardless of percentage
	// of valid time) after which a certificate is considered expiring
	expiringMinRemaining = time.Duration(10 * 24 * time.Hour)
)

// startAutoOrderService starts a go routine that completes existing orders that are
// not yet in a 'valid' or 'invalid' state and also places new orders forexpiring certs
// The service runs daily at the time specified in consts.
func (service *Service) startAutoOrderService(cfg *Config, ctx context.Context, wg *sync.WaitGroup) {
	// dont run if not enabled
	if !*cfg.AutomaticOrderingEnable {
		return
	}

	// calculate timing based on config
	refreshHour := *cfg.RefreshTimeHour
	refreshMinute := *cfg.RefreshTimeMinute

	// log start and update wg
	service.logger.Infof("orders: starting automatic certificate ordering service; percent valid remaining threshold: %.0f%%; expiration threshold: %.01f days; "+
		"orders will be placed every day at %02d:%02d", expiringRemainingValidFraction*100, (expiringMinRemaining.Hours() / 24), refreshHour, refreshMinute)
	wg.Add(1)

	// service routine
	go func() {
		defer wg.Done()
		var nextRunTime time.Time

		// indefinite service loop
		for {
			// run time for today
			nextRunTime = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(),
				refreshHour, refreshMinute, 0, 0, time.Local)

			// if today's run already passed, run tomorrow
			if !nextRunTime.After(time.Now()) {
				nextRunTime = nextRunTime.Add(24 * time.Hour)
			}

			// add random second to runtime, as preferred by Let's Encrypt
			// see: https://letsencrypt.org/docs/integration-guide/#when-to-renew
			// added after timestamp calc to avoid accidental duplicate run on same day
			// e.g. if runs at :12 and then next timestamp is :50, it is possible for the
			// new stamp to not be after now and therefore would run a second time
			nextRunTime = nextRunTime.Add(time.Duration(randomness.GenerateInsecureInt(60)) * time.Second)

			select {
			case <-ctx.Done():
				// close routine
				service.logger.Info("orders: automatic certificate ordering service shutdown complete")
				return

			case <-time.After(time.Until(nextRunTime)):
				// proceed to next run
			}

			// complete existing orders that are not 'valid' or 'invalid' (i.e. not completed)
			err := service.retryIncompleteOrders()
			if err != nil {
				service.logger.Errorf("orders: error retying incomplete orders: %s", err)
			}

			// order expiring certificates
			service.orderExpiringCerts()
		}
	}()
}

// retryIncompleteOrders retries all incomplete orders within storage. this should
// move all orders to valid or invalid state.
func (service *Service) retryIncompleteOrders() (err error) {
	service.logger.Info("orders: adding incomplete orders to order queue")

	// get all incomplete order ids from storage
	incompleteOrderIds, err := service.storage.GetAllIncompleteOrderIds()
	if err != nil {
		return err
	}

	// add all incompletes to the low priority order queue
	for _, orderId := range incompleteOrderIds {
		err = service.fulfillOrder(orderId, false)
		if err != nil {
			// log error, but keep going through remaining range
			service.logger.Errorf("orders: failed to add order %d to processing queue (%s)", orderId, err)
		}
	}

	service.logger.Info("orders: incomplete orders added to order queue")
	return nil
}

// orderExpiringCerts automatically orders any certficates that have surpassed their expiration
// threshold (either percentage wise or the hardcoded backstop value)
func (service *Service) orderExpiringCerts() {
	service.logger.Info("orders: adding expiring certificates to order queue")

	// get slice of all currently valid orders (to evaluate re-order criteria)
	allValidOrders, _, err := service.storage.GetAllValidCurrentOrders(pagination_sort.Query{})
	if err != nil {
		service.logger.Errorf("orders: error ordering expiring certs: %s", err)
		return
	}

	// use a common Now time
	now := time.Now()

	// review every currently valid cert
	for _, validOrder := range allValidOrders {
		// nil checks (should not be possible)
		if validOrder.ValidFrom == nil {
			service.logger.Errorf("orders: valid order somehow missing validFrom time")
			continue
		}
		if validOrder.ValidTo == nil {
			service.logger.Errorf("orders: valid order somehow missing validTo time")
			continue
		}

		// calculate the threshhold dates using the ratio and backstop values
		// Option 1: validTo - (validTo - validFrom) * expiringRemainingValidFraction
		totalDuration := validOrder.ValidTo.Sub(*validOrder.ValidFrom)
		remainingValidFractionThresholdDate := validOrder.ValidTo.Add(-1 * time.Duration(float64(totalDuration)*expiringRemainingValidFraction))

		// Option 2: validTo - expiringMinRemaining
		remainingValidMinThresholdDate := validOrder.ValidTo.Add(-1 * expiringMinRemaining)

		// abort if not past either threshold (% remaining and backstop time remaining)
		if now.Before(remainingValidFractionThresholdDate) && now.Before(remainingValidMinThresholdDate) {
			// Now is before the thresholds, so not expiring, skip this one
			continue
		}

		// this order is considered expiring -- proceed

		// check for an existing incomplete order
		orderId, err := service.storage.GetNewestIncompleteCertOrderId(validOrder.Certificate.ID)
		if err != nil {
			// unable to get existing incomplete order -> place new order
			// if error other than NoRows, log it
			if err != sql.ErrNoRows {
				service.logger.Errorf("orders: failed to fetch newest incomplete order id for cert %s (%s); will try to place new order", validOrder.Certificate.Name, err)
			}

			// place new order
			service.logger.Debugf("orders: placing new order for expiring cert %s", validOrder.Certificate.Name)
			_, outErr := service.placeNewOrderAndFulfill(validOrder.Certificate.ID, false)
			if outErr != nil {
				service.logger.Errorf("orders: failed to place new order for cert %s (%s)", validOrder.Certificate.Name, err)
			}

		} else {
			// no error, retry existing order
			service.logger.Debugf("orders: retrying order %d to refresh cert %s", orderId, validOrder.Certificate.Name)
			err = service.fulfillOrder(orderId, false)
			if err != nil {
				service.logger.Errorf("orders: failed to retry order %d for cert %s (%s)", orderId, validOrder.Certificate.Name, err)
			}
		}

		// sleep a little so slew of new orders don't hit ACME all at once
		// cancel on shutdown context
		select {
		case <-service.shutdownContext.Done():
			// abort refreshing due to shutdown
			service.logger.Info("orders: expiring certificates refresh canceled due to shutdown")
			return

		case <-time.After(15 * time.Second):
			// proceed to next
		}
	}

	service.logger.Info("orders: expiring certificates added to order queue")
}
