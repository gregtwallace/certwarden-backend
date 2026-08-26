package safecert

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io"
	"math/rand/v2"
	"net"
	"net/http"
	"strings"
	"sync"

	"golang.org/x/crypto/ocsp"
)

// SafeCert is a struct to hold and manage a tls certificate
type SafeCert struct {
	cert         *tls.Certificate
	ocspResp     *ocsp.Response
	leafCert     *x509.Certificate
	issuerCert   *x509.Certificate
	stopOCSPMgmt context.CancelFunc

	shutdownWg  *sync.WaitGroup
	shutdownCtx context.Context
	httpClient  *http.Client

	mu sync.RWMutex
}

// NewSafeCert returns a new SafeCert and also starts a routine to manage the
// cert's stapled OCSP response (if the cert supports it).
func NewSafeCert(httpClient *http.Client, wg *sync.WaitGroup, shutdownCtx context.Context) *SafeCert {
	sc := &SafeCert{
		shutdownWg:  wg,
		shutdownCtx: shutdownCtx,
		httpClient:  httpClient,
	}

	return sc
}

// TlsCertFunc returns the function to get the tls.Certificate from SafeCert
func (sc *SafeCert) TlsCertFunc() func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	return func(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
		return sc.Read(), nil
	}
}

// Read returns the current tls certificate
func (sc *SafeCert) Read() *tls.Certificate {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	// while this is a "write", ocspResp isn't written without Write lock,
	// thus this will never cause an issue; that is, ocspResp is static here
	// so this should usually be a no-op except for the first Read()
	// nil check ensures missing ocspResp doesn't cause nil deref
	if sc.cert != nil && sc.ocspResp != nil {
		sc.cert.OCSPStaple = sc.ocspResp.Raw
	}

	return sc.cert
}

// ContainsHostname returns true if the certificate is valid for the specified hostname
// (DNS or IP). This includes the subject CommonName as well.
func (sc *SafeCert) ContainsHostname(hostname string) bool {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if sc.cert.Leaf == nil {
		return false
	}
	leafPtr := sc.cert.Leaf

	// check: common name
	if leafPtr.Issuer.CommonName != "" && leafPtr.Issuer.CommonName == hostname {
		return true
	}

	// check: dns names
	for _, dnsname := range leafPtr.DNSNames {
		if dnsname == hostname {
			return true
		}

		// wildcard? if so, check for match
		wildSuffix, isWild := strings.CutPrefix(dnsname, "*.")
		if isWild && strings.HasSuffix(hostname, wildSuffix) {
			return true
		}
	}

	// check: ip addresses
	ipAddr := net.ParseIP(hostname)
	if ipAddr != nil {
		for _, ip := range leafPtr.IPAddresses {
			if ip.Equal(ipAddr) {
				return true
			}
		}
	}

	return false
}

// Update updates the certificate with the specified cert
func (sc *SafeCert) Update(tlsCert *tls.Certificate) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	// set new cert & stop any previous OCSP routine
	sc.cert = tlsCert
	if sc.stopOCSPMgmt != nil {
		sc.stopOCSPMgmt()
	}

	// parse leaf & issuer certs (for OCSP)
	var err error
	if len(tlsCert.Certificate) >= 1 {
		sc.leafCert, err = x509.ParseCertificate(tlsCert.Certificate[0])
		if err != nil {
			// clear leaf and issuer before panic
			sc.leafCert = nil
			sc.issuerCert = nil
			panic(err)
		}
	} else {
		// no leaf, set to nil
		sc.leafCert = nil
	}

	switch {
	// we already have the issuer certificate
	case len(tlsCert.Certificate) >= 2:
		sc.issuerCert, err = x509.ParseCertificate(tlsCert.Certificate[1])
		if err != nil {
			// clear issuer before panic
			sc.issuerCert = nil
			panic(err)
		}

	// issuer not in tlsCert but can try to get it if there are URLS in the
	// leaf certificate (randomly choose which URL to start with and then loop
	// through them until find working or run out of options)
	case sc.leafCert != nil && len(sc.leafCert.IssuingCertificateURL) > 0:
		httpGetter := func(url string) (respBodyBytes []byte, err error) {
			response, err := sc.httpClient.Get(url)
			if err != nil {
				return nil, err
			}
			defer response.Body.Close()

			bodyBytes, err := io.ReadAll(response.Body)
			if err != nil {
				return nil, err
			}

			return bodyBytes, nil
		}

		// random start index
		indx := rand.IntN(len(sc.leafCert.IssuingCertificateURL))
		issuerOk := false
		for i := range len(sc.leafCert.IssuingCertificateURL) {
			// cycle through all indexes
			url := sc.leafCert.IssuingCertificateURL[(indx+i)%len(sc.leafCert.IssuingCertificateURL)]

			issuerCertBytes, err := httpGetter(url)
			if err != nil {
				// this one failed
				continue
			}

			sc.issuerCert, err = x509.ParseCertificate(issuerCertBytes)
			if err != nil {
				// this one failed
				continue
			}

			// don't bother verifying, just let ocsp fail later if CA is misconfigured and
			// sent wrong cert for some reason

			issuerOk = true
			break
		}

		if !issuerOk {
			// failed to fetch valid issuer cert
			sc.issuerCert = nil
		}

	// no issuer cert and unable to fetch it, set to nil
	default:
		sc.issuerCert = nil
	}

	// start new ocsp management
	sc.startOCSPManagement()
}
