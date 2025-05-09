package acme

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap/zapcore"
)

// get does an unauthenticated GET request to an ACME endpoint
func (service *Service) get(url string) (bodyBytes []byte, _ http.Header, _ error) {
	// do GET
	resp, err := service.httpClient.Get(url)
	if err != nil {
		return nil, nil, fmt.Errorf("acme: get %s failed (%v)", url, err)
	}
	defer resp.Body.Close()

	// read body of response
	bodyBytes, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("acme: get %s failed to read body (%v)", url, err)
	}

	// ACME response body (debugging)
	// indent (if possible) before debug logging
	if service.logger.Level() == zapcore.DebugLevel {
		var prettyBytes bytes.Buffer
		prettyErr := json.Indent(&prettyBytes, bodyBytes, "", "\t")
		if prettyErr != nil {
			service.logger.Debugf("acme: get %s response code: %d ; body: %s", url, resp.StatusCode, string(bodyBytes))
		} else {
			service.logger.Debugf("acme: get %s response code: %d ; body: %s", url, resp.StatusCode, prettyBytes.String())
		}
	}

	// try to decode AcmeError
	acmeError := unmarshalErrorResponse(bodyBytes)
	if acmeError != nil {
		return nil, nil, acmeError
	}

	// verify status code is success (catch all in case acmeError wasn't present)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return bodyBytes, resp.Header, fmt.Errorf("acme: get %s error: status code %d", url, resp.StatusCode)
	}

	return bodyBytes, resp.Header, nil
}

// postAsGet implements POST-as-GET as specified in rfc8555 6.3.
// Specific functions that use this will also need to be defined
func (service *Service) PostAsGet(url string, accountKey AccountKey) (body []byte, headers http.Header, err error) {
	return service.postToUrlSigned("", url, accountKey)
}
