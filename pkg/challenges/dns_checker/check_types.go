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
	cnameRecord
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

// CheckCNAME checks if the specified fqdn has a cname record pointing at the specified
// value; it works through the consifigured dns servers sequentially; it either returns
// the value immediately or if a dns server isn't working, it continued trying servers
// until one works or they all fail; if a server is working but returns the incorrect
// value, false is returned without trying more dns servers
func (service *Service) CheckCNAME(fqdn string, value string) (propagated bool) {
	// if no resolvers (i.e. configured to skip)
	if service.dnsResolvers == nil {
		service.logger.Debugf("dns_checker: dns servers not configured, skipping check of CNAME %s", fqdn)
		return true
	}

	// go through dns services sequentially until one works
	for i := range service.dnsResolvers {
		exists, e := service.dnsResolvers[i].checkDnsRecord(fqdn, value, cnameRecord)
		if e != nil {
			// should not be getting errors, so write each to log if we get any and go to next server
			service.logger.Errorf("dns_checker: check CNAME %s failed (%s)", fqdn, e)
			continue
		}

		// no error, done (accept either result without trying more)
		return exists
	}

	return false
}
