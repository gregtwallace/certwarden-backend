package noncemanager

import (
	"context"
	"errors"
	"time"
)

// maximum number of attempts to get a nonce with an http request
const nonceFetchMaxRetries = 3

var (
	errContextCanceled  = errors.Join(context.Canceled, errors.New("nonce manager: failed to fetch nonce (ctx canceled)"))
	errRetriesExhausted = errors.New("nonce manager: failed to fetch nonce from acme server (exhausted retries)")
)

// fetchNonce gets a nonce from themanager's newNonceUrl; if fetching fails or the header
// does not contain a nonce, an error is returned
func (manager *Manager) fetchNonce(ctx context.Context) (string, error) {
	// if this fails for some reason, give a sane amount of retries
	// dont bother getting fancy with exponential backoff, just fail if not resolved relatively quickly

	// ensure we dont try even once if ctx is already canceled
	if err := ctx.Err(); err != nil {
		return "", errContextCanceled
	}

	// retry loop
	for range nonceFetchMaxRetries {
		nonce, wait := fetchNonceOnce(ctx, manager.httpClient, manager.newNonceUrl)
		// success?
		if nonce != "" {
			return nonce, nil
		}

		// do the waiting
		select {
		case <-ctx.Done():
			// abort nonce fetching due to shutdown
			return "", errContextCanceled

		case <-time.After(wait):
			// do the waiting then proceed to next
		}
	}

	return "", errRetriesExhausted
}
