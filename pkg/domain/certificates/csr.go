package certificates

import (
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"legocerthub-backend/pkg/domain/private_keys/key_crypto"
)

// MakeCsrDer generates the CSR bytes for ACME to POST To a Finalize URL
func (cert *Certificate) MakeCsrDer() (csr []byte, err error) {
	// create Subject
	subj := pkix.Name{
		CommonName:         cert.Subject,
		Organization:       []string{cert.Organization},
		OrganizationalUnit: []string{cert.OrganizationalUnit},
		Country:            []string{cert.Country},
		Province:           []string{cert.State},
		Locality:           []string{cert.City},
		// unused: StreetAddress, PostalCode	[]string
		// unused: SerialNumber               string
		// unused: Names, ExtraNames					[]AttributeTypeAndValue
	}

	// CSR template to create CSR from
	template := x509.CertificateRequest{
		SignatureAlgorithm: cert.CertificateKey.Algorithm.CsrSigningAlg(),
		Subject:            subj,
		DNSNames:           cert.SubjectAltNames,
		// unused: EmailAddresses, IPAddresses, URIs, Attributes (deprecated), ExtraExtensions
	}

	// cert's private key for signing
	certKey, err := key_crypto.PemStringToKey(cert.CertificateKey.Pem, cert.CertificateKey.Algorithm)
	if err != nil {
		return nil, err
	}

	// create CSR
	csr, err = x509.CreateCertificateRequest(rand.Reader, &template, certKey)
	if err != nil {
		return nil, err
	}

	return csr, nil
}
