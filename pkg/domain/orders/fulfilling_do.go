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
	defer j.service.logger.Infof("orders: fulfilling worker %d: order %d done", workerID, j.orderID)

	// get the relevant order from db
	order, err := j.service.storage.GetOneOrder(j.orderID)
	if err != nil {
		j.service.logger.Errorf("orders: fulfilling worker %d: error: %s", workerID, err)
		return // done, failed
	}

	// always info log ordering
	j.service.logger.Infof("orders: fulfilling worker %d: ordering order id %d (certificate name: %s, subject: %s)", workerID, order.ID, order.Certificate.Name, order.Certificate.Subject)

	// update certificate timestamp after fulfiller is done
	defer func() {
		err = j.service.storage.UpdateCertUpdatedTime(order.Certificate.ID)
		if err != nil {
			j.service.logger.Errorf("orders: fulfilling worker %d: update cert time error: %s", workerID, err)
		}
	}()

	// get account key
	key, err := order.Certificate.CertificateAccount.AcmeAccountKey()
	if err != nil {
		j.service.logger.Errorf("orders: fulfilling worker %d: get account key error: %s", workerID, err)
		return // done, failed
	}

	// make cert CSR
	csr, err := order.Certificate.MakeCsrDer()
	if err != nil {
		j.service.logger.Errorf("orders: fulfilling worker %d: make csr error: %s", workerID, err)
		return // done, failed
	}

	// acmeOrder to hold the Order responses and to later update storage
	var acmeOrder acme.Order

	// acmeService to avoid repeated logic
	acmeService, err := j.service.acmeServerService.AcmeService(order.Certificate.CertificateAccount.AcmeServer.ID)
	if err != nil {
		j.service.logger.Errorf("orders: fulfilling worker %d: select acme service error: %s", workerID, err)
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
			// if ACME returned 404, the order object is now invalid;
			// assume the ACME server deleted it and update accordingly
			// see: RFC8555 7.4 ("If the client fails to complete the required
			// actions before the "expires" time, then the server SHOULD change the
			// status of the order to "invalid" and MAY delete the order resource.")
			acmeErr := new(acme.Error)
			if errors.As(err, &acmeErr) && acmeErr.Status == http.StatusNotFound {
				j.service.storage.PutOrderInvalid(order.ID)
				return // done, permanent status
			}

			j.service.logger.Errorf("orders: fulfilling worker %d: get order error: %s", workerID, err)
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
			err = j.service.authorizations.FulfillAuths(acmeOrder.Authorizations, key, acmeService)
			if err != nil {
				j.service.logger.Errorf("orders: fulfilling worker %d: fulfill auths error: %s", workerID, err)
				return // done, failed
			}

		case "ready": // needs to be finalized
			// save finalized_key_id in storage (if finalize ACME cmd below fails, this will save any key change
			// upon next attempt to finalize with ACME; therefore this should always occur BEFORE the ACME finalize
			// command)
			err = j.service.storage.UpdateFinalizedKey(order.ID, order.Certificate.CertificateKey.ID)
			if err != nil {
				j.service.logger.Errorf("orders: fulfilling worker %d: update finalized key error: %s", workerID, err)
				return // done, failed
			}

			// finalize the order
			_, err = acmeService.FinalizeOrder(acmeOrder.Finalize, csr, key)
			if err != nil {
				j.service.logger.Errorf("orders: fulfilling worker %d: finalize order error: %s", workerID, err)
				return // done, failed
			}

			// should be valid on next check (or maybe processing - sleep a little to try and avoid 'processing')
			time.Sleep(7 * time.Second)

		case "valid": // can be downloaded - final status
			// nil check (make sure there is a cert URL)
			if acmeOrder.Certificate == nil {
				// if cert url is missing (nil), sleep a little and loop again (which will refresh order info)
				time.Sleep(7 * time.Second)
				continue
			}

			cert, err := acmeService.DownloadCertificate(*acmeOrder.Certificate, key, order.Certificate.PreferredRootCN)
			if err != nil {
				j.service.logger.Errorf("orders: fulfilling worker %d: download cert error: %s", workerID, err)
				return // done, failed
			}

			// process pem and save to storage
			err = j.saveAcmeCert(order.ID, cert)
			if err != nil {
				j.service.logger.Errorf("orders: fulfilling worker %d: save pem error: %s", workerID, err)
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

				j.service.logger.Errorf("orders: fulfilling worker %d: order job canceled due to shutdown", workerID)
				return

			case <-delayTimer.C:
				// retry after exponential backoff
			}

		case "invalid": // break, irrecoverable - final status
			j.service.logger.Infof("orders: fulfilling worker %d: order status invalid; acme error: %s", workerID, acmeOrder.Error)
			break fulfillLoop

		// Note: there is no 'expired' Status case. If the order expires it simply moves to 'invalid'.

		// should never happen
		default:
			j.service.logger.Errorf("orders: fulfilling worker %d: error: order status unknown", workerID)
			return // done, failed
		}
	}

	// did loop timeout?
	loopTimedOut := time.Since(startTime) >= timeoutLength

	// update order in storage (regardless of loop outcome)
	err = j.service.storage.PutOrderAcme(makeUpdateOrderAcmePayload(order.ID, acmeOrder))
	if err != nil {
		j.service.logger.Errorf("orders: fulfilling worker %d: update order db error: %s", workerID, err)
	}

	// if loop timed out, log error and finish
	if loopTimedOut {
		j.service.logger.Errorf("orders: fulfilling worker %d: order id %d exhausted retry loop time and terminated with status %s (certificate name: %s, subject: %s)", workerID, order.ID, acmeOrder.Status, order.Certificate.Name, order.Certificate.Subject)
		return
	}

	// if order valid, do post processing
	if acmeOrder.Status == "valid" {
		// send to post-processing queue
		if order.hasPostProcessingToDo() {
			err = j.service.postProcess(j.orderID, j.IsHighPriority())
			if err != nil {
				j.service.logger.Errorf("orders: fulfilling worker %d: failed to add post process job (%s)")
			}
		}

		// also update Server Cert (if this order was for this app)
		if j.service.serverCertificateName != nil && *j.service.serverCertificateName == order.Certificate.Name {
			err = j.service.loadHttpsCertificateFunc()
			if err != nil {
				j.service.logger.Errorf("orders: fulfilling worker %d: failed to load app's new https certificate (%s)", workerID, err)
			} else {
				j.service.logger.Debugf("orders: fulfilling worker %d: new app https certificate loaded", workerID)
			}
		}
	}

	// success
	j.service.logger.Infof("orders: fulfilling worker %d: order id %d completed with status %s (certificate name: %s, subject: %s)", workerID, order.ID, acmeOrder.Status, order.Certificate.Name, order.Certificate.Subject)

}
