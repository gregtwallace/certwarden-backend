package app

import (
	"crypto/tls"
	"crypto/x509"
	"sync"
	"time"
)

// generate the TLS Config for the app
func (app *Application) TlsConf() (*tls.Config, error) {
	tlsConf := &tls.Config{
		// func to return the TLS Cert from app
		GetCertificate: app.appCert.getTlsCertFunc(),
	}

	return tlsConf, nil
}

// serverCert is a struct to hold and manage the app's tls
// certificate
type appCert struct {
	cert *tls.Certificate
	mu   sync.RWMutex
}

// newAppCert creates the application cert struct on app and
// also starts a refresh go routine
func (app *Application) newAppCert() (*appCert, error) {
	sc := new(appCert)
	var err error

	// get cert from storage
	sc.cert, err = app.getAppCertFromStorage()
	if err != nil {
		return nil, err
	}

	// go routine to periodically try to refresh the cert
	go func() {
		var parsedCert *x509.Certificate
		var remainingValidTime time.Duration
		var sleepFor time.Duration
		var newCert *tls.Certificate

		for {
			// determine time to expiration
			sc.mu.RLock()
			parsedCert, err = x509.ParseCertificate(sc.cert.Certificate[0])
			sc.mu.RUnlock()
			if err != nil {
				app.logger.Panicf("lego tls certificate error: %s", err)
			}
			remainingValidTime = time.Until(parsedCert.NotAfter)

			// sleep duration based on expiration
			switch {
			// 45 days + (weekly)
			case remainingValidTime > (45 * time.Hour * 24):
				sleepFor = 7 * time.Hour * 24
			// 35 - 45 days (every other day)
			case remainingValidTime > (35 * time.Hour * 24):
				sleepFor = 2 * time.Hour * 24
			// anything else (daily)
			default:
				sleepFor = 1 * time.Hour * 24
			}

			// sleep
			time.Sleep(sleepFor)

			// attempt refresh from storage
			newCert, err = app.getAppCertFromStorage()
			if err != nil {
				// no op
				app.logger.Error("failed to update lego's certificate")
			} else {
				// no error, refresh cert
				sc.refreshCert(newCert)
				app.logger.Info("updated lego's certificate")
			}
		}
	}()

	return sc, nil
}

// getTlsCertFunc returns the function to get the tls.Certificate from appCert
func (ac *appCert) getTlsCertFunc() func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	return func(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
		ac.mu.RLock()
		defer ac.mu.RUnlock()
		return ac.cert, nil
	}
}

// refreshCert updates the certificate with the specified cert
func (ac *appCert) refreshCert(tlsCert *tls.Certificate) {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	ac.cert = tlsCert
}

// getAppCertFromStorage returns the current key/cert pair for the app
func (app *Application) getAppCertFromStorage() (*tls.Certificate, error) {
	// get key and cert for API server
	key, err := app.storage.GetOneKeyByName(*app.config.PrivateKeyName)
	if err != nil {
		return nil, err
	}

	certPem, err := app.storage.GetCertPemByName(*app.config.CertificateName)
	if err != nil {
		return nil, err
	}

	tlsCert, err := tls.X509KeyPair([]byte(certPem), []byte(*key.Pem))
	if err != nil {
		return nil, err
	}

	return &tlsCert, nil
}
