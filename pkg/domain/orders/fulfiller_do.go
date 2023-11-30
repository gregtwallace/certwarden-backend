package orders

import (
	"legocerthub-backend/pkg/acme"
	"legocerthub-backend/pkg/randomness"
	"net/http"
	"time"
)

// doJob works the specified job. It moves the job from the manager waiting queue to the worker
// that is working the job. Once complete, the job is removed from the manager queue. No results
// are returned as results are saved directly to storage as part of doing the job.
func (of *orderFulfiller) doJob(j orderFulfillerJob, workerId int) {
	// move job to worker in tracker
	err := of.moveJobToWorking(j, workerId)
	if err != nil {
		of.logger.Errorf("worker %d: error: %s", workerId, err)
	}

	// defer remove job from tracker
	defer func() {
		err := of.removeJob(j, workerId)
		if err != nil {
			of.logger.Errorf("worker %d: error: %s", workerId, err)
		}
	}()

	// refresh the relevant order from db
	j.order, err = of.storage.GetOneOrder(j.order.ID)
	if err != nil {
		of.logger.Errorf("worker %d: error: %s", workerId, err)
		return // done, failed
	}

	// always info log ordering
	of.logger.Infof("worker %d: ordering order id %d (certificate name: %s, subject: %s)", workerId, j.order.ID, j.order.Certificate.Name, j.order.Certificate.Subject)

	// update certificate timestamp after fulfiller is done
	defer func() {
		err = of.storage.UpdateCertUpdatedTime(j.order.Certificate.ID)
		if err != nil {
			of.logger.Errorf("worker %d: error: %s", workerId, err)
		}
	}()

	// get account key
	key, err := j.order.Certificate.CertificateAccount.AcmeAccountKey()
	if err != nil {
		of.logger.Errorf("worker %d: error: %s", workerId, err)
		return // done, failed
	}

	// make cert CSR
	csr, err := j.order.Certificate.MakeCsrDer()
	if err != nil {
		of.logger.Errorf("worker %d: error: %s", workerId, err)
		return // done, failed
	}

	// acmeOrder to hold the Order responses and to later update storage
	var acmeOrder acme.Order

	// acmeService to avoid repeated logic
	acmeService, err := of.acmeServerService.AcmeService(j.order.Certificate.CertificateAccount.AcmeServer.ID)
	if err != nil {
		of.logger.Errorf("worker %d: error: %s", workerId, err)
		return // done, failed
	}

	// exponential backoff for retrying while 'processing'
	bo := randomness.BackoffACME(of.shutdownContext)

	// Use loop to retry order. Cap loop at 2 hours to avoid indefinite loop if something unexpected
	// occurs (e.g., somethign broken with the acme server).
	startTime := time.Now()
	timeoutLength := 2 * time.Hour

fulfillLoop:
	for time.Since(startTime) <= timeoutLength {

		// Get the order (for most recent Order object and Status)
		acmeOrder, err = acmeService.GetOrder(j.order.Location, key)
		if err != nil {
			// if ACME returned 404, the order object is now invalid
			// assume the ACME server deleted it and update accordingly
			// see: RFC8555 7.4 ("If the client fails to complete the required
			// actions before the "expires" time, then the server SHOULD change the
			// status of the order to "invalid" and MAY delete the order resource.")
			if acmeErr, ok := err.(acme.Error); ok && acmeErr.Status == http.StatusNotFound {
				of.storage.PutOrderInvalid(j.order.ID)
				return // done, permanent status
			}

			of.logger.Errorf("worker %d: error: %s", workerId, err)
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
			authStatus, err = of.authorizations.FulfillAuths(acmeOrder.Authorizations, key, acmeService)
			if err != nil {
				of.logger.Errorf("worker %d: error: %s", workerId, err)
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
			err = of.storage.UpdateFinalizedKey(j.order.ID, j.order.Certificate.CertificateKey.ID)
			if err != nil {
				of.logger.Errorf("worker %d: error: %s", workerId, err)
				return // done, failed
			}

			// finalize the order
			_, err = acmeService.FinalizeOrder(acmeOrder.Finalize, csr, key)
			if err != nil {
				of.logger.Errorf("worker %d: error: %s", workerId, err)
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
				of.logger.Errorf("worker %d: error: %s", workerId, err)
				return // done, failed
			}

			// process pem and save to storage
			err = of.savePemChain(j.order.ID, certPemChain)
			if err != nil {
				of.logger.Errorf("worker %d: error: %s", workerId, err)
				return // done, failed
			}

			// done
			break fulfillLoop

		case "processing":
			// sleep and loop again, ACME server is working on it

			select {
			// cancel on shutdown context
			case <-of.shutdownContext.Done():
				of.logger.Errorf("worker %d: order job canceled due to shutdown", workerId)
				return

			case <-time.After(bo.NextBackOff()):
				// sleep and retry using exponential backoff
			}

		case "invalid": // break, irrecoverable
			of.logger.Infof("worker %d: order status invalid; acme error: %s", workerId, acmeOrder.Error)
			break fulfillLoop

		// Note: there is no 'expired' Status case. If the order expires it simply moves to 'invalid'.

		// should never happen
		default:
			of.logger.Errorf("worker %d: error: order status unknown", workerId)
			return // done, failed
		}
	}

	// update order in storage (regardless of loop outcome)
	err = of.storage.PutOrderAcme(makeUpdateOrderAcmePayload(j.order.ID, acmeOrder))
	if err != nil {
		of.logger.Errorf("worker %d: error: %s", workerId, err)
	}

	// if order success, check if this is app's certificate and if so inform app
	// to reload the new order
	if of.isHttps && of.serverCertificateName != nil &&
		*of.serverCertificateName == j.order.Certificate.Name &&
		acmeOrder.Status == "valid" {
		err = of.loadHttpsCertificate()
		if err != nil {
			of.logger.Errorf("worker %d: failed to load lego's new https certificate (%s)", workerId, err)
		} else {
			of.logger.Debugf("worker %d: new lego https certificate loaded", workerId)
		}
	}

	// log error if loop exhausted somehow
	if time.Since(startTime) >= timeoutLength {
		of.logger.Errorf("worker %d: order id %d exhausted retry loop time and terminated with status %s (certificate name: %s, subject: %s)", workerId, j.order.ID, acmeOrder.Status, j.order.Certificate.Name, j.order.Certificate.Subject)
	} else {
		// always info log order completed
		of.logger.Infof("worker %d: order id %d completed with status %s (certificate name: %s, subject: %s)", workerId, j.order.ID, acmeOrder.Status, j.order.Certificate.Name, j.order.Certificate.Subject)
	}
}
