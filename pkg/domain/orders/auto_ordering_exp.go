package orders

import (
	"certwarden-backend/pkg/pagination_sort"
	"certwarden-backend/pkg/randomness"
	"sync"
	"time"
)

// orderExpiringCerts updates ARI information (where available) and then places orders for expiring certificates
// based either on the ARI information or the thresholds hardcoded above; pending tasks are aborted if the context
// is canceled
func (service *Service) orderExpiringCerts() {
	// get slice of all currently valid orders (to evaluate re-order criteria)
	orders, _, err := service.storage.GetAllValidCurrentOrders(pagination_sort.Query{})
	if err != nil {
		service.logger.Errorf("orders: error attempting order of expiring certs (%s)", err)
		return
	}

	// aysnc checking and updating the authz for validity
	var wg sync.WaitGroup
	wg.Add(len(orders))

	addedCount := 0
	addedMu := sync.Mutex{}

	for i := range orders {
		go func() {
			defer wg.Done()

			// Get relevant ACME Server service
			acmeService, err := service.acmeServerService.AcmeService(orders[i].Certificate.CertificateAccount.AcmeServer.ID)
			if err != nil {
				service.logger.Errorf("orders: auto order failed to get acme service for order %d (%s)", orders[i].ID, err)
				return // done, failed
			}

			// Step 1: Update RenewalInfo (including initial population)
			// In the event ACME Server does not support ARI, CW will generate its own ARI suggested renewal window
			// and use that when deciding when to do renewal
			ari := orders[i].RenewalInfo
			newARI := false

			if acmeService.SupportsARIExtension() &&
				(orders[i].RenewalInfo == nil || orders[i].RenewalInfo.RetryAfter == nil || time.Now().After(*orders[i].RenewalInfo.RetryAfter)) {
				acmeARI, err := acmeService.GetACMERenewalInfo(*orders[i].Pem)
				// TODO: Add retry / exponential backoff for a couple attempts ?
				if err != nil {
					service.logger.Errorf("orders: auto order failed to fetch new ari info for %d (%s)", orders[i].ID, err)
				} else {
					// success
					ari = &renewalInfo{
						SuggestedWindow: struct {
							Start time.Time "json:\"start\""
							End   time.Time "json:\"end\""
						}{
							Start: acmeARI.SuggestedWindow.Start,
							End:   acmeARI.SuggestedWindow.End,
						},
						ExplanationURL: acmeARI.ExplanationURL,
						RetryAfter:     &acmeARI.RetryAfter,
					}
					newARI = true
				}
			}

			// if didn't fetch (server doesn't support, or fetch failed, whatever), and there is no existing ari, use a sane default
			if orders[i].RenewalInfo == nil && !newARI {
				ari = MakeRenewalInfo(*orders[i].ValidFrom, *orders[i].ValidTo)
				newARI = true
			}

			// update storage if we have new ari
			if newARI {
				payload := UpdateRenewalInfoPayload{
					OrderID:     orders[i].ID,
					RenewalInfo: ari,
					UpdatedAt:   int(time.Now().Unix()),
				}
				err = service.storage.PutRenewalInfo(payload)
				if err != nil {
					service.logger.Errorf("orders: auto order failed to save renewal info for order %d to storage (%s)", orders[i].ID, err)
				}
			}

			// Step 2: Check if renewal is due, if so, place order
			// select a time within ari by calculating the duration in minutes and adding that number of minutes to the window start
			// Note: Renewal time is not stored, a random time in the window is selected every time
			windowDuration := ari.SuggestedWindow.End.Sub(ari.SuggestedWindow.Start)
			renewalTime := ari.SuggestedWindow.Start.Add(time.Duration(randomness.GenerateInsecureInt(int(windowDuration.Minutes()))) * time.Minute)

			// If the selected renewalTime is before the approximate next wakeup (now + autoOrderRunInterval),
			// then renew now, otherwise do nothing and see what happens next run
			if renewalTime.Before(time.Now().Add(autoOrderRunInterval)) {
				service.logger.Debugf("orders: auto order placing new order for expiring cert %s (window from: %s; to: %s; selected renewal time: %s)",
					orders[i].Certificate.Name, ari.SuggestedWindow.Start, ari.SuggestedWindow.End, renewalTime)
				_, outErr := service.placeNewOrderAndFulfill(orders[i].Certificate.ID, false)
				if outErr != nil {
					service.logger.Errorf("orders: auto order failed to place new order for cert %s (%s)", orders[i].Certificate.Name, err)
				} else {
					addedMu.Lock()
					addedCount++
					addedMu.Unlock()
				}
			}
		}()
	}

	// wait for task completion
	wg.Wait()

	service.logger.Infof("orders: auto order added %d orders to queue", addedCount)
}
