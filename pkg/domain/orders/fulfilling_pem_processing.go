package orders

import (
	"certwarden-backend/pkg/acme"
	"time"
)

// this relates to the order's issued certificate, not to be conflated with the 'certificates'
// package

// CertPayload is the data to store for an issued certificate
type CertPayload struct {
	AcmeCert  *acme.Certificate
	UpdatedAt time.Time
}

// savePemChain calls a func to determine the valid from and to dates for the issued pem chain
// and then saves the pem chain and valid dates to storage
func (j *orderFulfillJob) saveAcmeCert(orderId int, cert *acme.Certificate) (err error) {
	// payload to save
	payload := &CertPayload{
		AcmeCert:  cert,
		UpdatedAt: time.Now(),
	}

	// save to storage
	err = j.service.storage.UpdateOrderCert(orderId, payload)
	if err != nil {
		return err
	}

	return nil
}
