package orders

import (
	"certwarden-backend/pkg/acme"
	"time"
)

// this relates to the order's issued certificate, not to be conflated with the 'certificates'
// package

// CertPayload is the data to store for an issued certificate
type CertPayload struct {
	AcmeCert    *acme.Certificate
	RenewalInfo *renewalInfo
	UpdatedAt   time.Time
}

// savePemChain calls a func to determine the valid from and to dates for the issued pem chain
// and then saves the pem chain and valid dates to storage
func (j *orderFulfillJob) saveAcmeCert(orderId int, cert *acme.Certificate, acmeARI *acme.ACMERenewalInfo) (err error) {
	// if acme ARI is available, use it, else make a sane default
	var ari *renewalInfo
	if acmeARI != nil {
		ari = &renewalInfo{
			SuggestedWindow: struct {
				Start time.Time "json:\"start\""
				End   time.Time "json:\"end\""
			}{
				Start: acmeARI.SuggestedWindow.Start,
				End:   acmeARI.SuggestedWindow.End,
			},
			ExplanationURL: acmeARI.ExplanationURL,
			RetryAfter:     &acmeARI.RetryAfter,
		}
	} else {
		ari = MakeRenewalInfo(cert.NotBefore(), cert.NotAfter())
	}

	// payload to save
	payload := &CertPayload{
		AcmeCert:    cert,
		RenewalInfo: ari,
		UpdatedAt:   time.Now(),
	}

	// save to storage
	err = j.service.storage.UpdateOrderCert(orderId, payload)
	if err != nil {
		return err
	}

	return nil
}
