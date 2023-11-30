package dns_checker

import (
	"errors"
	"time"

	"github.com/cenkalti/backoff/v4"
)

// define dnsRecordType
type dnsRecordType int

// consts for supported record types (probably not needed but in the event of
// future expansion)
const (
	UnknownRecordType dnsRecordType = iota

	txtRecord
)

// CheckTXTWithRetry checks for the specified record. If the check fails, use exponential
// backoff until that times out and then return false if propagation still hasn't occurred.
func (service *Service) CheckTXTWithRetry(fqdn string, recordValue string) (propagated bool) {
	// func to try with exponential backoff
	checkAllServicesFunc := func() (bool, error) {
		// check for propagation
		propagated := service.checkDnsRecordAllServices(fqdn, recordValue, txtRecord)

		// if propagated, done & success
		if propagated {
			return true, nil
		}

		// return err to trigger retry
		return false, errors.New("try again")
	}

	// backoff properties
	bo := backoff.NewExponentialBackOff()
	bo.InitialInterval = 15 * time.Second
	bo.RandomizationFactor = 0.2
	bo.Multiplier = 1.2
	bo.MaxInterval = 2 * time.Minute
	bo.MaxElapsedTime = 30 * time.Minute

	boWithContext := backoff.WithContext(bo, service.shutdownContext)

	// log failures
	notifyFunc := func(_err error, dur time.Duration) {
		service.logger.Infof("dns TXT record %s value %s is still not propagated, will check again in %s", fqdn, recordValue, dur.Round(100*time.Millisecond))
	}

	// (re)try with backoff
	propagated, err := backoff.RetryNotifyWithData(checkAllServicesFunc, boWithContext, notifyFunc)
	if err != nil || !propagated {
		return false
	}

	return true
}
