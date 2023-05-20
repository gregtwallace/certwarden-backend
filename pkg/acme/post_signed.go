package acme

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
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

	service.logger.Debugf("unencoded acme header: %s", header)
	// header (end)

	// payload won't change in the loop
	// if payload is empty, don't encode it
	if payload == "" {
		message.Payload = ""
	} else {
		message.Payload, err = encodeJson(payload)
		if err != nil {
			return nil, nil, err
		}
	}

	// post
	var messageJson []byte
	var response *http.Response
	var bodyBytes []byte
	var acmeError Error

	// loop to allow retry on badNonce error, capped at 3 tries
	for i, done := 0, false; i < 3 && !done; i++ {
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
		messageJson, err = json.Marshal(message)
		if err != nil {
			return nil, nil, err
		}

		// ACME to post (debugging)
		service.logger.Debugf(string(messageJson))

		// post to ACME
		response, err = service.httpClient.Post(url, "application/jose+json", bytes.NewBuffer(messageJson))
		if err != nil {
			return nil, nil, err
		}
		defer response.Body.Close() // TODO: do something with this to avoid leaving stuff hanging during loop?

		service.logger.Debugf("acme response status code: %d", response.StatusCode)

		// read body of response
		bodyBytes, err = io.ReadAll(response.Body)
		if err != nil {
			return nil, nil, err
		}

		// ACME response body (debugging)
		service.logger.Debugf(string(bodyBytes))

		// check if the response was an AcmeError
		acmeError, err = unmarshalErrorResponse(bodyBytes)

		// if nonce error, update nonce with new scoped nonce before retry loop continues
		if acmeError.Type == "urn:ietf:params:acme:error:badNonce" {
			header.Nonce = response.Header.Get("Replay-Nonce")
		} else {
			// if anything other than nonce error:
			// save nonce in manager
			nonceErr := service.nonceManager.SaveNonce(response.Header.Get("Replay-Nonce"))
			if nonceErr != nil {
				// no need to error out of routine, just log the save failure
				service.logger.Error(nonceErr)
			}
			// loop ends for any result other than badNonce
			done = true
		}
	}

	// re: acmeError decode
	// if it didn't error, that means an error response WAS decoded
	if err == nil {
		return nil, nil, acmeError
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
