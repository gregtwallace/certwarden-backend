package app

import (
	"crypto/tls"
	"errors"
	"fmt"
)

// tlsConf returns the app's tls Config for an https web server to use.
func (app *Application) tlsConf() *tls.Config {
	tlsConf := &tls.Config{
		// func to return the TLS Cert from app
		GetCertificate: app.httpsCert.TlsCertFunc(),
	}

	return tlsConf
}

// HttpsCertificateName returns the db `Name` of the certificate this app is using.
// This allows orders package to call the refresh function whenever the app's
// certificate is reordered.
func (app *Application) HttpsCertificateName() *string {
	return app.config.CertificateName
}

// LoadHttpsCertificate fetches the most recent order for the app's certificate
// and loads it as the app's https certificate. If there is an error, the app
// retains its previous certificate.
func (app *Application) LoadHttpsCertificate() error {
	// if not running in https, this is no-op
	if !app.IsHttps() {
		return errors.New("cannot load https certificate, server is in http mode")
	}

	// get order for this app
	order, err := app.storage.GetCertNewestValidOrderByName(*app.config.CertificateName)
	if err != nil {
		return err
	}

	// nil check of key
	if order.FinalizedKey == nil {
		return errors.New("tls key pem is empty")
	}

	// nil check of cert pem
	if order.Pem == nil {
		return errors.New("tls cert pem is empty")
	}

	// make tls certificate
	tlsCert, err := tls.X509KeyPair([]byte(*order.Pem), []byte(order.FinalizedKey.Pem))
	if err != nil {
		return fmt.Errorf("failed to make x509 key pair (%s)", err)
	}

	// update certificate
	app.httpsCert.Update(&tlsCert)

	return nil
}
