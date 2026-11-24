package nonces

import (
	"certwarden-backend/pkg/datatypes/ringbuffer"
	"certwarden-backend/pkg/randomness"
	"context"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"
)

// Manager buffer size
const bufferSize = 32

// Manager is the NonceManager
type Manager struct {
	httpClient      *http.Client
	shutdownContext context.Context

	newNonceUrl *string
	nonces      *ringbuffer.RingBuffer[string]
}

// NewManager creates a new nonce manager
func NewManager(client *http.Client, shutdownCtx context.Context, nonceUrl *string) *Manager {
	// Note: nonceUrl must be a pointer so directory updates are reflected in Manager

	return &Manager{
		httpClient:      client,
		shutdownContext: shutdownCtx,
		newNonceUrl:     nonceUrl,
		nonces:          ringbuffer.NewRingBuffer[string](bufferSize),
	}
}

// fetchNonce gets a nonce from the manager's newNonceUrl
// if fetching fails or the header does not contain a nonce,
// an error is returned
func (manager *Manager) fetchNonce() (string, error) {
	// if this fails for some reason, give a sane amount of retries
	// dont bother getting fancy with exponential backoff, just fail if not resolved relatively quickly
	const maxRetries = 5

	// use separate function to ensure proper resource release timing
	tryGetNonce := func() (nonc string, tilNext time.Duration) {
		// default wait time til next try
		const defaultWait = 1 * time.Second
		defaultTilNext := defaultWait + time.Duration(randomness.GenerateInsecureInt(30))

		response, err := manager.httpClient.Head(*manager.newNonceUrl)
		if err != nil {
			// request failed, return default wait
			return "", defaultTilNext
		}
		defer response.Body.Close()

		// read entire body (to keep single tls connection open and avoid redundant cert
		// log messages) see: https://stackoverflow.com/questions/17948827/reusing-http-connections-in-go
		// for explanation
		//nolint:errcheck // don't care about errors since we're discarding
		io.Copy(io.Discard, response.Body)

		// done if got nonce
		nonc = response.Header.Get("Replay-Nonce")
		if nonc != "" {
			return nonc, 0
		}

		// request did something; try to get a valid Retry-After value

		// no header? use default
		retryAfter := response.Header.Get("Retry-After")
		if retryAfter == "" {
			return "", defaultTilNext
		}

		// check if header was in seconds and ensure > 0
		secs, err := strconv.Atoi(retryAfter)
		if err == nil && secs > 0 {
			return "", time.Duration(secs) * time.Second
		}

		// seconds didnt work, try to parse date and ensure > 0
		t, err := http.ParseTime(retryAfter)
		if err == nil {
			until := time.Until(t)
			if until > 0 {
				return "", until
			}
		}

		// nothing worked, use default
		return "", defaultTilNext
	}

	// first wait is 0
	var nonce string
	var wait time.Duration

	// retry loop
	for range maxRetries {
		// do the waiting
		select {
		case <-manager.shutdownContext.Done():
			// abort nonce fetching due to shutdown
			return "", errors.New("nonce manager: failed to fetch nonce due to shutdown")

		case <-time.After(wait):
			// do the waiting then proceed to next
		}

		nonce, wait = tryGetNonce()
		// success?
		if nonce != "" {
			return nonce, nil
		}
	}

	return "", errors.New("nonce manager: failed to fetch nonce from acme server (exhausted retries)")
}

// Nonce returns the oldest nonce from the nonce buffer.
// If the buffer cannot be read, a new nonce will be acquired by
// fetching from the newNonceUrl
func (manager *Manager) Nonce() (string, error) {
	// try to read, if error fetch new
	nonce, err := manager.nonces.Read()

	// if read failed, fetch from url
	if err != nil {
		return manager.fetchNonce()
	}

	return nonce, nil
}

// SaveNonce saves the nonce string to the nonces buffer. If the
// buffer is full, the oldest nonce is evicted and the new nonce
// is saved.
func (manager *Manager) SaveNonce(nonce string) error {
	// if nonce is empty, don't save
	if nonce == "" {
		return errors.New("nonce manager: cannot save empty nonce")
	}

	// write new nonce and evict oldest if buffer is full
	err := manager.nonces.Write(nonce, true)
	if err != nil {
		return err
	}

	return nil
}
