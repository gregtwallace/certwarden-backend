package directory

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// updateDirectory: success paths
// ---------------------------------------------------------------------------

func TestUpdateDirectory_Success(t *testing.T) {
	dm, tr, logs := newTestMgr(t, testDirURI, respondWith(http.StatusOK, fullDirectoryJSON))

	before := time.Now()
	dm.updateDirectory(context.Background())
	after := time.Now()

	assert.Equal(t, int64(1), tr.hits.Load())
	assert.Empty(t, errorLogs(logs))

	d := dm.directoryObj
	assert.Equal(t, "https://acme.test/new-nonce", d.NewNonce)
	assert.Equal(t, "https://acme.test/new-account", d.NewAccount)
	assert.Equal(t, "https://acme.test/new-order", d.NewOrder)
	assert.Equal(t, "https://acme.test/new-authz", d.NewAuthz)
	assert.Equal(t, "https://acme.test/revoke-cert", d.RevokeCert)
	assert.Equal(t, "https://acme.test/key-change", d.KeyChange)
	assert.Equal(t, "https://acme.test/tos", d.Meta.TermsOfService)
	assert.Equal(t, "https://acme.test", d.Meta.Website)
	assert.Equal(t, []string{"acme.test"}, d.Meta.CAAIdentities)
	assert.True(t, d.Meta.ExternalAccountRequired)
	assert.Equal(t, map[string]string{
		"classic":    "The default profile",
		"shortlived": "A short lived profile",
	}, d.Meta.Profiles)
	require.NotNil(t, d.RenewalInfo)
	assert.Equal(t, "https://acme.test/renewal-info", *d.RenewalInfo)

	// raw response is stored byte for byte
	assert.Equal(t, fullDirectoryJSON, string(d.raw))

	// next fetch is 15 minutes out, plus 0-59 seconds of jitter
	assertNextFetchWithin(t, dm.nextFetchAt, before, after, dirUpdateInterval, 59*time.Second)

	assert.Equal(t, 1, logs.FilterMessageSnippet("directory fetched and updated").Len())
	assertSemaphoreFree(t, dm)
}

func TestUpdateDirectory_RequiredFieldsOnly(t *testing.T) {
	// newAuthz and keyChange are intentionally not required, and neither are
	// the extensions
	dm, _, logs := newTestMgr(t, testDirURI, respondWith(http.StatusOK, requiredOnlyDirectoryJSON))

	dm.updateDirectory(context.Background())

	assert.Empty(t, errorLogs(logs))
	d := dm.directoryObj
	assert.Equal(t, "https://acme.test/new-nonce", d.NewNonce)
	assert.Equal(t, "https://acme.test/new-account", d.NewAccount)
	assert.Equal(t, "https://acme.test/new-order", d.NewOrder)
	assert.Equal(t, "https://acme.test/revoke-cert", d.RevokeCert)
	assert.Empty(t, d.NewAuthz)
	assert.Empty(t, d.KeyChange)
	assert.Nil(t, d.RenewalInfo)
	assert.Nil(t, d.Meta.Profiles)
	assert.False(t, d.Meta.ExternalAccountRequired)
	assert.True(t, dm.nextFetchAt.After(time.Now().Add(dirUpdateInterval-time.Minute)),
		"a successful update must schedule the next fetch ~15 minutes out")
}

func TestUpdateDirectory_RequestShape(t *testing.T) {
	type ctxKey struct{}

	var gotReq *http.Request
	var gotCtxValue any

	dm, _, _ := newTestMgr(t, testDirURI, func(req *http.Request) (*http.Response, error) {
		gotReq = req
		gotCtxValue = req.Context().Value(ctxKey{})
		return respondWith(http.StatusOK, fullDirectoryJSON)(req)
	})

	ctx := context.WithValue(context.Background(), ctxKey{}, "from-caller")
	dm.updateDirectory(ctx)

	require.NotNil(t, gotReq)
	assert.Equal(t, http.MethodGet, gotReq.Method)
	assert.Equal(t, testDirURI, gotReq.URL.String())
	assert.Equal(t, "from-caller", gotCtxValue, "values on the caller's context must reach the request")

	deadline, ok := gotReq.Context().Deadline()
	require.True(t, ok, "the request must carry the internal timeout")
	assert.WithinDuration(t, time.Now().Add(10*time.Second), deadline, 2*time.Second)
}

