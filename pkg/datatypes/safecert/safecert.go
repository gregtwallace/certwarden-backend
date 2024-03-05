package safecert

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io"
	"legocerthub-backend/pkg/httpclient"
	"math/rand"
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
	httpClient  *httpclient.Client
	sync.RWMutex
}

// NewSafeCert returns a new SafeCert and also starts a routine to manage the
// cert's stapled OCSP response (if the cert supports it).
func NewSafeCert(httpClient *httpclient.Client, wg *sync.WaitGroup, shutdownCtx context.Context) *SafeCert {
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
	sc.RLock()
	defer sc.RUnlock()

	// while this is a "write", since ocspResp isn't written without Write
	// lock, this will never cause an issue; that is, ocspResp is fixed so this
	// should usually be a no-op except for the first Read()
	sc.cert.OCSPStaple = sc.ocspResp.Raw

	return sc.cert
}

// Update updates the certificate with the specified cert
func (sc *SafeCert) Update(tlsCert *tls.Certificate) {
	sc.Lock()
	defer sc.Unlock()

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

	if len(tlsCert.Certificate) >= 2 {
		sc.issuerCert, err = x509.ParseCertificate(tlsCert.Certificate[1])
		if err != nil {
			// clear issuer before panic
			sc.issuerCert = nil
			panic(err)
		}
	} else if sc.leafCert != nil && len(sc.leafCert.IssuingCertificateURL) > 0 {
		// issuer not in tlsCert but can try to get it if there are URLS in the
		// leaf certificate (randomly choose which URL to start with and then loop
		// through them until find working or run out of options)
		startIndex := rand.Intn(len(sc.leafCert.IssuingCertificateURL))
		issuerOk := false
		for i := 0; i < len(sc.leafCert.IssuingCertificateURL); i++ {
			issuerCertResp, err := sc.httpClient.Get(sc.leafCert.IssuingCertificateURL[(startIndex+i)%len(sc.leafCert.IssuingCertificateURL)])
			if err != nil {
				// this one failed
				continue
			}
			defer issuerCertResp.Body.Close()

			issuerCertBytes, err := io.ReadAll(issuerCertResp.Body)
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
	} else {
		// no issuer cert and unable to fetch it, set to nil
		sc.issuerCert = nil
	}

	// start new ocsp management
	sc.startOCSPManagement()
}
