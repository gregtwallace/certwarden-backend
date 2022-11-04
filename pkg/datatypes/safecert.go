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
func (sc *SafeCert) TlsCertFunc() func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	return func(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
		return sc.Read(), nil
	}
}

// Read returns the current tls certificate
func (sc *SafeCert) Read() *tls.Certificate {
	sc.RLock()
	defer sc.RUnlock()

	return sc.cert
}

// Update updates the certificate with the specified cert
func (sc *SafeCert) Update(tlsCert *tls.Certificate) {
	sc.Lock()
	defer sc.Unlock()

	sc.cert = tlsCert
}
