package authorizations

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrCacheFailedToRemove  = errors.New("auth working cannot remove non existent url")
	ErrCacheAuthDoesntExist = errors.New("cache does not contain specified auth url")
)

// an auth cache item with any current auth error as well as context
// that can be canceled to cancel the outstanding ttl expirer
type auth struct {
	status       string
	err          error
	cancelExpire context.CancelFunc
}

// cache stores auth urls and any corresponding error from the last
// attempt to complete the auth
type cache struct {
	auths map[string]*auth // map[url]*auth
	ttl   time.Duration
	mu    sync.RWMutex
}

// newCache creates a cache and sets the ttl (which applies to each auth in the cache)
func newCache() *cache {
	ttlSeconds := 30

	a := make(map[string]*auth)
	t := time.Duration(ttlSeconds) * time.Second

	return &cache{
		auths: a,
		ttl:   t,
	}
}

// add adds the specified auth URL and any current error
func (cache *cache) add(authUrl string, authStatus string, authErr error) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	// check if exists in cache, if so, cancel the old expirer
	oldAuth, exists := cache.auths[authUrl]
	if exists {
		oldAuth.cancelExpire()
	}

	// context for new auth's expirer
	ctx, cancel := context.WithCancel(context.Background())

	// make new auth for cache
	newAuth := new(auth)
	newAuth = &auth{
		status:       authStatus,
		err:          authErr,
		cancelExpire: cancel,
	}

	// spawn expirer
	cache.newExpirer(authUrl, ctx)

	// add the auth to cache (overwrite if already exists)
	cache.auths[authUrl] = newAuth
}

// newExpirer starts a go routine that will expire the specified cache item
// (authUrl) after the cache's ttl. Expirer can be cancel via its context,
// which is done in the event the same auth is added to the cache before the
// prior auth has expired.
func (cache *cache) newExpirer(authUrl string, ctx context.Context) {
	go func() {
		select {
		// expire after the ttl elapses
		case <-time.After(cache.ttl):
			cache.auths[authUrl].cancelExpire() // done with the context
			_ = cache.remove(authUrl)

		// context was canceled
		case <-ctx.Done():
			// break
		}
	}()
}

// remove removes the specified authUrl from the cache
func (cache *cache) remove(authUrl string) (err error) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	// if url was not found, error
	_, exists := cache.auths[authUrl]
	if !exists {
		return ErrCacheFailedToRemove
	}

	// remove the authUrl
	delete(cache.auths, authUrl)

	return nil
}

// read cache auth and return the err that was received when the auth was
// cached
func (cache *cache) read(authUrl string) (status string, err error) {
	cache.mu.RLock()
	defer cache.mu.RUnlock()

	// if url was not found, error
	cachedAuth, exists := cache.auths[authUrl]
	if !exists {
		return "", ErrCacheAuthDoesntExist
	}
	return cachedAuth.status, cachedAuth.err
}
