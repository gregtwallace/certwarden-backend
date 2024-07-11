package orders

import (
	"certwarden-backend/pkg/acme"
	"crypto/x509"
	"encoding/pem"
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

// validDates anlayzes the first cert in a pem chain and returns the valid from
// and valid to dates. If it fails to do so, it returns an error
func validDates(pemChain string) (validFrom int, validTo int, err error) {
	// decode first pem from chain
	cert, _ := pem.Decode([]byte(pemChain))

	// parse DER bytes
	derCert, err := x509.ParseCertificate(cert.Bytes)
	if err != nil {
		return 0, 0, err
	}

	// Note: Let's Encrypt sets "NotBefore" to one hour prior to issuance.
	// The code here is correct.
	return int(derCert.NotBefore.Unix()), int(derCert.NotAfter.Unix()), nil
}
