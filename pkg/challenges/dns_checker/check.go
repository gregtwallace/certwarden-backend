package dns_checker

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

// thresholds to decide if checking succeeded for not.
// propagationRequirement is the portion of functioning dns services that need
// to return the expected record for the check to yield TRUE (e.g. 1 = 100%)
// functioningRequirement is the portion of DNS services that must not fail to
// resolve in order for a check to not produce an Error.
const (
	functioningRequirement = 0.5
	propagationRequirement = 1.0
)

// checkDnsRecord checks if the fqdn has a record of the specified type, set to the specified
// value, on the specified dns resolver. If the record does not exist or exists but the value is
// different, false is returned. If there is an error querying for the record, an error is returned.
func checkDnsRecord(fqdn string, recordValue string, recordType dnsRecordType, r *net.Resolver) (exists bool, err error) {
	var values []string

	// nil check
	if r == nil {
		return false, errors.New("dns_checker: can't check record, resolver is nil (should never happen)")
	}

	// timeout context
	ctx, cancel := context.WithTimeout(context.Background(), timeoutSeconds*time.Second)
	defer cancel()

	// run appropriate query function
	switch recordType {
	// TXT records
	case txtRecord:
		values, err = r.LookupTXT(ctx, fqdn)

	// CNAME records
	case cnameRecord:
		val, e := r.LookupCNAME(ctx, fqdn)
		values, err = []string{val}, e

		// for comparison to domain, value should not end in `.`
		for i := range values {
			values[i] = strings.TrimSuffix(values[i], ".")
		}

	// any other (unsupported)
	default:
		return false, errors.New("dns_checker: unsupported dns record type (should never happen)")
	}

	// error check
	if err != nil {
		// if host wasn't found, this isn't a real error, it actually just means
		// the record does not exist
		dnsErr := new(net.DNSError)
		if errors.As(err, &dnsErr) && dnsErr.IsNotFound {
			return false, nil
		}

		// any other error, server failed
		return false, err
	}

	// check for desired value
	for i := range values {
		// if value found
		if values[i] == recordValue {
			return true, nil
		}
	}

	// records exist but desired value wasn't found
	return false, nil
}

// checkDnsRecordPropagationAllServices sends concurrent dns requests using all configured
// resolvers to check for the existence of the specified record. If both the
// functional resolver threshold and propagation thresholds are met, nil is
// returned, otherwise an error is returned.
func (service *Service) checkDnsRecordPropagationAllServices(fqdn string, recordValue string, recordType dnsRecordType) error {
	// if no resolvers (i.e. configured to skip)
	if service.dnsResolvers == nil {
		// sleep the skip wait and then return true (assume propagated)
		service.logger.Debugf("dns_checker: skipping check of %s and sleeping %d seconds", fqdn, int(service.skipWait.Seconds()))

		select {
		case <-service.shutdownContext.Done():
			// cancel if shutting down
			return errors.New("dns_checker: shutting down")

		case <-time.After(service.skipWait):
			// no-op, continue
		}

		return nil
	}

	// use waitgroup for concurrent checking
	var wg sync.WaitGroup
	resolverTotal := len(service.dnsResolvers)

	wg.Add(resolverTotal)
	wgResults := make(chan bool, resolverTotal)
	wgErrors := make(chan error, resolverTotal)

	// for each resolver pair, start a Go Routine
	for i := range service.dnsResolvers {
		go func(i int) {
			defer wg.Done()
			result, e := service.dnsResolvers[i].checkDnsRecord(fqdn, recordValue, recordType)
			if e != nil {
				// should not be getting errors, so write each to log if we get any
				service.logger.Errorf("dns_checker: check %s failed (%s)", fqdn, e)
			}
			wgResults <- result
			wgErrors <- e
		}(i)
	}

	// wait for all queries to finish
	wg.Wait()

	// close channels
	close(wgResults)
	close(wgErrors)

	// count functioning (total - any that returned err) & functional calc rate
	functionalCount := resolverTotal
	for err := range wgErrors {
		if err != nil {
			functionalCount--
		}
	}
	functionalRate := float32(functionalCount) / float32(resolverTotal)

	// count propagation confirmed result & calculate propagation rate
	propagationCount := 0
	for existed := range wgResults {
		if existed {
			propagationCount++
		}
	}
	propagationRate := float32(propagationCount) / float32(functionalCount)

	// debug log counts and rates
	functionalErr := fmt.Errorf("check %s: functional: %d (%.0f%%, min: %.0f%%)", fqdn, functionalCount, functionalRate*100, functioningRequirement*100)
	service.logger.Debugf("dns_checker: %s", functionalErr)
	propagationErr := fmt.Errorf("check %s: propagated: %d (%.0f%%, min: %.0f%%)", fqdn, propagationCount, propagationRate*100, propagationRequirement*100)
	service.logger.Debugf("dns_checker: %s", propagationErr)

	// return err if threshold(s) not met
	if functionalRate < functioningRequirement {
		return functionalErr
	} else if propagationRate < propagationRequirement {
		return propagationErr
	}

	return nil
}
