package acme

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/webpackager/resource/httplink"
)

// Certificate is a struct that contains the certificate pem returned by the ACME
// server as well as some information about the cert
type Certificate struct {
	pem         string
	notBefore   time.Time
	notAfter    time.Time
	chainRootCN string
}

func (c *Certificate) PEM() string          { return c.pem }
func (c *Certificate) NotBefore() time.Time { return c.notBefore }
func (c *Certificate) NotAfter() time.Time  { return c.notAfter }
func (c *Certificate) ChainRootCN() string  { return c.chainRootCN }

// responseToCertificate checks that the ACME server returned a valid cert/chain response
// when this client downloaded a certificate. If valid, the response is parsed into the
// Certificate struct. If not valid, an error is returned.
func responseToCertificate(bodyBytes []byte, headers http.Header) (*Certificate, error) {
	// this server only supports pem (application/pem-certificate-chain)
	contentType, _, err := mime.ParseMediaType(headers.Get("Content-type"))
	if err != nil {
		return nil, fmt.Errorf("error parsing Content-Type (%v)", err)
	}

	if !strings.EqualFold(contentType, "application/pem-certificate-chain") {
		return nil, errors.New("certificate content type is not application/pem-certificate-chain")
	}

	// validate ACME server didn't return malicious pem (see: RFC8555 s 11.4)
	pemCheck := string(bodyBytes)
	beginString := "-----BEGIN"
	mustBeFollowedBy := " CERTIFICATE"

	// if there is never a begin, invalid pem
	i := strings.Index(pemCheck, beginString)
	if i == -1 {
		return nil, errors.New("certificate missing BEGIN")
	}

	// check every begin to ensure it is followed by CERTIFICATE
	for ; i != -1; i = strings.Index(pemCheck, beginString) {
		pemCheck = pemCheck[(i + len(beginString)):]

		if !strings.HasPrefix(pemCheck, mustBeFollowedBy) {
			return nil, errors.New("certificate has at least one BEGIN not followed by CERTIFICATE")
		}
	}

	// make return struct
	cert := &Certificate{
		pem: string(bodyBytes),
	}

	// parse chain and do more extensive validation
	var tlsCert tls.Certificate
	var certDERBlock *pem.Block

	for {
		// try to decode each pem block
		certDERBlock, bodyBytes = pem.Decode(bodyBytes)

		// no more pem blocks
		if certDERBlock == nil {
			break
		}

		// check the found block
		if certDERBlock.Type == "CERTIFICATE" {
			tlsCert.Certificate = append(tlsCert.Certificate, certDERBlock.Bytes)
		} else {
			// found a PEM block that is NOT a certificate; error (should not happen due
			// to above BEGIN CERTIFICATE check, but just in case)
			return nil, errors.New("certificate has at least one non- CERTIFICATE pem block")
		}
	}

	// ensure at least one valid pem certificate block was decoded
	if len(tlsCert.Certificate) < 1 {
		return nil, errors.New("no valid CERTIFICATE pem blocks")
	}

	// parse topmost cert to get issuer CN (which should be root CN)
	derTopCert := tlsCert.Certificate[len(tlsCert.Certificate)-1]
	topCert, err := x509.ParseCertificate(derTopCert)
	if err != nil {
		return nil, errors.New("failed to parse top cert in chain")
	}

	cert.chainRootCN = topCert.Issuer.CommonName

	// parse bottom most cert (aka the Leaf) to get valid times
	derLeafCert := tlsCert.Certificate[0]
	leafCert, err := x509.ParseCertificate(derLeafCert)
	if err != nil {
		return nil, errors.New("failed to parse leaf cert in chain")
	}
	cert.notBefore = leafCert.NotBefore
	cert.notAfter = leafCert.NotAfter

	return cert, nil
}

