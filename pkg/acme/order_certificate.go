package acme

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/webpackager/resource/httplink"
)

// validateCertificate checks that the ACME server returned a valid cert/chain response
// when this client downloaded a certificate. If valid, the Subject Common Name for the
// issuer of the topmost certificate in the chain is returned. If not valid, an error
// is returned.
func validateCertificate(bodyBytes []byte, headers http.Header) (rootCN string, err error) {
	// this server only supports pem (application/pem-certificate-chain)
	contentType := headers.Get("Content-type")
	if contentType != "application/pem-certificate-chain" {
		return "", errors.New("certificate content type is not application/pem-certificate-chain")
	}

	// validate ACME server didn't return malicious pem (see: RFC8555 s 11.4)
	pemCheck := string(bodyBytes)
	beginString := "-----BEGIN"
	mustBeFollowedBy := " CERTIFICATE"

	// if there is never a begin, invalid pem
	i := strings.Index(pemCheck, beginString)
	if i == -1 {
		return "", errors.New("certificate missing BEGIN")
	}

	// check every begin to ensure it is followed by CERTIFICATE
	for ; i != -1; i = strings.Index(pemCheck, beginString) {
		pemCheck = pemCheck[(i + len(beginString)):]

		if !strings.HasPrefix(pemCheck, mustBeFollowedBy) {
			return "", errors.New("certificate has at least one BEGIN not followed by CERTIFICATE")
		}
	}

	// parse chain and do more extensive validation
	var cert tls.Certificate
	var certDERBlock *pem.Block

	// copy body to avoid mutating it
	certPEMBlock := make([]byte, len(bodyBytes))
	_ = copy(certPEMBlock, bodyBytes)

	for {
		// try to decode each pem block
		certDERBlock, certPEMBlock = pem.Decode(certPEMBlock)

		// no more pem blocks
		if certDERBlock == nil {
			break
		}

		// check the found block
		if certDERBlock.Type == "CERTIFICATE" {
			cert.Certificate = append(cert.Certificate, certDERBlock.Bytes)
		} else {
			// found a PEM block that is NOT a certificate; error (should not happen due
			// to above BEGIN CERTIFICATE check, but just in case)
			return "", errors.New("certificate has at least one non- CERTIFICATE pem block")
		}
	}

	// ensure at least one valid pem certificate block was decoded
	if len(cert.Certificate) < 1 {
		return "", errors.New("no valid CERTIFICATE pem blocks")
	}

	// get topmost cert's Issuer CN
	derTopCert := cert.Certificate[len(cert.Certificate)-1]
	topCert, err := x509.ParseCertificate(derTopCert)
	if err != nil {
		return "", errors.New("failed to parse top cert in chain")
	}

	return topCert.Issuer.CommonName, nil
}

// DownloadCertificate uses POST-as-GET to download a valid certificate from the specified
// url. If this first download passes basic sanity checks, it becomes the default cert/chain
// that is returned by this function. If the sanity check fails, alternate chain options are
// tried until one passes the sanity check. The first to pass the sanity check becomes the
// default cert/chain returned.
// If a preferredChain is specified, prefer the chain whose topmost certificate was issued
// from this Subject Common Name (assuming the basic sanity checks pass). If no match or the
// sanity check fails, the default chain is returned instead.
func (service *Service) DownloadCertificate(certificateUrl string, accountKey AccountKey, preferredChain *string) (pemChain string, err error) {
	defaultChain := ""

	// POST-as-GET
	bodyBytes, defaultHeaders, err := service.postAsGet(certificateUrl, accountKey)
	if err != nil {
		return "", err
	}

	// validate initial response
	rootCN, err := validateCertificate(bodyBytes, defaultHeaders)
	// if default chain didn't validate, log issue and continue to alts
	if err != nil {
		service.logger.Warnf("acme: %s default cert chain failed validation (see: rfc8555 s 11.4) (%s); will try others if available", certificateUrl, err)
		// don't return, instead try any alt chains first
	} else {
		// return now if no preferred chain specified, OR if this chain is the preferred chain
		if preferredChain == nil || strings.EqualFold(*preferredChain, rootCN) {
			return string(bodyBytes), nil
		}

		// validated but not the preferred chain we're seeking -- set this as default chain
		defaultChain = string(bodyBytes)
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
		bodyBytes, headers, err := service.postAsGet(altChainURL.String(), accountKey)
		if err != nil {
			service.logger.Warnf("acme: failed to fetch alt chain %s (%s); will try other available chains", altChainURL.String(), err)
			// don't return, continue to next alt to keep trying
			continue
		}

		// validate alt chain
		rootCN, err = validateCertificate(bodyBytes, headers)
		// if default chain didn't validate, log issue
		if err != nil {
			service.logger.Warnf("acme: %s alt cert chain failed validation (see: rfc8555 s 11.4) (%s); will try others if available", altChainURL.String(), err)
			// don't return, continue to next alt to keep trying
			continue
		}

		// this chain is now confirmed valid

		// return this chain if no preferred chain specified, OR if this chain is the preferred chain
		if preferredChain == nil || strings.EqualFold(*preferredChain, rootCN) {
			return string(bodyBytes), nil
		}

		// set defaultChain if there isn't one yet
		if defaultChain == "" {
			defaultChain = string(bodyBytes)
		}

		// preferred chain is set and this chain did not match, continue loop to keep trying alts
	}

	// made it through all of the alt chains without success (either no valid chain at all and/or no chain match preferred chain

	// check if defaultChain was set by any valid chain; if not, error as no valid chains
	if defaultChain == "" {
		return "", fmt.Errorf("acme: no valid cert chains found for %s", certificateUrl)
	}

	// at least one valid chain was found (even though preferred didn't match), return that one
	service.logger.Warnf("acme: went through all alt chains of %s without preferred chain match, returning default chain", certificateUrl)
	return defaultChain, nil
}
