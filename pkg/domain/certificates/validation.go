package certificates

import (
	"errors"
	"legocerthub-backend/pkg/acme"
	"legocerthub-backend/pkg/challenges"
	"legocerthub-backend/pkg/output"
	"legocerthub-backend/pkg/storage"
	"legocerthub-backend/pkg/validation"
)

var (
	// id
	ErrIdBad = errors.New("certificate id is invalid")

	// name
	ErrNameBad = errors.New("certificate name is not valid")

	// api key
	ErrApiKeyBad    = errors.New("api key is not valid (must be at least 10 chars in length)")
	ErrApiKeyNewBad = errors.New("api key (new) is not valid (must be at least 10 chars in length)")

	// domain
	ErrDomainBad = errors.New("domain or subject name not valid")
)

// GetCertificate returns the Certificate for the specified id.
func (service *Service) GetCertificate(id int) (Certificate, error) {
	// if id is not in valid range, it is definitely not valid
	if !validation.IsIdExistingValidRange(id) {
		service.logger.Debug(ErrIdBad)
		return Certificate{}, output.ErrValidationFailed
	}

	// get from storage
	account, err := service.storage.GetOneCertById(id)
	if err != nil {
		// special error case for no record found
		if err == storage.ErrNoRecord {
			service.logger.Debug(err)
			return Certificate{}, output.ErrNotFound
		} else {
			service.logger.Error(err)
			return Certificate{}, output.ErrStorageGeneric
		}
	}

	return account, nil
}

// nameValid returns if a name is valid (meets char requirements
// and is not in use in storage OR is in use by the specified certId)
func (service *Service) nameValid(certName string, certId *int) bool {
	// basic check
	if !validation.NameValid(certName) {
		return false
	}

	// make sure the name isn't already in use in storage
	cert, err := service.storage.GetOneCertByName(certName)
	if err == storage.ErrNoRecord {
		// no rows means name is not in use
		return true
	} else if err != nil {
		// any other error, invalid
		return false
	}

	// if the returned account is the account being edited, no error
	if certId != nil && cert.ID == *certId {
		return true
	}

	return false
}

// privateKeyIdValid returns true if the specified keyId is available
// for use by a certificate. If a certId is specified, this func will
// also return true if the keyId is the current keyId of the cert specified
// by the certId
func (service *Service) privateKeyIdValid(keyId int, certId *int) bool {
	if service.keys.KeyAvailable(keyId) {
		return true
	}

	// if there is a certId, check if the keyId is already assigned
	// to the cert
	if certId != nil {
		cert, err := service.storage.GetOneCertById(*certId)
		if err != nil {
			return false
		}

		// if certificate's key id matches keyId, valid
		if cert.CertificateKey.ID == keyId {
			return true
		}

	}

	return false
}

// subjectValid validates domain name and if it is a wildcard
// domain name it also verifies the method is dns-01
func subjectValid(domain string, challMethod challenges.Method) bool {
	// wild is only valid for Dns challenges
	wildOk := challMethod.ChallengeType == acme.ChallengeTypeDns01

	// check domain is valid
	return validation.DomainValid(domain, wildOk)
}

// subjectAltsValid validates each domain contained in the slice
// of subject alt domain names
func subjectAltsValid(alts []string, challMethod challenges.Method) bool {
	for _, altName := range alts {
		if !subjectValid(altName, challMethod) {
			return false
		}
	}

	return true
}
