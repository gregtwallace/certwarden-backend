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
func unmarshalACMERenewalInfo(jsonResp json.RawMessage, headers http.Header) (ari ACMERenewalInfo, _ error) {
	err := json.Unmarshal(jsonResp, &ari)
	if err != nil {
		return ACMERenewalInfo{}, err
	}

	// get Retry-After value from header
	retryAfter := headers.Get("Retry-After")
	if retryAfter == "" {
		return ACMERenewalInfo{}, errors.New("acme: ari response missing Retry-After header")
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
		return ACMERenewalInfo{}, fmt.Errorf("acme: ari response Retry-After header value '%s' could not be parsed", retryAfter)
	}

	return ari, nil
}

// GatACMERenewalInfo sends an unauthenticated GET request to retrieve the ARI information
// for the specified certificate PEM.
func (service *Service) GatACMERenewalInfo(certPem string) (ACMERenewalInfo, error) {
	// only some servers support this, ensure its in the directory
	if service.dir.RenewalInfo == nil {
		return ACMERenewalInfo{}, errors.New("acme: server does not support ARI (directory missing 'renewalInfo' key)")
	}

	// decode and parse the pem to a Certificate
	certBlock, _ := pem.Decode([]byte(certPem))
	if certBlock == nil {
		return ACMERenewalInfo{}, errors.New("acme: cert pem block is nil")
	}

	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return ACMERenewalInfo{}, fmt.Errorf("acme: cert pem block failed to parse (%v)", err)
	}

	// assemble the link and do GET
	akiStr := base64.RawURLEncoding.EncodeToString(cert.AuthorityKeyId)
	serialStr := base64.RawURLEncoding.EncodeToString(cert.SerialNumber.Bytes())
	url := *service.dir.RenewalInfo + "/" + akiStr + "." + serialStr

	resp, headers, err := service.get(url)
	if err != nil {
		return ACMERenewalInfo{}, fmt.Errorf("acme: ari get failed (%v)", err)
	}

	// unmarshal response
	ari, err := unmarshalACMERenewalInfo(resp, headers)
	if err != nil {
		return ACMERenewalInfo{}, err
	}

	return ari, nil
}
