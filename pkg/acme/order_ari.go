package acme

import (
	"certwarden-backend/pkg/validation"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// ACMERenewalInfo contains the ACME Renewal Info (ARI) response
type ACMERenewalInfo struct {
	SuggestedWindow struct {
		Start time.Time `json:"start"`
		End   time.Time `json:"end"`
	} `json:"suggestedWindow"`
	ExplanationURL *string   `json:"explanationURL"`
	RetryAfter     time.Time `json:"-"` // parsed from Retry-After header
}

// unmarshalACMERenewalInfo attempts to unmarshal
func unmarshalACMERenewalInfo(jsonResp json.RawMessage, headers http.Header) (ari *ACMERenewalInfo, _ error) {
	err := json.Unmarshal(jsonResp, &ari)
	if err != nil {
		return nil, err
	}

	// if window has a 0 value, didn't unmarshal a valid response
	if ari.SuggestedWindow.Start.IsZero() || ari.SuggestedWindow.End.IsZero() {
		return nil, errors.New("acme: ari response invalid (suggested window has 0 time value)")
	}

	// s 4.2 dictates "A RenewalInfo object in which the end timestamp equals or precedes
	// the start timestamp is invalid" and requires it to be treated as if it was not received
	if ari.SuggestedWindow.End.Equal(ari.SuggestedWindow.Start) || ari.SuggestedWindow.End.Before(ari.SuggestedWindow.Start) {
		return nil, errors.New("acme: ari response invalid (suggested window not valid)")
	}

	// get Retry-After value from header
	retryAfterHeaderStr := headers.Get("Retry-After")
	if retryAfterHeaderStr == "" {
		return nil, errors.New("acme: ari response missing Retry-After header")
	}

	// parse and validate Retry-After
	ari.RetryAfter, err = validation.ParseRetryAfter(retryAfterHeaderStr)
	if err != nil {
		return nil, fmt.Errorf("acme: ari response Retry-After header value '%s' could not be parsed (%w)", retryAfterHeaderStr, err)
	}

	// s 4.3.2 - impose limits on Retry-After
	// if < 1 minute, set to 1 minute
	if ari.RetryAfter.Add(-1 * time.Minute).Before(time.Now()) {
		ari.RetryAfter = time.Now().Add(1 * time.Minute)
	} else if time.Now().Before(ari.RetryAfter.Add(-24 * time.Hour)) {
		// if > 24 hours, set to 24 hours
		ari.RetryAfter = time.Now().Add(24 * time.Hour)
	}

	return ari, nil
}

// SupportsARIExtension returns true if the ACME Service supports ARI (ACME Renewal Info) extension
func (service *Service) SupportsARIExtension() bool {
	return service.dir.RenewalInfo != nil
}

// ACMERenewalInfoIdentifier returns the unique identifer for the passed in cert
// as specified in the ACME ARI Draft v8 s 4.1
func ACMERenewalInfoIdentifier(cert *x509.Certificate) (string, error) {
	// aki (1st portion)
	if len(cert.AuthorityKeyId) == 0 {
		return "", errorARIIdentifierFailed("cert has no authority key id")
	}
	akiStr := base64.RawURLEncoding.EncodeToString(cert.AuthorityKeyId)

	// serial (2nd portion)
	// This field is required, and therefore should never be nil, but check just in case
	if cert.SerialNumber == nil {
		return "", errorARIIdentifierFailed("cert has no serial number (malformed)")
	}
	serialBytes := cert.SerialNumber.Bytes()

	// serial should always be non-negative (https://datatracker.ietf.org/doc/html/rfc5280#section-4.1.2.2)
	sign := cert.SerialNumber.Sign()
	switch {
	case sign < 0:
		// invalid (negative) serial number
		return "", errorARIIdentifierFailed("cert serial number is invalid (negative)")

	case sign > 0 && len(serialBytes) > 0 && serialBytes[0]&0x80 != 0:
		// serial is positive but looks negative since the first bit is high, so prepend a 0x00 byte
		serialBytes = append([]byte{0x00}, serialBytes...)

	case sign == 0:
		// serial is (confusingly, but still valid) 0
		serialBytes = []byte{0x00}

	default:
		// do nothing -- sign is positive and does not need to be modified
	}

	// assemble the unique id
	serialStr := base64.RawURLEncoding.EncodeToString(serialBytes)
	return akiStr + "." + serialStr, nil
}

// GetACMERenewalInfo sends an unauthenticated GET request to retrieve the ARI information
// for the specified certificate PEM.
func (service *Service) GetACMERenewalInfo(certPem string) (*ACMERenewalInfo, error) {
	// only some servers support this
	if !service.SupportsARIExtension() {
		return nil, ErrARIUnsupported
	}

	// decode and parse the pem to a Certificate
	certBlock, _ := pem.Decode([]byte(certPem))
	if certBlock == nil {
		return nil, errors.New("acme: cert pem block is nil")
	}

	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("acme: cert pem block failed to parse (%w)", err)
	}

	// s 4.3 forbids checking renewal info after cert expiration
	if cert.NotAfter.Before(time.Now()) {
		return nil, fmt.Errorf("acme: cert expired, ari check is forbidden")
	}

	// create link and do GET
	id, err := ACMERenewalInfoIdentifier(cert)
	if err != nil {
		return nil, err
	}

	url := *service.dir.RenewalInfo + "/" + id
	resp, headers, err := service.get(url)
	if err != nil {
		return nil, fmt.Errorf("acme: ari get failed (%w)", err)
	}

	// unmarshal response
	ari, err := unmarshalACMERenewalInfo(resp, headers)
	if err != nil {
		return nil, err
	}

	return ari, nil
}
