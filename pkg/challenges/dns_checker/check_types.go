package dns_checker

import (
	"errors"
	"time"
)

// define dnsRecordType
type dnsRecordType int

// consts for supported record types (probably not needed but in the event of
// future expansion)
const (
	UnknownRecordType dnsRecordType = iota

	txtRecord
)

var (
	ErrDnsRecordNotFound = errors.New("dns record not found by checker")
	errShutdown          = errors.New("dns provisioning canceled due to shutdown")
)

// CheckTXTWithRetry checks for the specified record. If the check fails, sleep and retry
// the check again up to the maxTries specified. After exhausing retries, return false
// if still not successful.
func (service *Service) CheckTXTWithRetry(fqdn string, recordValue string, maxTries int) (propagated bool, err error) {
	// retry loop
	for i := 1; i <= maxTries; i++ {
		// check for propagation
		propagated, err := service.checkDnsRecordAllServices(fqdn, recordValue, txtRecord)
		// if error, log error but still retry
		if err != nil {
			service.logger.Error(err)
		} else if propagated {
			// if propagated, done & success
			return true, nil
		}

		// sleep or cancel/error if shutdown is called
		select {
		case <-service.shutdownContext.Done():
			// cancel/error if shutting down
			return false, errShutdown

		case <-time.After(time.Duration(i) * 15 * time.Second):
			// sleep and retry
		}
	}

	// loop exhausted without success
	service.logger.Error(ErrDnsRecordNotFound)
	return false, nil
}
