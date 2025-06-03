package acme

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

var ErrUnsupportedARI = errors.New("acme: server does not support ARI (directory missing 'renewalInfo' key)")

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
	retryAfter := headers.Get("Retry-After")
	if retryAfter == "" {
		return nil, errors.New("acme: ari response missing Retry-After header")
	}

	// check if header was in seconds and ensure > 0
	parsedOk := false
	secs, err := strconv.Atoi(retryAfter)
	if err == nil && secs > 0 {
		ari.RetryAfter = time.Now().Add(time.Duration(secs) * time.Second).Round(time.Second).UTC()
		parsedOk = true
	} else {
		// wasn't in seconds, try to parse date and ensure > 0
		t, err := http.ParseTime(retryAfter)
		if err == nil {
			until := time.Until(t)
			if until > 0 {
				ari.RetryAfter = t.Round(time.Second).UTC()
				parsedOk = true
			}
		}
	}

	if !parsedOk {
		return nil, fmt.Errorf("acme: ari response Retry-After header value '%s' could not be parsed", retryAfter)
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

// GetACMERenewalInfo sends an unauthenticated GET request to retrieve the ARI information
// for the specified certificate PEM.
func (service *Service) GetACMERenewalInfo(certPem string) (*ACMERenewalInfo, error) {
	// only some servers support this
	if !service.SupportsARIExtension() {
		return nil, ErrUnsupportedARI
	}

	// decode and parse the pem to a Certificate
	certBlock, _ := pem.Decode([]byte(certPem))
	if certBlock == nil {
		return nil, errors.New("acme: cert pem block is nil")
	}

	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("acme: cert pem block failed to parse (%v)", err)
	}

	// s 4.3 forbids checking renewal info after cert expiration
	if cert.NotAfter.Before(time.Now()) {
		return nil, fmt.Errorf("acme: cert expired, ari check is forbidden")
	}

	// assemble the link and do GET
	akiStr := base64.RawURLEncoding.EncodeToString(cert.AuthorityKeyId)
	serialStr := base64.RawURLEncoding.EncodeToString(cert.SerialNumber.Bytes())
	url := *service.dir.RenewalInfo + "/" + akiStr + "." + serialStr

	resp, headers, err := service.get(url)
	if err != nil {
		return nil, fmt.Errorf("acme: ari get failed (%v)", err)
	}

	// unmarshal response
	ari, err := unmarshalACMERenewalInfo(resp, headers)
	if err != nil {
		return nil, err
	}

	return ari, nil
}
