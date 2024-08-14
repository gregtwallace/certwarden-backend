package safecert

import (
	"certwarden-backend/pkg/httpclient"
	"context"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"io"
	"math/rand/v2"
	"net/http"
	"time"

	"github.com/cenkalti/backoff/v4"
	"golang.org/x/crypto/ocsp"
)

var (
	errOCSPStaplingNotAvailable = errors.New("safecert: ocsp stapling not available for leaf certificate")
	errOCSPValidTooLong         = errors.New("safecert: received ocsp response that expires after leaf (which is invalid)")
	errOCSPStatusNotGood        = errors.New("safecert: ocsp response was not Good (0)")
)

// getOCSPResponse fetches the OCSP response from the OCSP server for the specified
// leaf certificate
func getOCSPResponse(leafCert, issuerCert *x509.Certificate, httpClient *httpclient.Client) (*ocsp.Response, error) {
	// if no leaf or issuer, ocsp isn't supported
	if leafCert == nil || issuerCert == nil {
		return nil, errOCSPStaplingNotAvailable
	}

	// make sure there is at least one OCSPServer for leaf
	if len(leafCert.OCSPServer) <= 0 {
		return nil, errOCSPStaplingNotAvailable
	}

	// make request
	ocspReq, err := ocsp.CreateRequest(leafCert, issuerCert, nil)
	if err != nil {
		// fail, no stapling
		return nil, err
	}
	ocspReqBase64 := base64.StdEncoding.EncodeToString(ocspReq)

	// headers for GET
	headers := http.Header{}
	headers.Set("Content-Language", "application/ocsp-request")
	headers.Set("Accept", "application/ocsp-response")

	// randomly select starting ocsp server from list
	serverIndex := rand.IntN(len(leafCert.OCSPServer))

	// fetch response (try each server until valid response, or run out of servers)
	var ocspResp *ocsp.Response
	for i := 0; i < len(leafCert.IssuingCertificateURL); i++ {
		reqURL := leafCert.OCSPServer[(serverIndex+i)%len(leafCert.OCSPServer)] + "/" + ocspReqBase64

		var resp *http.Response
		resp, err = httpClient.GetWithHeader(reqURL, headers)
		if err != nil {
			// this loop iteration failed
			continue
		}
		defer resp.Body.Close()

		// read and parse
		var respBody []byte
		respBody, err = io.ReadAll(resp.Body)
		if err != nil {
			// this loop iteration failed
			continue
		}

		ocspResp, err = ocsp.ParseResponse(respBody, issuerCert)
		if err != nil {
			// this loop iteration failed
			continue
		}

		// if OCSP Response is valid longer than the leaf cert, it is no bueno
		// and should be discarded
		if ocspResp.NextUpdate.After(leafCert.NotAfter.Truncate(time.Second).Add(1 * time.Second)) {
			// this loop iteration failed
			err = errOCSPValidTooLong
			continue
		}

		// only return the response if it is good
		if ocspResp.Status != ocsp.Good {
			// this loop iteration failed
			err = errOCSPStatusNotGood
			continue
		}
	}

	// check last err from loop
	if err != nil {
		return nil, err
	}

	return ocspResp, nil
}

