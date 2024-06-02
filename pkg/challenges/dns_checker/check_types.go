package dns_checker

import (
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
	checkAllServicesFunc := func() error {
		// check for propagation
		return service.checkDnsRecordPropagationAllServices(fqdn, recordValue, txtRecord)
	}

	// backoff properties
	bo := backoff.NewExponentialBackOff()
	bo.InitialInterval = 15 * time.Second
	bo.RandomizationFactor = 0.2
	bo.Multiplier = 1.2
	bo.MaxInterval = 2 * time.Minute
	bo.MaxElapsedTime = 30 * time.Minute

	boWithContext := backoff.WithContext(bo, service.shutdownContext)

	// log failures / delays
	notifyFunc := func(err error, dur time.Duration) {
		service.logger.Infof("dns_checker: %s, will check again in %s", err, dur.Round(100*time.Millisecond))
	}

	// (re)try with backoff
	err := backoff.RetryNotify(checkAllServicesFunc, boWithContext, notifyFunc)

	return err == nil
}
