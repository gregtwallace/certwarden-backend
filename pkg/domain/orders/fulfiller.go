package orders

import (
	"errors"
	"legocerthub-backend/pkg/acme"
	"time"
)

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

		// update certificate timestamp after fulfiller is done
		defer func(certId int) {
			err = service.storage.UpdateCertUpdatedTime(certId)
			if err != nil {
				service.logger.Error(err)
			}
		}(*order.Certificate.ID)

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
		if *order.Certificate.AcmeAccount.IsStaging {
			acmeService = service.acmeStaging
		} else {
			acmeService = service.acmeProd
		}

		// Use loop to retry order. Cap retries to avoid indefinite loop.
		maxTries := 5
	fulfillLoop:
		for i := 1; i <= maxTries; i++ {
			// Get the order (for most recent Order object and Status)
			acmeOrder, err = acmeService.GetOrder(*order.Location, key)
			if err != nil {
				service.logger.Error(err)
				return // done, failed
			}

			// action depends on order's current Status
			switch acmeOrder.Status {
			case "pending": // needs to be authed
				var authStatus string
				authStatus, err = service.authorizations.FulfillAuths(acmeOrder.Authorizations, *order.Certificate.ChallengeMethod, key, *order.Certificate.AcmeAccount.IsStaging)
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
				err = service.storage.UpdateFinalizedKey(*order.ID, *order.Certificate.PrivateKey.ID)
				if err != nil {
					service.logger.Error(err)
					return // done, failed
				}

				// finalize the order
				acmeOrder, err = acmeService.FinalizeOrder(*order.Finalize, csr, key)
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
				certPemChain, err := acmeService.DownloadCertificate(acmeOrder.Certificate, key)
				if err != nil {
					service.logger.Error(err)
					return // done, failed
				}

				// process pem and save to storage
				err = service.savePemChain(*order.ID, certPemChain)
				if err != nil {
					service.logger.Error(err)
					return
				}

				break fulfillLoop

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
		err = service.storage.UpdateOrderAcme(orderId, acmeOrder)
		if err != nil {
			service.logger.Error(err)
		}

	}(orderId, service)

	return nil
}