// ---------------------------------------------------------------------------
// updateDirectory: scheduling / caching
// ---------------------------------------------------------------------------

func TestUpdateDirectory_NoFetchWhenNotDue(t *testing.T) {
	dm, tr, _ := newTestMgr(t, testDirURI, respondWith(http.StatusOK, fullDirectoryJSON))
	ctx := context.Background()

	dm.updateDirectory(ctx)
	require.Equal(t, int64(1), tr.hits.Load())

	for i := 0; i < 10; i++ {
		dm.updateDirectory(ctx)
	}

	assert.Equal(t, int64(1), tr.hits.Load(), "update must be a no-op until nextFetchAt has passed")
}

func TestUpdateDirectory_FetchesWhenDue(t *testing.T) {
	dm, tr, _ := newTestMgr(t, testDirURI, respondWith(http.StatusOK, fullDirectoryJSON))
	ctx := context.Background()

	dm.updateDirectory(ctx)
	require.Equal(t, int64(1), tr.hits.Load())

	expire(dm)
	dm.updateDirectory(ctx)

	assert.Equal(t, int64(2), tr.hits.Load())
	assert.True(t, dm.nextFetchAt.After(time.Now()), "next fetch must be rescheduled into the future")
}

func TestUpdateDirectory_NotDueWhenNextFetchInFuture(t *testing.T) {
	dm, tr, _ := newTestMgr(t, testDirURI, respondWith(http.StatusOK, fullDirectoryJSON))
	dm.nextFetchAt = time.Now().Add(time.Hour)

	dm.updateDirectory(context.Background())

	assert.Equal(t, int64(0), tr.hits.Load())
	assert.Empty(t, dm.directoryObj.NewNonce)
}

func TestUpdateDirectory_PicksUpChanges(t *testing.T) {
	dm, tr, logs := newTestMgr(t, testDirURI, respondWith(http.StatusOK, fullDirectoryJSON))
	ctx := context.Background()

	dm.updateDirectory(ctx)
	require.Equal(t, "https://acme.test/new-nonce", dm.directoryObj.NewNonce)

	tr.setHandler(respondWith(http.StatusOK, altDirectoryJSON))
	expire(dm)
	dm.updateDirectory(ctx)

	d := dm.directoryObj
	assert.Equal(t, "https://acme2.test/new-nonce", d.NewNonce)
	assert.Equal(t, "https://acme2.test/new-account", d.NewAccount)
	assert.Equal(t, "https://acme2.test/new-order", d.NewOrder)
	assert.Equal(t, "https://acme2.test/revoke-cert", d.RevokeCert)
	assert.Equal(t, "https://acme2.test/tos", d.Meta.TermsOfService)
	assert.Equal(t, altDirectoryJSON, string(d.raw))

	// fields that disappeared from the directory must not linger
	assert.Nil(t, d.RenewalInfo)
	assert.Nil(t, d.Meta.Profiles)
	assert.False(t, d.Meta.ExternalAccountRequired)

	assert.Equal(t, 2, logs.FilterMessageSnippet("directory fetched and updated").Len())
}

func TestUpdateDirectory_UnchangedDirectory(t *testing.T) {
	dm, _, logs := newTestMgr(t, testDirURI, respondWith(http.StatusOK, fullDirectoryJSON))
	ctx := context.Background()

	dm.updateDirectory(ctx)
	require.Equal(t, 1, logs.FilterMessageSnippet("directory fetched and updated").Len())
	first := dm.directoryObj

	expire(dm)
	dm.updateDirectory(ctx)

	assert.Equal(t, first, dm.directoryObj)
	assert.Equal(t, 1, logs.FilterMessageSnippet("directory fetched and updated").Len(),
		"an identical directory must not be logged as a change")
	assert.True(t, dm.nextFetchAt.After(time.Now().Add(dirUpdateInterval-time.Minute)),
		"next fetch must still be pushed out a full interval when nothing changed")
}

