package datatypes

import (
	"crypto/tls"
	"sync"
)

// SafeCert is a struct to hold and manage a tls certificate
type SafeCert struct {
	cert *tls.Certificate
	sync.RWMutex
}

// TlsCertFunc returns the function to get the tls.Certificate from SafeCert
func (ac *SafeCert) TlsCertFunc() func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	return func(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
		ac.RLock()
		defer ac.RUnlock()

		return ac.cert, nil
	}
}

// Read returns the current tls certificate
func (sc *SafeCert) Read() *tls.Certificate {
	sc.RLock()
	defer sc.RUnlock()

	return sc.cert
}

// Update updates the certificate with the specified cert
func (ac *SafeCert) Update(tlsCert *tls.Certificate) {
	ac.Lock()
	defer ac.Unlock()

	ac.cert = tlsCert
}
