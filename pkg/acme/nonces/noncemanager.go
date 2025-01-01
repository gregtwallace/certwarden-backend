package nonces

import (
	"certwarden-backend/pkg/datatypes/ringbuffer"
	"errors"
	"io"
	"net/http"
)

// Manager is the NonceManager
type Manager struct {
	httpClient  *http.Client
	newNonceUrl *string
	nonces      *ringbuffer.RingBuffer[string]
}

// Manager buffer size
const bufferSize = 32

// NewManager creates a new nonce manager
func NewManager(client *http.Client, nonceUrl *string) *Manager {
	manager := new(Manager)

	manager.httpClient = client
	manager.newNonceUrl = nonceUrl
	manager.nonces = ringbuffer.NewRingBuffer[string](bufferSize)

	return manager
}

// fetchNonce gets a nonce from the manager's newNonceUrl
// if fetching fails or the header does not contain a nonce,
// an error is returned
func (manager *Manager) fetchNonce() (string, error) {
	response, err := manager.httpClient.Head(*manager.newNonceUrl)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	// read entire body (to keep single tls connection open and avoid redundant cert
	// log messages) see: https://stackoverflow.com/questions/17948827/reusing-http-connections-in-go
	// for explanation
	_, _ = io.Copy(io.Discard, response.Body)

	// make sure the nonce isn't blank
	nonce := response.Header.Get("Replay-Nonce")
	if nonce == "" {
		return "", errors.New("failed to fetch nonce, no value in header")
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
		return errors.New("cannot save empty nonce")
	}

	// write new nonce and evict oldest if buffer is full
	err := manager.nonces.Write(nonce, true)
	if err != nil {
		return err
	}

	return nil
}
