package directory

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
	"golang.org/x/sync/semaphore"
)

const testDirURI = "https://acme.test/directory"

const fullDirectoryJSON = `{
	"newNonce": "https://acme.test/new-nonce",
	"newAccount": "https://acme.test/new-account",
	"newOrder": "https://acme.test/new-order",
	"newAuthz": "https://acme.test/new-authz",
	"revokeCert": "https://acme.test/revoke-cert",
	"keyChange": "https://acme.test/key-change",
	"renewalInfo": "https://acme.test/renewal-info",
	"meta": {
		"termsOfService": "https://acme.test/tos",
		"website": "https://acme.test",
		"caaIdentities": ["acme.test"],
		"externalAccountRequired": true,
		"profiles": {
			"classic": "The default profile",
			"shortlived": "A short lived profile"
		}
	}
}`

// same shape as fullDirectoryJSON, but with different content
const altDirectoryJSON = `{
	"newNonce": "https://acme2.test/new-nonce",
	"newAccount": "https://acme2.test/new-account",
	"newOrder": "https://acme2.test/new-order",
	"revokeCert": "https://acme2.test/revoke-cert",
	"keyChange": "https://acme2.test/key-change",
	"meta": {
		"termsOfService": "https://acme2.test/tos"
	}
}`

// only the strictly required fields (newAuthz, keyChange, renewalInfo and
// meta contents are all omitted)
const requiredOnlyDirectoryJSON = `{
	"newNonce": "https://acme.test/new-nonce",
	"newAccount": "https://acme.test/new-account",
	"newOrder": "https://acme.test/new-order",
	"revokeCert": "https://acme.test/revoke-cert",
	"meta": {}
}`

// renewalInfo key is present, but empty
const emptyRenewalInfoDirectoryJSON = `{
	"newNonce": "https://acme.test/new-nonce",
	"newAccount": "https://acme.test/new-account",
	"newOrder": "https://acme.test/new-order",
	"revokeCert": "https://acme.test/revoke-cert",
	"renewalInfo": "",
	"meta": {}
}`

// failure backoff window set before each fetch: 2s + up to 299ms of jitter
const (
	failBackoff       = 2 * time.Second
	failBackoffJitter = 299 * time.Millisecond
)

// ---------------------------------------------------------------------------
// test helpers
// ---------------------------------------------------------------------------

type roundTripHandler func(req *http.Request) (*http.Response, error)

// fakeTransport is an http.RoundTripper that never touches the network, which
// lets tests use a valid https:// directory URI.
type fakeTransport struct {
	hits atomic.Int64

	mu      sync.Mutex
	handler roundTripHandler
}

func (f *fakeTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	f.hits.Add(1)

	// behave like a real transport and honor the request's context
	if err := req.Context().Err(); err != nil {
		return nil, err
	}

	f.mu.Lock()
	h := f.handler
	f.mu.Unlock()

	return h(req)
}

func (f *fakeTransport) setHandler(h roundTripHandler) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.handler = h
}

func respondWith(status int, body string) roundTripHandler {
	return func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: status,
			Status:     http.StatusText(status),
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(body)),
			Request:    req,
		}, nil
	}
}

func respondWithError(err error) roundTripHandler {
	return func(*http.Request) (*http.Response, error) {
		return nil, err
	}
}

// respondSlowly blocks until the request's context ends, like a server that
// never answers.
func respondSlowly() roundTripHandler {
	return func(req *http.Request) (*http.Response, error) {
		<-req.Context().Done()
		return nil, req.Context().Err()
	}
}

// errBody is a response body that fails while being read
type errBody struct{}

func (errBody) Read([]byte) (int, error) { return 0, errors.New("read boom") }
func (errBody) Close() error             { return nil }

func respondWithBrokenBody() roundTripHandler {
	return func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{},
			Body:       errBody{},
			Request:    req,
		}, nil
	}
}

// newTestMgr returns a DirectoryMgr wired to a fake transport and an observed
// logger.
//
// It builds the struct directly instead of calling New so that New's
// background cache priming doesn't run; tests then control exactly when
// fetches happen. New itself is covered by the TestNew_* tests.
func newTestMgr(t *testing.T, dirURI string, h roundTripHandler) (*DirectoryMgr, *fakeTransport, *observer.ObservedLogs) {
	t.Helper()

	core, logs := observer.New(zapcore.DebugLevel)
	tr := &fakeTransport{handler: h}
	dm := &DirectoryMgr{
		logger:     zap.New(core).Sugar(),
		httpClient: &http.Client{Transport: tr},
		dirURI:     dirURI,
		semHttp:    semaphore.NewWeighted(1),
	}

	return dm, tr, logs
}

// expire makes the next updateDirectory call due. Only safe while no other
// goroutine is using the manager.
func expire(dm *DirectoryMgr) {
	dm.nextFetchAt = time.Now().Add(-time.Second)
}

func errorLogs(logs *observer.ObservedLogs) []observer.LoggedEntry {
	return logs.FilterLevelExact(zapcore.ErrorLevel).All()
}

