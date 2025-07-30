package acme

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap/zapcore"
)

// postToUrlSigned posts the payload to the specified url, using the specified AccountKeyInfo
// and returns the response body (data / bytes) and headers from ACME
func (service *Service) postToUrlSigned(payload any, url string, accountKey AccountKey) (bodyBytes []byte, headers http.Header, err error) {
	nonce, err := service.nonceManager.Nonce()
	if err != nil {
		return nil, nil, err
	}

	// make acme msg
	msg, err := makeAcmeSignedMessage(payload, nonce, url, accountKey)
	if err != nil {
		return nil, nil, fmt.Errorf("acme: failed to make signed post message (%s)", err)
	}

	// post
	var response *http.Response

	// loop to retry on badNonce error, capped at 4 tries
	for range 4 {
		// debug log payload content & destination
		if service.logger.Level() == zapcore.DebugLevel {
			// VERY VERBOSE, includes Header & Signature, in addition to Payload

			// prettyMsg, prettyErr := json.MarshalIndent(msg, "", "\t")
			// if prettyErr != nil {
			// 	service.logger.Debugf("sending acme signed post to: %s ; unencoded msg: %s", url, msg)
			// } else {
			// 	service.logger.Debugf("sending acme signed post to: %s ; unencoded msg: %s", url, string(prettyMsg))
			// }

			// just the payload
			prettyPayload, prettyErr := json.MarshalIndent(msg.Payload, "", "\t")
			if prettyErr != nil {
				service.logger.Debugf("sending acme signed post (using kid: %s) to: %s ; unencoded payload: %s", accountKey.Kid, url, msg.Payload)
			} else {
				service.logger.Debugf("sending acme signed post (using kid: %s) to: %s ; unencoded payload: %s", accountKey.Kid, url, string(prettyPayload))
			}
		}

		// post to ACME
		response, err = service.httpClient.Post(url, "application/jose+json", msg.SignedHTTPBody())
		if err != nil {
			return nil, nil, err
		}
		defer response.Body.Close()

		// read body of response
		bodyBytes, err = io.ReadAll(response.Body)
		if err != nil {
			// if body read fails, try post again
			continue
		}

		// ACME response body (debugging)
		// indent (if possible) before debug logging
		if service.logger.Level() == zapcore.DebugLevel {
			var prettyBytes bytes.Buffer
			prettyErr := json.Indent(&prettyBytes, bodyBytes, "", "\t")
			if prettyErr != nil {
				service.logger.Debugf("acme signed post response code: %d ; body: %s", response.StatusCode, string(bodyBytes))
			} else {
				service.logger.Debugf("acme signed post response code: %d ; body: %s", response.StatusCode, prettyBytes.String())
			}
		}

		// try to decode AcmeError
		acmeError := unmarshalErrorResponse(bodyBytes)
		if acmeError != nil {
			// set err to check after loop ends
			err = acmeError

			// if acme error and it is specifically bad nonce, set header nonce and continue
			// to next loop iteration
			if acmeError.Type == "urn:ietf:params:acme:error:badNonce" {
				// no alternative nonce as server MUST return a valid nonce with the badNonce error;
				// RFC8555 s6.5: "An error response with the "badNonce" error type MUST include a
				// Replay-Nonce header field with a fresh nonce that the server will accept in a
				// retry of the original query"
				nextNonce := response.Header.Get("Replay-Nonce")

				// BANDAID for non-compliant ACME servers
				// if ACME server doesn't comply with spec (i.e. nonce header was empty); get a new
				// nonce from nonce manager
				if nextNonce == "" {
					service.logger.Warn("acme signed post: err badNonce but acme server did not provide new nonce in error response (server violates the spec; report it to the server dev)")

					var mgrErr error
					nextNonce, mgrErr = service.nonceManager.Nonce()
					if mgrErr != nil {
						// acme server didn't give proper nonce and getting one from nonce manager also failed;
						// break and return original acmeError
						service.logger.Errorf("acme signed post: failed to get new nonce after badNonce error (%s)", mgrErr)
						break
					}
				}
				// BANDAID - END

				// update msg & try again
				msg.setNonceAndSign(nextNonce, accountKey)
				continue
			}
		}

		// not bad nonce error, loop is done
		break
	}

	// save response nonce in manager
	if response != nil && response.Header != nil {
		nonceErr := service.nonceManager.SaveNonce(response.Header.Get("Replay-Nonce"))
		if nonceErr != nil {
			// no need to error out of routine, just log the save failure
			service.logger.Errorf("failed to save response replay nonce (%s)", nonceErr)
		}
	}

	// if err from loop, return
	if err != nil {
		return bodyBytes, response.Header, err
	}

	// verify status code is success (catch all in case acmeError wasn't present)
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return bodyBytes, response.Header, fmt.Errorf("acme error: status code %d", response.StatusCode)
	}

	return bodyBytes, response.Header, nil
}
