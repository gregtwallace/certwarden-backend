package dns_checker

// CheckTXT queries each configured resolver and returns true if the record is found on
// all servers (in accord with error thresholds), otherwise it returns false.
func (service *Service) CheckTXT(fqdn string, recordValue string) (propagated bool, err error) {
	return service.checkDnsRecordAllServices(fqdn, recordValue, txtRecord)
}