func assertNextFetchWithin(t *testing.T, got, before, after time.Time, base, maxJitter time.Duration) {
	t.Helper()

	assert.False(t, got.Before(before.Add(base)),
		"next fetch %s is earlier than allowed (%s after the update started)", got, base)
	assert.False(t, got.After(after.Add(base+maxJitter)),
		"next fetch %s is later than allowed (%s + %s jitter after the update ended)", got, base, maxJitter)
}

func assertSemaphoreFree(t *testing.T, dm *DirectoryMgr) {
	t.Helper()

	require.True(t, dm.semHttp.TryAcquire(1), "semaphore must be released when updateDirectory returns")
	dm.semHttp.Release(1)
}

type getter struct {
	name string
	call func(dm *DirectoryMgr) any
}

// getters that take a context
func ctxGetters() []getter {
	ctx := context.Background()
	return []getter{
		{"NewNonceURL", func(dm *DirectoryMgr) any { return dm.NewNonceURL(ctx) }},
		{"NewAccountURL", func(dm *DirectoryMgr) any { return dm.NewAccountURL(ctx) }},
		{"NewOrderURL", func(dm *DirectoryMgr) any { return dm.NewOrderURL(ctx) }},
		{"RevokeCertURL", func(dm *DirectoryMgr) any { return dm.RevokeCertURL(ctx) }},
		{"KeyChangeURL", func(dm *DirectoryMgr) any { return dm.KeyChangeURL(ctx) }},
		{"RenewalInfoURL", func(dm *DirectoryMgr) any { return dm.RenewalInfoURL(ctx) }},
	}
}

// getters that end in Cached and cap their own wait
func cachedGetters() []getter {
	return []getter{
		{"TermsOfServiceURLCached", func(dm *DirectoryMgr) any { return dm.TermsOfServiceURLCached() }},
		{"RequiresEABCached", func(dm *DirectoryMgr) any { return dm.RequiresEABCached() }},
		{"RawResponseCached", func(dm *DirectoryMgr) any { return dm.RawResponseCached() }},
		{"ProfilesCached", func(dm *DirectoryMgr) any { return dm.ProfilesCached() }},
		{"ProfileIsPublishedCached", func(dm *DirectoryMgr) any { return dm.ProfileIsPublishedCached("classic") }},
		{"SupportsRenewalInfoCached", func(dm *DirectoryMgr) any { return dm.SupportsRenewalInfoCached() }},
	}
}

func allGetters() []getter {
	return append(ctxGetters(), cachedGetters()...)
}

// ---------------------------------------------------------------------------
// New
// ---------------------------------------------------------------------------

func TestNew_PrimesCache(t *testing.T) {
	core, logs := observer.New(zapcore.DebugLevel)
	tr := &fakeTransport{handler: respondWith(http.StatusOK, fullDirectoryJSON)}
	client := &http.Client{Transport: tr}

	dm := New(zap.New(core).Sugar(), client, testDirURI)

	require.NotNil(t, dm)
	assert.Same(t, client, dm.httpClient)
	assert.Equal(t, testDirURI, dm.dirURI)
	assert.NotNil(t, dm.logger)
	require.NotNil(t, dm.semHttp)

	// New fetches the directory in the background without being asked to
	require.Eventually(t, func() bool { return tr.hits.Load() >= 1 },
		5*time.Second, 5*time.Millisecond, "New must prime the cache")

	// acquiring the semaphore waits for the background update to finish
	require.NoError(t, dm.semHttp.Acquire(context.Background(), 1))
	defer dm.semHttp.Release(1)

	assert.Equal(t, int64(1), tr.hits.Load())
	assert.Empty(t, errorLogs(logs))

	dm.muData.RLock()
	d := dm.directoryObj
	dm.muData.RUnlock()

	assert.Equal(t, "https://acme.test/new-nonce", d.NewNonce)
	assert.Equal(t, fullDirectoryJSON, string(d.raw))
	assert.True(t, dm.nextFetchAt.After(time.Now().Add(dirUpdateInterval-time.Minute)),
		"a successful priming fetch must schedule the next fetch ~15 minutes out")
}

func TestNew_PrimingWithInvalidURI(t *testing.T) {
	core, logs := observer.New(zapcore.DebugLevel)
	tr := &fakeTransport{handler: respondWith(http.StatusOK, fullDirectoryJSON)}

	dm := New(zap.New(core).Sugar(), &http.Client{Transport: tr}, "http://acme.test/directory")

	require.Eventually(t, func() bool {
		return logs.FilterMessageSnippet("dirURI is invalid").Len() == 1
	}, 5*time.Second, 5*time.Millisecond)

	require.NoError(t, dm.semHttp.Acquire(context.Background(), 1))
	defer dm.semHttp.Release(1)

	assert.Equal(t, int64(0), tr.hits.Load(), "a non-https dirURI must never be fetched")
	assert.Equal(t, directory{}, dm.directoryObj)
}