// ---------------------------------------------------------------------------
// updateDirectory: failure paths
// ---------------------------------------------------------------------------

func TestUpdateDirectory_Failures(t *testing.T) {
	tests := []struct {
		name         string
		dirURI       string
		handler      roundTripHandler
		wantFetch    bool   // whether the transport should have been hit
		wantLogMatch string // substring expected in the error log
	}{
		{
			name:         "empty dirURI",
			dirURI:       "",
			handler:      respondWith(http.StatusOK, fullDirectoryJSON),
			wantFetch:    false,
			wantLogMatch: "dirURI is invalid",
		},
		{
			name:         "http (non-https) dirURI",
			dirURI:       "http://acme.test/directory",
			handler:      respondWith(http.StatusOK, fullDirectoryJSON),
			wantFetch:    false,
			wantLogMatch: "dirURI is invalid",
		},
		{
			name:         "garbage dirURI",
			dirURI:       "not a url",
			handler:      respondWith(http.StatusOK, fullDirectoryJSON),
			wantFetch:    false,
			wantLogMatch: "dirURI is invalid",
		},
		{
			name:         "transport error",
			dirURI:       testDirURI,
			handler:      respondWithError(errors.New("connection refused")),
			wantFetch:    true,
			wantLogMatch: "failed to do http request (",
		},
		{
			name:         "body read error",
			dirURI:       testDirURI,
			handler:      respondWithBrokenBody(),
			wantFetch:    true,
			wantLogMatch: "failed to read http response (",
		},
		{
			name:         "invalid json",
			dirURI:       testDirURI,
			handler:      respondWith(http.StatusOK, `{ this is not json`),
			wantFetch:    true,
			wantLogMatch: "failed to unmarshal json",
		},
		{
			name:         "empty body",
			dirURI:       testDirURI,
			handler:      respondWith(http.StatusOK, ``),
			wantFetch:    true,
			wantLogMatch: "failed to unmarshal json",
		},
		{
			name:         "server error with non-json body",
			dirURI:       testDirURI,
			handler:      respondWith(http.StatusInternalServerError, `boom`),
			wantFetch:    true,
			wantLogMatch: "failed to unmarshal json",
		},
		{
			name:         "json of the wrong type",
			dirURI:       testDirURI,
			handler:      respondWith(http.StatusOK, `["not", "an", "object"]`),
			wantFetch:    true,
			wantLogMatch: "failed to unmarshal json",
		},
		{
			name:         "empty json object",
			dirURI:       testDirURI,
			handler:      respondWith(http.StatusOK, `{}`),
			wantFetch:    true,
			wantLogMatch: "'newNonce'",
		},
		{
			name:   "missing newNonce",
			dirURI: testDirURI,
			handler: respondWith(http.StatusOK, `{
				"newAccount": "https://acme.test/new-account",
				"newOrder": "https://acme.test/new-order",
				"revokeCert": "https://acme.test/revoke-cert"
			}`),
			wantFetch:    true,
			wantLogMatch: "'newNonce'",
		},
		{
			name:   "missing newAccount",
			dirURI: testDirURI,
			handler: respondWith(http.StatusOK, `{
				"newNonce": "https://acme.test/new-nonce",
				"newOrder": "https://acme.test/new-order",
				"revokeCert": "https://acme.test/revoke-cert"
			}`),
			wantFetch:    true,
			wantLogMatch: "'newAccount'",
		},
		{
			name:   "missing newOrder",
			dirURI: testDirURI,
			handler: respondWith(http.StatusOK, `{
				"newNonce": "https://acme.test/new-nonce",
				"newAccount": "https://acme.test/new-account",
				"revokeCert": "https://acme.test/revoke-cert"
			}`),
			wantFetch:    true,
			wantLogMatch: "'newOrder'",
		},
		{
			name:   "missing revokeCert",
			dirURI: testDirURI,
			handler: respondWith(http.StatusOK, `{
				"newNonce": "https://acme.test/new-nonce",
				"newAccount": "https://acme.test/new-account",
				"newOrder": "https://acme.test/new-order"
			}`),
			wantFetch:    true,
			wantLogMatch: "'revokeCert'",
		},
		{
			name:   "required field present but empty string",
			dirURI: testDirURI,
			handler: respondWith(http.StatusOK, `{
				"newNonce": "",
				"newAccount": "https://acme.test/new-account",
				"newOrder": "https://acme.test/new-order",
				"revokeCert": "https://acme.test/revoke-cert"
			}`),
			wantFetch:    true,
			wantLogMatch: "'newNonce'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dm, tr, logs := newTestMgr(t, tt.dirURI, tt.handler)

			before := time.Now()
			assert.NotPanics(t, func() {
				dm.updateDirectory(context.Background())
			})
			after := time.Now()

			// directory is left untouched
			assert.Equal(t, directory{}, dm.directoryObj)

			// a failure backs off briefly (2s + jitter) rather than allowing
			// immediate back-to-back fetches, but does not wait a full interval
			assertNextFetchWithin(t, dm.nextFetchAt, before, after, failBackoff, failBackoffJitter)

			// transport only used when the URI was acceptable
			if tt.wantFetch {
				assert.Equal(t, int64(1), tr.hits.Load())
			} else {
				assert.Equal(t, int64(0), tr.hits.Load())
			}

			// exactly one error is logged and it names the problem
			errs := errorLogs(logs)
			require.Len(t, errs, 1)
			assert.Contains(t, errs[0].Message, tt.wantLogMatch)

			// nothing is reported as a successful change
			assert.Equal(t, 0, logs.FilterMessageSnippet("detected changes").Len())

			// the semaphore is always released
			assertSemaphoreFree(t, dm)
		})
	}
}

