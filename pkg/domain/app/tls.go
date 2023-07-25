package app

import (
	"crypto/tls"
	"crypto/x509"
	"legocerthub-backend/pkg/datatypes"
	"time"
)

// generate the TLS Config for the app
func (app *Application) tlsConf() (*tls.Config, error) {
	tlsConf := &tls.Config{
		// func to return the TLS Cert from app
		GetCertificate: app.httpsCert.TlsCertFunc(),
	}

	return tlsConf, nil
}

// newAppCert creates the application cert struct on app and
// also starts a refresh go routine
func (app *Application) newAppCert() (*datatypes.SafeCert, error) {
	sc := new(datatypes.SafeCert)
	var err error

	// get cert from storage and set SafeCert
	cert, err := app.getAppCertFromStorage()
	if err != nil {
		return nil, err
	}
	sc.Update(cert)

	// go routine to periodically try to refresh the cert
	// log start and update wg
	app.logger.Info("starting https cert refresh service")
	app.shutdownWaitgroup.Add(1)

	go func() {
		var parsedCert *x509.Certificate
		var remainingValidTime time.Duration
		var sleepFor time.Duration
		var newCert *tls.Certificate

		for {
			// determine time to expiration
			parsedCert, err = x509.ParseCertificate(sc.Read().Certificate[0])
			if err != nil {
				app.logger.Panicf("tls certificate error: %s", err)
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

			// sleep or wait for shutdown context to be done
			select {
			case <-app.shutdownContext.Done():
				// close routine
				app.logger.Info("https cert refresh service shutdown complete")
				app.shutdownWaitgroup.Done()
				return

			case <-time.After(sleepFor):
				// sleep and retry
			}

			// attempt refresh from storage
			newCert, err = app.getAppCertFromStorage()
			if err != nil {
				// no op
				app.logger.Error("failed to update lego's certificate")
			} else {
				// no error, refresh cert
				sc.Update(newCert)
				app.logger.Info("updated lego's certificate")
			}
		}
	}()

	return sc, nil
}

// getAppCertFromStorage returns the current key/cert pair for the app
func (app *Application) getAppCertFromStorage() (*tls.Certificate, error) {
	// get key and cert for API server
	keyPem, err := app.storage.GetKeyPemByName(*app.config.PrivateKeyName)
	if err != nil {
		return nil, err
	}

	certPem, _, err := app.storage.GetCertPemByName(*app.config.CertificateName)
	if err != nil {
		return nil, err
	}

	tlsCert, err := tls.X509KeyPair([]byte(certPem), []byte(keyPem))
	if err != nil {
		return nil, err
	}

	return &tlsCert, nil
}
