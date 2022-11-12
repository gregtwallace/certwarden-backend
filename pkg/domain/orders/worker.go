package orders

import (
	"errors"
	"legocerthub-backend/pkg/acme"
	"time"
)

// orderJob contains the info the worker needs to do a job
type orderJob struct {
	orderId int
}

// makeOrderWorker creates a indefinite thread to process incoming orderJobs
func (service *Service) makeOrderWorker(id int, highPriorityJobs <-chan orderJob, lowPriorityJobs <-chan orderJob) {
	for {
		select {
		case highJob := <-highPriorityJobs:
			service.doOrderJob(highJob)
			service.logger.Debugf("worker %d: end of high priority order fulfiller (orderId: %d)", id, highJob.orderId)
		case lowJob := <-lowPriorityJobs:
		lower:
			for {
				select {
				case highJob := <-highPriorityJobs:
					service.doOrderJob(highJob)
					service.logger.Debugf("worker %d: end of high priority order fulfiller (orderId: %d)", id, highJob.orderId)
				default:
					break lower
				}
			}
			service.doOrderJob(lowJob)
			service.logger.Debugf("worker %d: end of low priority order fulfiller (orderId: %d)", id, lowJob.orderId)
		}

	}
}

// doOrderJob works the orderJob specified. Once complete, the job is removed from the inProcess
// list to indicate completion. No results are returned as results are saved directly to storage
// as part of doing the job.
func (service *Service) doOrderJob(job orderJob) {
	// remove id from inProcess when goroutine is done
	defer func(orderId int, service *Service) {
		err := service.inProcess.remove(orderId)
		if err != nil {
			service.logger.Error(err)
		}
	}(job.orderId, service)

	// fetch the relevant order
	order, err := service.storage.GetOneOrder(job.orderId)
	if err != nil {
		service.logger.Error(err)
		return // done, failed
	}

	// update certificate timestamp after fulfiller is done
	defer func(certId int) {
		err = service.storage.UpdateCertUpdatedTime(certId)
		if err != nil {
			service.logger.Error(err)
		}
	}(order.Certificate.ID)

	// get account key
	key, err := order.Certificate.CertificateAccount.AcmeAccountKey()
	if err != nil {
		service.logger.Error(err)
		return // done, failed
	}

	// make cert CSR
	csr, err := order.Certificate.MakeCsrDer()
	if err != nil {
		service.logger.Error(err)
		return // done, failed
	}

	// acmeOrder to hold the Order responses and to later update storage
	var acmeOrder acme.Order

	// acmeService to avoid repeated isStaging logic
	var acmeService *acme.Service
	if order.Certificate.CertificateAccount.IsStaging {
		acmeService = service.acmeStaging
	} else {
		acmeService = service.acmeProd
	}

	// Use loop to retry order. Cap retries to avoid indefinite loop.
	maxTries := 5
fulfillLoop:
	for i := 1; i <= maxTries; i++ {
		// Get the order (for most recent Order object and Status)
		acmeOrder, err = acmeService.GetOrder(order.Location, key)
		if err != nil {
			service.logger.Error(err)
			return // done, failed
		}

		// action depends on order's current Status
		switch acmeOrder.Status {
		case "pending": // needs to be authed
			var authStatus string
			authStatus, err = service.authorizations.FulfillAuths(acmeOrder.Authorizations, order.Certificate.ChallengeMethod, key, order.Certificate.CertificateAccount.IsStaging)
			if err != nil {
				service.logger.Error(err)
				return // done, failed
			}
			// auth should be valid (thus making order ready)
			// if not valid, should be invalid, loop to get updated order (also now invalid)
			if authStatus != "valid" {
				break
			}

			// auths were valid, fallthrough to "ready" (which order should now be in)
			fallthrough

		case "ready": // needs to be finalized
			// save finalized_key_id in storage
			err = service.storage.UpdateFinalizedKey(order.ID, order.Certificate.CertificateKey.ID)
			if err != nil {
				service.logger.Error(err)
				return // done, failed
			}

			// finalize the order
			acmeOrder, err = acmeService.FinalizeOrder(order.Finalize, csr, key)
			if err != nil {
				service.logger.Error(err)
				return // done, failed
			}

			// should now be valid, if not, probably processing
			if acmeOrder.Status != "valid" {
				break
			}

			// if order is valid, fallthrough to valid
			fallthrough

		case "valid": // can be downloaded
			// download cert pem
			// nil check (make sure there is a cert URL)
			if acmeOrder.Certificate != nil {
				certPemChain, err := acmeService.DownloadCertificate(*acmeOrder.Certificate, key)
				if err != nil {
					service.logger.Error(err)
					return // done, failed
				}

				// process pem and save to storage
				err = service.savePemChain(order.ID, certPemChain)
				if err != nil {
					service.logger.Error(err)
					return
				}

				break fulfillLoop
			}

			// if cert url is missing (nil), loop again (which will refresh order info)

		case "processing":
			// TODO: Implement exponential backoff
			if i != maxTries {
				time.Sleep(time.Duration(i) * 30 * time.Second)
			}

		case "invalid": // break, irrecoverable
			service.logger.Debugf("order status invalid; acme error: %s", acmeOrder.Error)
			break fulfillLoop

		// Note: there is no 'expired' Status case. If the order expires it simply moves to 'invalid'.

		default:
			service.logger.Error(errors.New("order status unknown"))
			return // done, failed
		}
	}

	// update order in storage
	err = service.storage.PutOrderAcme(makeUpdateOrderAcmePayload(job.orderId, acmeOrder))
	if err != nil {
		service.logger.Error(err)
	}
}
