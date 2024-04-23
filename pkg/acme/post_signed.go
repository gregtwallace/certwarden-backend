package acme

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"go.uber.org/zap/zapcore"
)

// acmeSignedMessage is the ACME signed message payload
type acmeSignedMessage struct {
	Payload         string `json:"payload"`
	ProtectedHeader string `json:"protected"`
	Signature       string `json:"signature"`
}

// dataToSign assembles the byte slice that should be signed by
// a signing method to create the signature field
func (asm *acmeSignedMessage) dataToSign() []byte {
	return []byte(strings.Join([]string{asm.ProtectedHeader, asm.Payload}, "."))
}

// ProtectedHeader piece of the ACME payload
type protectedHeader struct {
	Algorithm  string      `json:"alg"`
	JsonWebKey *jsonWebKey `json:"jwk,omitempty"`
	KeyId      string      `json:"kid,omitempty"`
	Nonce      string      `json:"nonce,omitempty"`
	Url        string      `json:"url"`
}

// postToUrlSigned posts the payload to the specified url, using the specified AccountKeyInfo
// and returns the response body (data / bytes) and headers from ACME
func (service *Service) postToUrlSigned(payload any, url string, accountKey AccountKey) (body []byte, headers http.Header, err error) {
	// message is what will ultimately be posted to ACME
	var message acmeSignedMessage

	// build most of the header (pieces that won't change in the loop)
	var header protectedHeader

	// alg
	header.Algorithm, err = accountKey.signingAlg()
	if err != nil {
		return nil, nil, err
	}

	// key or kid
	// use kid if available, otherwise use jsonWebKey
	if accountKey.Kid != "" {
		header.JsonWebKey = nil
		header.KeyId = accountKey.Kid
	} else {
		header.JsonWebKey, err = accountKey.jwk()
		if err != nil {
			return nil, nil, err
		}
		header.KeyId = ""
	}

	// nonce
	header.Nonce, err = service.nonceManager.Nonce()
	if err != nil {
		return nil, nil, err
	}

	// url
	header.Url = url

	// Debugging
	//service.logger.Debugf("unencoded acme header: %s", header)

	// header (end)

	// debug log payload content & destination
	if service.logger.Level() == zapcore.DebugLevel {
		prettyPayload, prettyErr := json.MarshalIndent(payload, "", "\t")
		if prettyErr != nil {
			service.logger.Debugf("sending acme signed post to: %s ; unencoded payload: %s", url, payload)
		} else {
			service.logger.Debugf("sending acme signed post to: %s ; unencoded payload: %s", url, string(prettyPayload))
		}
	}

	// set payload - won't change in the loop (if payload is empty, don't encode it)
	if payload == "" {
		message.Payload = ""
	} else {
		message.Payload, err = encodeJson(payload)
		if err != nil {
			return nil, nil, err
		}
	}

	// post
	var response *http.Response
	var bodyBytes []byte

	// loop to retry on badNonce error, capped at 4 tries
	for i := 0; i < 4; i++ {
		// encord and insert header
		message.ProtectedHeader, err = encodeJson(header)
		if err != nil {
			return nil, nil, err
		}

		// sign
		err = message.Sign(accountKey)
		if err != nil {
			return nil, nil, err
		}

		// marshal for posting
		var messageBodyJson []byte
		messageBodyJson, err = json.Marshal(message)
		if err != nil {
			return nil, nil, err
		}

		// data to POST (debugging)
		//service.logger.Debugf("acme post signed body: %s", string(messageBodyJson))

		// post to ACME
		response, err = service.httpClient.Post(url, "application/jose+json", bytes.NewBuffer(messageBodyJson))
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
				header.Nonce = response.Header.Get("Replay-Nonce")

				// no need to sleep, remote server is working ok
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
		return nil, nil, err
	}

	// verify status code is success (catch all in case acmeError didn't decode somehow)
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, nil, fmt.Errorf("acme error: status code %d", response.StatusCode)
	}

	return bodyBytes, response.Header, nil
}

// postAsGet implements POST-as-GET as specified in rfc8555 6.3.
// Specific functions that use this will also need to be defined
func (service *Service) postAsGet(url string, accountKey AccountKey) (body []byte, headers http.Header, err error) {
	return service.postToUrlSigned("", url, accountKey)
}
