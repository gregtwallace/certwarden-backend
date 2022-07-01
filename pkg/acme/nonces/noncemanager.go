package nonces

import (
	"errors"
	"net/http"
	"sync"
)

// Manager is the NonceManager
type Manager struct {
	newNonceUrl *string
	nonces      *ringBuffer
	mu          sync.Mutex
}

// Manager buffer size
const bufferSize = 32

// NewManager creates a new nonce manager
func NewManager(nonceUrl *string) *Manager {
	manager := new(Manager)

	manager.newNonceUrl = nonceUrl
	manager.nonces = newRingBuffer(bufferSize)

	return manager
}

// fetchNonce gets a nonce from the manager's newNonceUrl
// if fetching fails or the header does not contain a nonce,
// an error is returned
func (manager *Manager) fetchNonce() (nonce string, err error) {
	response, err := http.Get(*manager.newNonceUrl)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	// make sure the nonce isn't blank
	nonce = response.Header.Get("Replay-Nonce")
	if nonce == "" {
		return "", errors.New("failed to fetch nonce, no value in header")
	}

	return nonce, nil
}

// GetNonce returns the oldest nonce from the nonces buffer.
// If the buffer is empty, a new nonce will be acquired by
// fetching from the newNonceUrl
func (manager *Manager) GetNonce() (nonce string, err error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	if manager.nonces.length() == 0 {
		return manager.fetchNonce()
	}

	return manager.nonces.read()
}

// SaveNonce saves the nonce string to the nonces buffer.
// if the buffer is full, the oldest nonce is evicted and the new
// nonce is saved
func (manager *Manager) SaveNonce(nonce string) (err error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	// if nonce is empty, don't save
	if nonce == "" {
		return errors.New("cannot save empty nonce")
	}

	// if full, evict (read) the oldest nonce
	if manager.nonces.isFull {
		_, err = manager.nonces.read()
	}

	return manager.nonces.write(nonce)
}