// startOCSPManagement starts a go routine that will manage OCSP responses
// and staple them to the cert; if the current cert does not support this
// then the OCSP mgmt thread isn't started
// This function MUST ONLY be called from a WRITE LOCKED state!
func (sc *SafeCert) startOCSPManagement() {
	// stop any previous management task
	if sc.stopOCSPMgmt != nil {
		sc.stopOCSPMgmt()
	}

	// only do something if the need certs exist and at least one OSCP server
	// if no leaf or issuer, ocsp isn't supported
	if sc.leafCert == nil || sc.issuerCert == nil {
		return
	}

	// make sure there is at least one OCSPServer for leaf
	if len(sc.leafCert.OCSPServer) <= 0 {
		return
	}

	// backoff object for retry of failed OCSP response
	bo := backoff.NewExponentialBackOff()
	bo.InitialInterval = 10 * time.Second
	bo.RandomizationFactor = 0.25
	bo.Multiplier = 3
	bo.MaxInterval = 6 * 60 * time.Minute
	bo.MaxElapsedTime = 0 // never stop trying

	// go routine to update the stapled OCSP response
	stopCtx, cancel := context.WithCancel(sc.shutdownCtx)
	sc.stopOCSPMgmt = cancel

	sc.shutdownWg.Add(1)
	go func() {
		// defer wg Done
		defer sc.shutdownWg.Done()

		for {
			// need write lock for OCSP updates
			// Note: This will block until the original called of startOCSPManagement()
			// releases its write lock.
			sc.Lock()

			// get OCSP response from OCSP server
			var nextRunTime time.Time
			ocspResp, err := getOCSPResponse(sc.leafCert, sc.issuerCert, sc.httpClient)
			if err != nil {
				// remove old OCSP if it is expired (aka Now is past non-zero NextUpdate)
				if sc.ocspResp != nil && !sc.ocspResp.NextUpdate.IsZero() && time.Now().After(sc.ocspResp.NextUpdate) {
					sc.ocspResp = nil
				}

				// use exponential backoff to try again
				nextRunTime = time.Now().Add(bo.NextBackOff())

				// if the existing stapled OCSP will expire before next bo retry, retry on the
				// existing stapled expiration instead
				if sc.ocspResp != nil && sc.ocspResp.NextUpdate.Before(nextRunTime) {
					nextRunTime = sc.ocspResp.NextUpdate
				}

			} else {
				// Good ocspResp
				// set safe cert's OCSP
				sc.ocspResp = ocspResp

				// calculate next OCSP check
				// see: https://cabforum.org/uploads/CA-Browser-Forum-TLS-BRs-v2.0.2.pdf
				if ocspResp.NextUpdate.IsZero() {
					// if no next update time, new info always available -- just use 1 day
					nextRunTime = time.Now().Add(24 * time.Hour)
				} else {
					// there is a next update; calculate validity and schedule next run accordingly
					validityDuration := ocspResp.NextUpdate.Sub(ocspResp.ThisUpdate)

					// https://cabforum.org/uploads/CA-Browser-Forum-TLS-BRs-v2.0.2.pdf s4.9.10, bullet 3
					if validityDuration < 16*time.Hour {
						// less than 16 hours of validity, CA must update prior to half-way point, so try
						// at half way
						nextRunTime = ocspResp.ThisUpdate.Add(validityDuration / 2)
					} else {
						// bullet 4 (>= 16 hour validity)
						// CA SHALL update at least 8 hours prior to NextUpdate
						// BUT no later than 4 days after ThisUpdate
						nextMinus8Hour := ocspResp.NextUpdate.Add(-8 * time.Hour) // note: add negative duration
						thisPlus4Day := ocspResp.ThisUpdate.Add(4 * 24 * time.Hour)

						// so do whichever of these is sooner
						if nextMinus8Hour.Before(thisPlus4Day) {
							nextRunTime = nextMinus8Hour
						} else {
							nextRunTime = thisPlus4Day
						}
					}
				}

				// if after all of that, somehow the nextRunTime is before now, use exponential
				// backoff for the delay
				if nextRunTime.Before(time.Now()) {
					nextRunTime = time.Now().Add(bo.NextBackOff())
				} else {
					// if time is sane and in the future, reset backoff
					bo.Reset()
				}
			}

			// make timer to check for next OCSP update
			nextTimer := time.NewTimer(time.Until(nextRunTime.Truncate(time.Second).Add(time.Second)))

			// OCSP update done
			sc.Unlock()

			select {
			case <-stopCtx.Done():
				// context canceled (stop OCSP task for this cert)

				// ensure timer releases resources
				if !nextTimer.Stop() {
					<-nextTimer.C
				}

				// end routine
				return

			case <-nextTimer.C:
				// run next OCSP update
			}
		}
	}()
}