func TestUpdateDirectory_FailureBackoff(t *testing.T) {
	dm, tr, _ := newTestMgr(t, testDirURI, respondWithError(errors.New("down")))
	ctx := context.Background()

	dm.updateDirectory(ctx)
	require.Equal(t, int64(1), tr.hits.Load())

	// quick failures must not cause a fetch storm
	for i := 0; i < 5; i++ {
		dm.updateDirectory(ctx)
	}
	assert.Equal(t, int64(1), tr.hits.Load(), "calls inside the backoff window must not refetch")

	// after the backoff window a retry is made
	expire(dm)
	dm.updateDirectory(ctx)
	assert.Equal(t, int64(2), tr.hits.Load())

	// once the server recovers the directory is loaded and normal caching resumes
	tr.setHandler(respondWith(http.StatusOK, fullDirectoryJSON))
	expire(dm)
	dm.updateDirectory(ctx)
	dm.updateDirectory(ctx)

	assert.Equal(t, int64(3), tr.hits.Load())
	assert.Equal(t, "https://acme.test/new-nonce", dm.directoryObj.NewNonce)
}

func TestUpdateDirectory_FailureKeepsLastGoodDirectory(t *testing.T) {
	failures := map[string]roundTripHandler{
		"transport error": respondWithError(errors.New("down")),
		"invalid json":    respondWith(http.StatusOK, `{nope`),
		"missing fields":  respondWith(http.StatusOK, `{"newNonce": "https://x.test/nonce"}`),
		"server error":    respondWith(http.StatusBadGateway, `<html>bad gateway</html>`),
	}

	for name, h := range failures {
		t.Run(name, func(t *testing.T) {
			dm, tr, _ := newTestMgr(t, testDirURI, respondWith(http.StatusOK, fullDirectoryJSON))
			ctx := context.Background()

			dm.updateDirectory(ctx)
			good := dm.directoryObj
			require.NotEmpty(t, good.NewNonce)

			tr.setHandler(h)
			expire(dm)

			before := time.Now()
			dm.updateDirectory(ctx)
			after := time.Now()

			assert.Equal(t, good, dm.directoryObj, "a failed refresh must not clobber the last good directory")
			assertNextFetchWithin(t, dm.nextFetchAt, before, after, failBackoff, failBackoffJitter)
		})
	}
}

// ---------------------------------------------------------------------------
// updateDirectory: contexts, timeouts and the semaphore
// ---------------------------------------------------------------------------

