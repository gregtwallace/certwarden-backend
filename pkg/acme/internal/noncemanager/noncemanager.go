package noncemanager

import (
	"certwarden-backend/pkg/datatypes/ringbuffer"
	"context"
	"errors"
	"net/http"
)

var errPushEmptyNonce = errors.New("nonce manager: cannot save empty nonce")

// Manager buffer size
const nonceBufferSize = 32

// Manager is the NonceManager
type Manager struct {
	httpClient  *http.Client
	newNonceUrl *string
	nonceValues *ringbuffer.Buffer
}

// New creates a new nonce manager which stores nonces and fetches new nonces if its
// buffer is empty.
func New(client *http.Client, nonceUrlPtr *string) *Manager {
	// Note: nonceUrl must be a pointer so directory updates are reflected in Manager

	return &Manager{
		httpClient:  client,
		newNonceUrl: nonceUrlPtr,
		nonceValues: ringbuffer.New(nonceBufferSize),
	}
}

// PushNonce saves the nonce string to the nonces buffer. If the buffer is full, the
// oldest nonce is evicted and the new nonce is saved.
func (manager *Manager) PushNonce(nonce string) error {
	// if nonce is empty, don't save
	if nonce == "" {
		return errPushEmptyNonce
	}

	// save new nonce (evicts oldest if buffer is full)
	manager.nonceValues.Push(nonce)

	return nil
}

// PopNonce returns the next nonce from the nonce buffer. If the buffer is empty, a new
// nonce will be acquired by fetching from the newNonceUrl
func (manager *Manager) PopNonce(ctx context.Context) (string, error) {
	// try to read from buffer
	nonce, err := manager.nonceValues.Pop()
	if err != nil {
		// if read failed, fetch from url
		return manager.fetchNonce(ctx)
	}

	return nonce, nil
}
