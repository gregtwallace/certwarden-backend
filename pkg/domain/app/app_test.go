package app

import (
	"certwarden-backend/pkg/datatypes/safecert"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net"
	"time"
)

// NewTestSafecert generates a self-signed certificate for use in testing; the first dnsName or
// ip address is promoted to CommonName
// TODO: move to `helpers_test` package as I'm sure this will come up again in other pkg tests
func NewTestSafecert(dnsNames, ips []string) (*safecert.SafeCert, error) {
	if len(dnsNames) == 0 && len(ips) == 0 {
		return nil, errors.New("no dnsnames or ips specified")
	}

	// validate ips
	ipAddrs := []net.IP{}
	for _, ip := range ips {
		ipAddrs = append(ipAddrs, net.ParseIP(ip))
	}

	cn := ""
	if len(dnsNames) != 0 {
		cn = dnsNames[0]
		dnsNames = dnsNames[1:]
	}
	// We deliberately ignore testing IP in CommonName as CAB generally requires they also be listed in the SAN:
	// https://cabforum.org/working-groups/server/guidance-ip-addresses-certificates/

	// make the private key
	pKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(pKey)

	// make the certificate
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, errors.New("failed to generate serial number: " + err.Error())
	}

	tmpl := &x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               pkix.Name{CommonName: cn},
		SignatureAlgorithm:    x509.SHA256WithRSA,
		NotBefore:             time.Now().Add(-30 * time.Minute),
		NotAfter:              time.Now().Add(time.Hour * 24 * 30), // valid for a month
		BasicConstraintsValid: true,
		DNSNames:              dnsNames,
		IPAddresses:           ipAddrs,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &pKey.PublicKey, pKey)
	if err != nil {
		return nil, fmt.Errorf("couldnt make cert (%w)", err)
	}

	// make pem and cert
	k := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: []byte(privateKeyBytes),
	})

	c := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: []byte(certDER),
	})

	// make tls certificate
	tlsCert, err := tls.X509KeyPair(c, k)
	if err != nil {
		return nil, fmt.Errorf("failed to make x509 key pair (%w)", err)
	}

	sc := safecert.NewSafeCert(nil, nil, context.TODO())
	sc.Update(&tlsCert)

	return sc, nil
}
