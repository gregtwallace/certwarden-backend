package certificates

import (
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
)

var errSubjectMissing = errors.New("certificate csr subject missing (required)")

// MakeCsrDer generates the CSR bytes for ACME to POST To a Finalize URL
func (cert *Certificate) MakeCsrDer() (csr []byte, err error) {
	// subject is mandatory (and is also used as common name)
	if cert.Subject == nil {
		return nil, errSubjectMissing
	}

	// create Subject
	subj := pkix.Name{
		CommonName:         *cert.Subject,
		Organization:       []string{derefString(cert.Organization)},
		OrganizationalUnit: []string{derefString(cert.OrganizationalUnit)},
		Country:            []string{derefString(cert.Country)},
		Province:           []string{derefString(cert.State)},
		Locality:           []string{derefString(cert.City)},
		// unused: StreetAddress, PostalCode	[]string
		// unused: SerialNumber               string
		// unused: Names, ExtraNames					[]AttributeTypeAndValue
	}

	// CSR template to create CSR from
	template := x509.CertificateRequest{
		SignatureAlgorithm: cert.PrivateKey.Algorithm.SignatureAlgorithm,
		Subject:            subj,
		DNSNames:           *cert.SubjectAltNames,
		// unused: EmailAddresses, IPAddresses, URIs, Attributes (deprecated), ExtraExtensions
	}

	// cert's private key for signing
	certKey, err := cert.PrivateKey.CryptoKey()
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

// derefString dereferences the string pointer, or if the string pointer
// is nil, it returns an empty string
func derefString(s *string) string {
	if s == nil {
		return ""
	}

	return *s
}
