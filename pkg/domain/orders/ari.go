package orders

import (
	"certwarden-backend/pkg/randomness"
	"encoding/json"
	"time"
)

// expiringAfterElapsedRatio is the ratio of a certificate's elapsed validity / total validity
// after which the certificate should be considered expiring
const (
	// shortLivedValidityThreshold is the amount of validity time under which this application will consider
	// a certificate to be "short-lived"
	shortLivedValidityThreshold = 10 * 24 * time.Hour
	// expiringRemainingValidFraction is the fraction of validity remaining under which this application will
	// consider a certificate expiring (for a "regular" (non-"short-lived") certificate)
	expiringRemainingValidFraction = 0.333
	// expiringShortLivedRemainingValidFraction is the fraction of validity remaining under which this application will
	// consider a "short-lived" certificate expiring
	expiringShortLivedRemainingValidFraction = 0.5
)

// renewalInfo is a struct to hold information that instructs Cert Warden when renewal
// should be attempted by the auto ordering task; it is similar but NOT identical the
// one specified in the ACME ARI specification
type renewalInfo struct {
	SuggestedWindow struct {
		Start time.Time `json:"start"`
		End   time.Time `json:"end"`
	} `json:"suggestedWindow"`
	ExplanationURL *string    `json:"explanationURL,omitempty"`
	RetryAfter     *time.Time `json:"retryAfter,omitempty"`
}

// UpdateRenewalInfoPayload is the object to update ARI in the database
type UpdateRenewalInfoPayload struct {
	OrderID     int
	RenewalInfo *renewalInfo
	UpdatedAt   int
}

// UnmarshalRenewalInfo unmarshals data into the renewalInfo struct; it does some basic checking
// of values and if any checks fail, null is returned instead
func UnmarshalRenewalInfo(data []byte) *renewalInfo {
	ri := &renewalInfo{}
	err := json.Unmarshal(data, ri)
	if err != nil {
		return nil
	}

	// ZeroTime suggests data was invalid, so return nil
	if ri.SuggestedWindow.Start.IsZero() || ri.SuggestedWindow.End.IsZero() {
		return nil
	}

	return ri
}

// MakeRenewalInfo returns a renewalInfo struct based on a certificate's validity. This is essentially
// a local ARI calculation, which will be used when the ACME Server does not implement ARI or is not
// returning valid ARI responses.
func MakeRenewalInfo(validFrom, validTo time.Time) *renewalInfo {
	// determine if the cert is "short lived"
	validDuration := validTo.Sub(validFrom)
	shortLived := false
	if validTo.Sub(validFrom) < shortLivedValidityThreshold {
		shortLived = true
	}

	// calculate validity threshold (for the approx. midpoint of the renewal window)
	var validityThreshold time.Time
	if shortLived {
		validityThreshold = validTo.Add(-1 * time.Duration(float64(validDuration)*float64(expiringShortLivedRemainingValidFraction)))
	} else {
		validityThreshold = validTo.Add(-1 * time.Duration(float64(validDuration)*float64(expiringRemainingValidFraction)))
	}

	// calculate start and end time (includes adding some jitter too)
	var startT time.Time
	var endT time.Time
	if shortLived {
		startT = validityThreshold.Add(-1 * 4 * time.Hour).Add(time.Duration(randomness.GenerateInsecureInt(60)) * time.Second)
		endT = validityThreshold.Add(4 * time.Hour).Add(time.Duration(randomness.GenerateInsecureInt(60)) * time.Second)
	} else {
		startT = validityThreshold.Add(-1 * 24 * time.Hour).Add(time.Duration(randomness.GenerateInsecureInt(60)) * time.Second)
		endT = validityThreshold.Add(24 * time.Hour).Add(time.Duration(randomness.GenerateInsecureInt(60)) * time.Second)
	}

	// return struct
	return &renewalInfo{
		SuggestedWindow: struct {
			Start time.Time "json:\"start\""
			End   time.Time "json:\"end\""
		}{
			Start: startT,
			End:   endT,
		},
	}
}