// DownloadCertificate uses POST-as-GET to download a valid certificate from the specified
// url. If this first download passes basic sanity checks, it becomes the default cert/chain
// that is returned by this function. If the sanity check fails, alternate chain options are
// tried until one passes the sanity check. The first to pass the sanity check becomes the
// default cert/chain returned.
// If a preferredChain is specified, prefer the chain whose topmost certificate was issued
// from this Subject Common Name (assuming the basic sanity checks pass). If no match or the
// sanity check fails, the default chain is returned instead.
func (service *Service) DownloadCertificate(certificateUrl string, accountKey AccountKey, preferredChain string) (*Certificate, error) {
	var defaultChainCert *Certificate

	// POST-as-GET
	bodyBytes, defaultHeaders, err := service.PostAsGet(certificateUrl, accountKey)
	if err != nil {
		return nil, err
	}

	// validate initial response
	cert, err := responseToCertificate(bodyBytes, defaultHeaders)
	// if default chain didn't validate, log issue and continue to alts
	if err != nil {
		service.logger.Warnf("acme: %s default cert chain failed validation (see: rfc8555 s 11.4) (%s); will try others if available", certificateUrl, err)
		// don't return, instead try any alt chains first
	} else {
		// return now if no preferred chain specified, OR if this chain is the preferred chain
		if preferredChain == "" || strings.EqualFold(preferredChain, cert.ChainRootCN()) {
			return cert, nil
		}

		// validated but not the preferred chain we're seeking -- set this as default
		defaultChainCert = cert
	}

	// at this point, either the default chain didn't validate, or it wasn't preferred, so
	// proceed to checking the alt options

	// make slice of the URLs for alt chains
	altChainUrls := []*url.URL{}
	// check each Link header
	for _, headerLink := range defaultHeaders.Values("Link") {
		// each Link header can potentially have multiple URLs, so check them all
		httpLinks, err := httplink.Parse(headerLink)
		if err != nil {
			// if failed to parse, discard this Link header and continue
			service.logger.Warnf("acme: %s sent bad Link header in certificate download response (%s)", err)
			continue
		}

		// check if each Link is the right rel type, add it to the list of alt chain URLs
		for _, httpLink := range httpLinks {
			if strings.EqualFold(httpLink.Params.Get("rel"), "alternate") {
				altChainUrls = append(altChainUrls, httpLink.URL)
			}
		}
	}

	// check alt chain URLs that are available
	for _, altChainURL := range altChainUrls {

		// POST-as-GET the alt chain
		bodyBytes, headers, err := service.PostAsGet(altChainURL.String(), accountKey)
		if err != nil {
			service.logger.Warnf("acme: failed to fetch alt chain %s (%s); will try other available chains", altChainURL.String(), err)
			// don't return, continue to next alt to keep trying
			continue
		}

		// validate alt chain
		cert, err = responseToCertificate(bodyBytes, headers)
		// if default chain didn't validate, log issue
		if err != nil {
			service.logger.Warnf("acme: %s alt cert chain failed validation (see: rfc8555 s 11.4) (%s); will try others if available", altChainURL.String(), err)
			// don't return, continue to next alt to keep trying
			continue
		}

		// this chain is now confirmed valid

		// return this chain if no preferred chain specified, OR if this chain is the preferred chain
		if preferredChain == "" || strings.EqualFold(preferredChain, cert.ChainRootCN()) {
			return cert, nil
		}

		// set defaultChain if there isn't one yet
		if defaultChainCert == nil {
			defaultChainCert = cert
		}

		// preferred chain is set and this chain did not match, continue loop to keep trying alts
	}

	// made it through all of the alt chains without success (either no valid chain at all and/or no chain match preferred chain

	// check if defaultChain was set by any valid chain; if not, error as no valid chains
	if defaultChainCert == nil {
		return nil, fmt.Errorf("acme: no valid cert chains found for %s", certificateUrl)
	}

	// at least one valid chain was found (even though preferred didn't match), return that one
	service.logger.Warnf("acme: went through all alt chains of %s without preferred chain match, returning default chain", certificateUrl)
	return defaultChainCert, nil
}
