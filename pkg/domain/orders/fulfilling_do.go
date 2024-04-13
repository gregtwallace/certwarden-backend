package orders

import (
	"certwarden-backend/pkg/acme"
	"certwarden-backend/pkg/randomness"
	"errors"
	"net/http"
	"time"
)

// Do executes the order fulfill job
func (j *orderFulfillJob) Do(workerID int) {
	// log end of Do (regardless of outcome)
	defer j.service.logger.Infof("order fulfilling worker %d: order %d done", workerID, j.orderID)

	// get the relevant order from db
	order, err := j.service.storage.GetOneOrder(j.orderID)
	if err != nil {
		j.service.logger.Errorf("order fulfilling worker %d: error: %w", workerID, err)
		return // done, failed
	}

	// always info log ordering
	j.service.logger.Infof("order fulfilling worker %d: ordering order id %d (certificate name: %s, subject: %s)", workerID, order.ID, order.Certificate.Name, order.Certificate.Subject)

	// update certificate timestamp after fulfiller is done
	defer func() {
		err = j.service.storage.UpdateCertUpdatedTime(order.Certificate.ID)
		if err != nil {
			j.service.logger.Errorf("order fulfilling worker %d: update cert time error: %w", workerID, err)
		}
	}()

	// get account key
	key, err := order.Certificate.CertificateAccount.AcmeAccountKey()
	if err != nil {
		j.service.logger.Errorf("order fulfilling worker %d: get account key error: %w", workerID, err)
		return // done, failed
	}

	// make cert CSR
	csr, err := order.Certificate.MakeCsrDer()
	if err != nil {
		j.service.logger.Errorf("order fulfilling worker %d: make csr error: %w", workerID, err)
		return // done, failed
	}

	// acmeOrder to hold the Order responses and to later update storage
	var acmeOrder acme.Order

	// acmeService to avoid repeated logic
	acmeService, err := j.service.acmeServerService.AcmeService(order.Certificate.CertificateAccount.AcmeServer.ID)
	if err != nil {
		j.service.logger.Errorf("order fulfilling worker %d: select acme service error: %w", workerID, err)
		return // done, failed
	}

	// exponential backoff for retrying while 'processing'
	bo := randomness.BackoffACME(j.service.shutdownContext)

	// Use loop to retry order. Cap loop at 2 hours to avoid indefinite loop if something unexpected
	// occurs (e.g., somethign broken with the acme server).
	startTime := time.Now()
	timeoutLength := 2 * time.Hour

fulfillLoop:
	for time.Since(startTime) <= timeoutLength {

		// Get the order (for most recent Order object and Status)
		acmeOrder, err = acmeService.GetOrder(order.Location, key)
		if err != nil {
			// if ACME returned 404, the order object is now invalid
			// assume the ACME server deleted it and update accordingly
			// see: RFC8555 7.4 ("If the client fails to complete the required
			// actions before the "expires" time, then the server SHOULD change the
			// status of the order to "invalid" and MAY delete the order resource.")
			acmeErr := new(acme.Error)
			if errors.As(err, &acmeErr) && acmeErr.Status == http.StatusNotFound {
				j.service.storage.PutOrderInvalid(order.ID)
				return // done, permanent status
			}

			j.service.logger.Errorf("order fulfilling worker %d: get order error: %w", workerID, err)
			return // done, failed
		}

		// if order is NOT processing, reset the backoff used when in processing; this ensures
		// that any given processing phase starts with a fresh backoff as opposed to including
		// time that elapsed during other statuses that were being worked on
		if acmeOrder.Status != "processing" {
			bo.Reset()
		}

		// action depends on order's current Status
		switch acmeOrder.Status {

		case "pending": // needs to be authed
			var authStatus string
			authStatus, err = j.service.authorizations.FulfillAuths(acmeOrder.Authorizations, key, acmeService)
			if err != nil {
				j.service.logger.Errorf("order fulfilling worker %d: fulfill auths error: %w", workerID, err)
				return // done, failed
			}

			// auth(s) should be valid (thus making order ready)
			// if not valid, should be invalid, loop to get updated order (also now invalid)
			if authStatus != "valid" {
				continue
			}

			// auths were valid, fallthrough to "ready" (which order should now be in)
			fallthrough

		case "ready": // needs to be finalized
			// save finalized_key_id in storage
			err = j.service.storage.UpdateFinalizedKey(order.ID, order.Certificate.CertificateKey.ID)
			if err != nil {
				j.service.logger.Errorf("order fulfilling worker %d: update finalized key error: %w", workerID, err)
				return // done, failed
			}

			// finalize the order
			_, err = acmeService.FinalizeOrder(acmeOrder.Finalize, csr, key)
			if err != nil {
				j.service.logger.Errorf("order fulfilling worker %d: finalize order error: %w", workerID, err)
				return // done, failed
			}

			// should be valid on next check (or maybe processing - sleep a little to try and
			// avoid 'processing')
			time.Sleep(7 * time.Second)
			continue

		case "valid": // can be downloaded
			// download cert pem

			// nil check (make sure there is a cert URL)
			if acmeOrder.Certificate == nil {
				// if cert url is missing (nil), loop again (which will refresh order info)
				continue
			}

			certPemChain, err := acmeService.DownloadCertificate(*acmeOrder.Certificate, key)
			if err != nil {
				j.service.logger.Errorf("order fulfilling worker %d: download cert error: %w", workerID, err)
				return // done, failed
			}

			// process pem and save to storage
			err = j.savePemChain(order.ID, certPemChain)
			if err != nil {
				j.service.logger.Errorf("order fulfilling worker %d: save pem error: %w", workerID, err)
				return // done, failed
			}

			// done
			break fulfillLoop

		case "processing":
			// sleep and loop again, ACME server is working on it
			delayTimer := time.NewTimer(bo.NextBackOff())

			select {
			// cancel on shutdown context
			case <-j.service.shutdownContext.Done():
				// ensure timer releases resources
				if !delayTimer.Stop() {
					<-delayTimer.C
				}

				j.service.logger.Errorf("order fulfilling worker %d: order job canceled due to shutdown", workerID)
				return

			case <-delayTimer.C:
				// retry after exponential backoff
			}

		case "invalid": // break, irrecoverable
			j.service.logger.Infof("order fulfilling worker %d: order status invalid; acme error: %w", workerID, acmeOrder.Error)
			break fulfillLoop

		// Note: there is no 'expired' Status case. If the order expires it simply moves to 'invalid'.

		// should never happen
		default:
			j.service.logger.Errorf("order fulfilling worker %d: error: order status unknown", workerID)
			return // done, failed
		}
	}

	// update order in storage (regardless of loop outcome)
	err = j.service.storage.PutOrderAcme(makeUpdateOrderAcmePayload(order.ID, acmeOrder))
	if err != nil {
		j.service.logger.Errorf("order fulfilling worker %d: update order db error: %w", workerID, err)
	}

	// if order valid, do post processing
	if acmeOrder.Status == "valid" {
		// send to post-processing queue
		if order.hasPostProcessingToDo() {
			err = j.service.postProcess(j.orderID, j.IsHighPriority())
			if err != nil {
				j.service.logger.Errorf("order fulfilling worker %d: failed to post process (%w)")
			}
		}

		// also update Server Cert (if this order was for this app)
		if j.service.serverCertificateName != nil && *j.service.serverCertificateName == order.Certificate.Name {
			err = j.service.loadHttpsCertificateFunc()
			if err != nil {
				j.service.logger.Errorf("order fulfilling worker %d: failed to load app's new https certificate (%s)", workerID, err)
			} else {
				j.service.logger.Debugf("order fulfilling worker %d: new app https certificate loaded", workerID)
			}
		}
	}

	// log error if loop exhausted somehow
	if time.Since(startTime) >= timeoutLength {
		j.service.logger.Errorf("order fulfilling worker %d: order id %d exhausted retry loop time and terminated with status %s (certificate name: %s, subject: %s)", workerID, order.ID, acmeOrder.Status, order.Certificate.Name, order.Certificate.Subject)
	} else {
		// always info log order completed
		j.service.logger.Infof("order fulfilling worker %d: order id %d completed with status %s (certificate name: %s, subject: %s)", workerID, order.ID, acmeOrder.Status, order.Certificate.Name, order.Certificate.Subject)
	}
}
