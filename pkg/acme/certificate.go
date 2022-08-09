package acme

import "encoding/pem"

// revokePayload is the struct to send to ACME to perform a certificate revocation
type revokePayload struct {
	Certificate string `json:"certificate"`
	Reason      int    `json:"reason"`
}

// RevokeCertificate revokes the certificate pem (or pem chain) that is passed in
// using the specfied reason code and account. If ACME revokes the cert, no content
// is returned. Otherwise, if revocation fails, error is returned.
func (service *Service) RevokeCertificate(pemCert string, reasonCode int, accountKey AccountKey) (err error) {
	// decode pem (if a chain, take the first cert and discard the rest)
	pemBlock, _ := pem.Decode([]byte(pemCert))

	// encode the pem bytes for ACME
	derCert := encodeString(pemBlock.Bytes)

	// make payload
	payload := revokePayload{
		Certificate: derCert,
		Reason:      reasonCode,
	}

	// revoke
	_, _, err = service.postToUrlSigned(payload, service.dir.RevokeCert, accountKey)
	if err != nil {
		return err
	}

	return nil
}