func TestUpdateDirectory_CancelledContext(t *testing.T) {
	dm, _, _ := newTestMgr(t, testDirURI, respondWith(http.StatusOK, fullDirectoryJSON))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	assert.NotPanics(t, func() {
		dm.updateDirectory(ctx)
	})
	assert.Equal(t, directory{}, dm.directoryObj, "nothing may be loaded with a cancelled context")
	assertSemaphoreFree(t, dm)

	// a later call with a live context succeeds
	expire(dm)
	dm.updateDirectory(context.Background())
	assert.Equal(t, "https://acme.test/new-nonce", dm.directoryObj.NewNonce)
}

func TestUpdateDirectory_ParentContextTimeoutAbortsFetch(t *testing.T) {
	dm, _, logs := newTestMgr(t, testDirURI, respondSlowly())

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	start := time.Now()
	dm.updateDirectory(ctx)
	elapsed := time.Since(start)

	assert.Less(t, elapsed, 5*time.Second, "the caller's shorter deadline must win over the internal 10s timeout")
	assert.Equal(t, directory{}, dm.directoryObj)
	require.Len(t, errorLogs(logs), 1)
	assert.Contains(t, errorLogs(logs)[0].Message, "failed to do http request")
	assertSemaphoreFree(t, dm)
}

func TestUpdateDirectory_GivesUpWhenSemaphoreIsBusy(t *testing.T) {
	dm, tr, logs := newTestMgr(t, testDirURI, respondWith(http.StatusOK, fullDirectoryJSON))

	// simulate another fetch being in flight
	require.NoError(t, dm.semHttp.Acquire(context.Background(), 1))

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	start := time.Now()
	dm.updateDirectory(ctx)
	elapsed := time.Since(start)

	assert.Less(t, elapsed, 2*time.Second)
	assert.Equal(t, int64(0), tr.hits.Load(), "must not fetch without holding the semaphore")
	assert.Equal(t, directory{}, dm.directoryObj)
	assert.True(t, dm.nextFetchAt.IsZero(), "an update that never ran must not schedule anything")
	assert.Empty(t, errorLogs(logs))

	// once the semaphore frees up, updates work again
	dm.semHttp.Release(1)
	dm.updateDirectory(context.Background())

	assert.Equal(t, int64(1), tr.hits.Load())
	assert.Equal(t, "https://acme.test/new-nonce", dm.directoryObj.NewNonce)
}

func TestUpdateDirectory_CallerDuringInFlightFetchWaitsAndDoesNotRefetch(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	var once sync.Once

	dm, tr, _ := newTestMgr(t, testDirURI, func(req *http.Request) (*http.Response, error) {
		once.Do(func() { close(started) })
		<-release
		return respondWith(http.StatusOK, fullDirectoryJSON)(req)
	})
	ctx := context.Background()

	doneA := make(chan struct{})
	go func() {
		defer close(doneA)
		dm.updateDirectory(ctx)
	}()

	<-started // the first fetch is now in flight and holds the semaphore

	doneB := make(chan struct{})
	go func() {
		defer close(doneB)
		dm.updateDirectory(ctx)
	}()

	// the second caller must be waiting on the in-flight fetch, not finished
	select {
	case <-doneB:
		t.Fatal("second caller returned before the in-flight fetch completed")
	case <-time.After(200 * time.Millisecond):
	}

	close(release)

	<-doneA
	select {
	case <-doneB:
	case <-time.After(5 * time.Second):
		t.Fatal("second caller did not return after the in-flight fetch completed")
	}

	assert.Equal(t, int64(1), tr.hits.Load(), "overlapping callers must share one fetch")
	assert.Equal(t, "https://acme.test/new-nonce", dm.directoryObj.NewNonce)
}

func TestUpdateDirectory_ConcurrentFirstFetch(t *testing.T) {
	dm, tr, _ := newTestMgr(t, testDirURI, respondWith(http.StatusOK, fullDirectoryJSON))
	ctx := context.Background()

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			dm.updateDirectory(ctx)
		}()
	}
	wg.Wait()

	assert.Equal(t, int64(1), tr.hits.Load(), "concurrent callers must result in a single fetch")
	assert.Equal(t, "https://acme.test/new-nonce", dm.directoryObj.NewNonce)
	assertSemaphoreFree(t, dm)
}
