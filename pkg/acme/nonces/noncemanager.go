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
	manager := new(Manager)

	manager.httpClient = client
	manager.shutdownContext = shutdownCtx

	manager.newNonceUrl = nonceUrl
	manager.nonces = ringbuffer.NewRingBuffer[string](bufferSize)

	return manager
}

// fetchNonce gets a nonce from the manager's newNonceUrl
// if fetching fails or the header does not contain a nonce,
// an error is returned
func (manager *Manager) fetchNonce() (string, error) {
	// if this fails for some reason, give a sane amount of retries
	// dont bother getting fancy with exponential backoff, just fail if not resolved relatively quickly
	maxRetries := 5
	defaultWait := 1 * time.Minute
	nonce := ""

	for range maxRetries {
		response, err := manager.httpClient.Head(*manager.newNonceUrl)
		if err != nil {
			return "", err
		}
		defer response.Body.Close()

		// read entire body (to keep single tls connection open and avoid redundant cert
		// log messages) see: https://stackoverflow.com/questions/17948827/reusing-http-connections-in-go
		// for explanation
		_, _ = io.Copy(io.Discard, response.Body)

		// verify we got a nonce and break if so
		nonce = response.Header.Get("Replay-Nonce")
		if nonce != "" {
			break
		}

		// no valid nonce value, delay and retry (add semi-random amount of waiting to default)
		wait := defaultWait + time.Duration(randomness.GenerateInsecureInt(30))

		// check for and use Retry-After if the server sent one
		retryAfter := response.Header.Get("Retry-After")
		if retryAfter != "" {
			// check if header was in seconds and ensure > 0
			secs, err := strconv.Atoi(retryAfter)
			if err == nil && secs > 0 {
				wait = time.Duration(secs) * time.Second
			} else {
				// wasn't in seconds, try to parse date and ensure > 0
				t, err := http.ParseTime(retryAfter)
				if err == nil {
					tempWait := time.Until(t)
					if tempWait > 0 {
						wait = tempWait
					}
				}
			}
		}
		// dont log or fail if header was missing or didn't parse, just use the default wait

		// do the waiting
		select {
		case <-manager.shutdownContext.Done():
			// abort nonce fetching due to shutdown
			return "", errors.New("nonce manager: failed to fetch nonce due to shutdown")

		case <-time.After(wait):
			// do the waiting then proceed to next
		}
	}

	if nonce == "" {
		return "", errors.New("nonce manager: failed to fetch nonce from acme server (exhausted retries)")
	}

	return nonce, nil
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
