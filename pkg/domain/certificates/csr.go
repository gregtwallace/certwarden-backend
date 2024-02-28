package certificates

import (
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"legocerthub-backend/pkg/domain/private_keys/key_crypto"
)

// MakeCsrDer generates the CSR bytes for ACME to POST To a Finalize URL
func (cert *Certificate) MakeCsrDer() (csr []byte, err error) {
	// omit empty fields
	org := []string{}
	if cert.Organization != "" {
		org = append(org, cert.Organization)
	}

	ou := []string{}
	if cert.OrganizationalUnit != "" {
		ou = append(ou, cert.OrganizationalUnit)
	}

	country := []string{}
	if cert.Country != "" {
		country = append(country, cert.Country)
	}

	province := []string{}
	if cert.State != "" {
		province = append(province, cert.State)
	}

	locality := []string{}
	if cert.City != "" {
		locality = append(locality, cert.City)
	}

	// create Subject
	subj := pkix.Name{
		CommonName:         cert.Subject,
		Organization:       org,
		OrganizationalUnit: ou,
		Country:            country,
		Province:           province,
		Locality:           locality,
		// unused: StreetAddress, PostalCode	[]string
		// unused: SerialNumber               string
		// unused: Names, ExtraNames					[]AttributeTypeAndValue
	}

	// convert any extra extensions to proper pkix obj
	extraExts := []pkix.Extension{}
	for i := range cert.CSRExtraExtensions {
		extraExts = append(extraExts, cert.CSRExtraExtensions[i].Extension)
	}

	// CSR template to create CSR from
	template := x509.CertificateRequest{
		SignatureAlgorithm: cert.CertificateKey.Algorithm.CsrSigningAlg(),
		Subject:            subj,
		DNSNames:           append([]string{cert.Subject}, cert.SubjectAltNames...),
		// unused: EmailAddresses, IPAddresses, URIs, Attributes (deprecated)
		ExtraExtensions: extraExts,
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
