package nonces

import (
	"errors"
	"net/http"
)

// Manager is the NonceManager
type Manager struct {
	newNonceUrl *string
	nonces      *ringBuffer
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

// Nonce returns the oldest nonce from the nonce buffer.
// If the buffer cannot be read, a new nonce will be acquired by
// fetching from the newNonceUrl
func (manager *Manager) Nonce() (nonce string, err error) {
	// try to read, if error fetch new
	nonce, err = manager.nonces.read()

	// if read failed, fetch from url
	if err != nil {
		return manager.fetchNonce()
	}

	return nonce, err
}

// SaveNonce saves the nonce string to the nonces buffer.
// if the buffer is full, the oldest nonce is evicted and the new
// nonce is saved
func (manager *Manager) SaveNonce(nonce string) (err error) {
	// if nonce is empty, don't save
	if nonce == "" {
		return errors.New("cannot save empty nonce")
	}

	// lock nonces
	manager.nonces.mu.Lock()
	defer manager.nonces.mu.Unlock()

	// if full, evict (read) a nonce. Ring is FIFO so the oldest
	// nonce will be read and discarded.
	if manager.nonces.isFull {
		_, err = manager.nonces.readUnsafe()
	}

	return manager.nonces.writeUnsafe(nonce)
}
