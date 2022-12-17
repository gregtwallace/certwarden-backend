package orders

import (
	"database/sql"
	"legocerthub-backend/pkg/pagination_sort"
	"time"
)

// background reorder config options
// TODO: Move to customizable setting in config.yaml and frontent->settings
// reorderTime: If less than this duration of time remaining on cert,
// LeGo Certhub will try to obtain a newer cert.
var reorderTime = 40 * (24 * time.Hour)

// daily time to refresh
const refreshHour = 03
const refreshMinute = 12

// backgroundCertRefresher orders expiring certs on a daily basis at a time
// specified
func (service *Service) backgroundCertRefresher() {
	service.logger.Info("starting background cert refresh go routine")

	// go routine for refreshing
	go func(service *Service) {
		var nextRunTime time.Time

		// indefinite refresh loop
		for {
			// today's runtime
			todayRunTime := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(),
				refreshHour, refreshMinute, 0, 0, time.Local)

			// calculate next run based on if today's runtime has passed or not
			if todayRunTime.After(time.Now()) {
				// if today's run hasn't passed, next run is today
				nextRunTime = todayRunTime
			} else {
				// if today's time HAS passed, next run is tomorrow
				nextRunTime = todayRunTime.Add(24 * time.Hour)
			}

			// sleep until next run
			time.Sleep(time.Until(nextRunTime))

			// run refresh
			err := service.orderExpiringCerts()
			if err != nil {
				service.logger.Errorf("error ordering expiring certs: %s", err)
			}
		}
	}(service)
}

// orderExpiringCerts automatically orders any certficates that are within
// the specified expiration window.
func (service *Service) orderExpiringCerts() (err error) {
	service.logger.Info("ordering any expiring certificates")

	// get orders relating to all currently valid certs, filtering with the
	// reorderTime so only expiring orders are returned
	expiringOrders, _, err := service.storage.GetAllValidCurrentOrders(pagination_sort.QueryAll, &reorderTime)
	if err != nil {
		return err
	}

	// for each order, check expiration
	for i := range expiringOrders {
		// refresh
		err = service.refreshCert(expiringOrders[i].Certificate.ID)
		if err != nil {
			// log error, but keep going through remaining range
			service.logger.Errorf("failed to refresh cert (%d): %s", expiringOrders[i].Certificate.ID, err)
		}
		// sleep a little so slew of new orders doesn't hit ACME all at once
		time.Sleep(15 * time.Second)
	}

	service.logger.Info("placement of expiring certificate orders complete")
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

	// place order and kickoff low-priority fulfillment
	_, err = service.placeNewOrderAndFulfill(certId, false)
	if err != nil {
		return err
	}

	return nil
}
