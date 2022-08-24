package frontend

import "crypto/tls"

// generate the TLS Config for the app
func (service *Service) tlsConf() (*tls.Config, error) {
	tlsConf := &tls.Config{
		// func to return the TLS Cert from app
		GetCertificate: service.httpsCert.TlsCertFunc(),
	}

	return tlsConf, nil
}
