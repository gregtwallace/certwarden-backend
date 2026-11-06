package noncemanager

import (
	"certwarden-backend/pkg/randomness"
	"certwarden-backend/pkg/validation"
	"context"
	"io"
	"net/http"
	"time"
)

// default wait time til next try to get a nonce
func defaultTilNext() time.Duration {
	return (1 * time.Second) + (time.Duration(randomness.GenerateInsecureInt(500)) * time.Millisecond)
}

// fetchNonceOnce fetches a nonce from the URL using the http client. If the request fails for some reason, it
// returns a duration for when to try again.
func fetchNonceOnce(ctx context.Context, cli *http.Client, nonceURL *string) (nonce string, nextTry time.Duration) {
	if nonceURL == nil || *nonceURL == "" {
		return "", defaultTilNext()
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, *nonceURL, http.NoBody)
	if err != nil {
		return "", defaultTilNext()
	}

	response, err := cli.Do(req)
	if err != nil {
		// request failed, return default wait
		return "", defaultTilNext()
	}
	defer response.Body.Close()

	// read entire body (to keep single tls connection open and avoid redundant cert
	// log messages) see: https://stackoverflow.com/questions/17948827/reusing-http-connections-in-go
	// for explanation
	//nolint:errcheck // don't care about errors since we're discarding
	io.Copy(io.Discard, response.Body)

	// success!
	nonce = response.Header.Get("Replay-Nonce")
	if nonce != "" {
		return nonce, 0
	}

	// try to get and parse Retry-After value
	retryAt, err := validation.ParseRetryAfter(response.Header.Get("Retry-After"))
	if err != nil {
		// invalid Retry-After header, use default
		return "", defaultTilNext()
	}

	// return retryAt value
	return "", time.Until(retryAt)
}
