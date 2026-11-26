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
// Getters
// ---------------------------------------------------------------------------

func TestGetters_FullDirectory(t *testing.T) {
	dm, _, _ := newTestMgr(t, testDirURI, respondWith(http.StatusOK, fullDirectoryJSON))
	ctx := context.Background()

	assert.Equal(t, "https://acme.test/new-nonce", dm.NewNonceURL(ctx))
	assert.Equal(t, "https://acme.test/new-account", dm.NewAccountURL(ctx))
	assert.Equal(t, "https://acme.test/new-order", dm.NewOrderURL(ctx))
	assert.Equal(t, "https://acme.test/revoke-cert", dm.RevokeCertURL(ctx))
	assert.Equal(t, "https://acme.test/key-change", dm.KeyChangeURL(ctx))
	assert.Equal(t, "https://acme.test/renewal-info", dm.RenewalInfoURL(ctx))

	assert.Equal(t, "https://acme.test/tos", dm.TermsOfServiceURLCached())
	assert.True(t, dm.RequiresEABCached())
	assert.Equal(t, map[string]string{
		"classic":    "The default profile",
		"shortlived": "A short lived profile",
	}, dm.ProfilesCached())
	assert.True(t, dm.ProfileIsPublishedCached("classic"))
	assert.True(t, dm.SupportsRenewalInfoCached())
	assert.Equal(t, fullDirectoryJSON, string(dm.RawResponseCached()))
}

func TestGetters_RequiredFieldsOnly(t *testing.T) {
	dm, _, _ := newTestMgr(t, testDirURI, respondWith(http.StatusOK, requiredOnlyDirectoryJSON))
	ctx := context.Background()

	assert.Equal(t, "https://acme.test/new-nonce", dm.NewNonceURL(ctx))
	assert.Equal(t, "https://acme.test/new-account", dm.NewAccountURL(ctx))
	assert.Equal(t, "https://acme.test/new-order", dm.NewOrderURL(ctx))
	assert.Equal(t, "https://acme.test/revoke-cert", dm.RevokeCertURL(ctx))

	// optional values fall back to their zero values
	assert.Empty(t, dm.KeyChangeURL(ctx))
	assert.Empty(t, dm.RenewalInfoURL(ctx))
	assert.Empty(t, dm.TermsOfServiceURLCached())
	assert.False(t, dm.RequiresEABCached())
	assert.Nil(t, dm.ProfilesCached())
	assert.False(t, dm.ProfileIsPublishedCached("classic"))
	assert.False(t, dm.SupportsRenewalInfoCached())
}

func TestGetters_ReflectDirectoryChanges(t *testing.T) {
	dm, tr, _ := newTestMgr(t, testDirURI, respondWith(http.StatusOK, fullDirectoryJSON))
	ctx := context.Background()

	require.Equal(t, "https://acme.test/new-nonce", dm.NewNonceURL(ctx))
	require.True(t, dm.RequiresEABCached())
	require.True(t, dm.ProfileIsPublishedCached("classic"))
	require.True(t, dm.SupportsRenewalInfoCached())

	tr.setHandler(respondWith(http.StatusOK, altDirectoryJSON))
	expire(dm)

	assert.Equal(t, "https://acme2.test/new-nonce", dm.NewNonceURL(ctx))
	assert.False(t, dm.RequiresEABCached())
	assert.False(t, dm.ProfileIsPublishedCached("classic"))
	assert.False(t, dm.SupportsRenewalInfoCached())
	assert.Empty(t, dm.RenewalInfoURL(ctx))
	assert.Equal(t, altDirectoryJSON, string(dm.RawResponseCached()))
}

func TestGetters_EachTriggersUpdate(t *testing.T) {
	// per the note in the source, every getter must call updateDirectory
	// itself and cannot rely on another getter having run first
	for _, g := range allGetters() {
		t.Run(g.name, func(t *testing.T) {
			dm, tr, _ := newTestMgr(t, testDirURI, respondWith(http.StatusOK, fullDirectoryJSON))

			got := g.call(dm)

			assert.Equal(t, int64(1), tr.hits.Load(), "getter must trigger the initial fetch")
			assert.NotEmpty(t, got, "getter must return data from the freshly fetched directory")
		})
	}
}

func TestGetters_DirectoryFetchedOnce(t *testing.T) {
	dm, tr, _ := newTestMgr(t, testDirURI, respondWith(http.StatusOK, fullDirectoryJSON))
	getters := allGetters()

	for i := 0; i < 10; i++ {
		for _, g := range getters {
			_ = g.call(dm)
		}
	}

	assert.Equal(t, int64(1), tr.hits.Load())
}

func TestGetters_FetchFailureReturnsZeroValues(t *testing.T) {
	dm, _, _ := newTestMgr(t, testDirURI, respondWithError(errors.New("down")))
	ctx := context.Background()

	assert.NotPanics(t, func() {
		assert.Empty(t, dm.NewNonceURL(ctx))
		assert.Empty(t, dm.NewAccountURL(ctx))
		assert.Empty(t, dm.NewOrderURL(ctx))
		assert.Empty(t, dm.RevokeCertURL(ctx))
		assert.Empty(t, dm.KeyChangeURL(ctx))
		assert.Empty(t, dm.RenewalInfoURL(ctx))
		assert.Empty(t, dm.TermsOfServiceURLCached())
		assert.False(t, dm.RequiresEABCached())
		assert.Nil(t, dm.RawResponseCached())
		assert.Nil(t, dm.ProfilesCached())
		assert.False(t, dm.ProfileIsPublishedCached("classic"))
		assert.False(t, dm.SupportsRenewalInfoCached())
	})
}

func TestGetters_InvalidDirURIReturnsZeroValues(t *testing.T) {
	dm, tr, _ := newTestMgr(t, "http://acme.test/directory", respondWith(http.StatusOK, fullDirectoryJSON))

	for _, g := range allGetters() {
		assert.Empty(t, g.call(dm), g.name)
	}
	assert.Equal(t, int64(0), tr.hits.Load(), "a non-https dirURI must never be fetched")
}

func TestGetters_CtxGettersRespectCallerContext(t *testing.T) {
	for _, g := range []struct {
		name string
		call func(dm *DirectoryMgr, ctx context.Context) string
	}{
		{"NewNonceURL", func(dm *DirectoryMgr, ctx context.Context) string { return dm.NewNonceURL(ctx) }},
		{"NewAccountURL", func(dm *DirectoryMgr, ctx context.Context) string { return dm.NewAccountURL(ctx) }},
		{"NewOrderURL", func(dm *DirectoryMgr, ctx context.Context) string { return dm.NewOrderURL(ctx) }},
		{"RevokeCertURL", func(dm *DirectoryMgr, ctx context.Context) string { return dm.RevokeCertURL(ctx) }},
		{"KeyChangeURL", func(dm *DirectoryMgr, ctx context.Context) string { return dm.KeyChangeURL(ctx) }},
		{"RenewalInfoURL", func(dm *DirectoryMgr, ctx context.Context) string { return dm.RenewalInfoURL(ctx) }},
	} {
		t.Run(g.name, func(t *testing.T) {
			t.Parallel()

			dm, _, _ := newTestMgr(t, testDirURI, respondSlowly())

			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()

			start := time.Now()
			got := g.call(dm, ctx)

			assert.Empty(t, got)
			assert.Less(t, time.Since(start), 5*time.Second)
		})
	}
}

func TestGetters_CachedGettersCapTheirWait(t *testing.T) {
	// Cached getters have no ctx parameter; they must give up after
	// noCacheWait (1s) instead of waiting out the 10s internal fetch timeout.
	for _, g := range cachedGetters() {
		t.Run(g.name, func(t *testing.T) {
			t.Parallel()

			dm, tr, _ := newTestMgr(t, testDirURI, respondSlowly())

			start := time.Now()
			got := g.call(dm)
			elapsed := time.Since(start)

			assert.Equal(t, int64(1), tr.hits.Load())
			assert.Empty(t, got, "no directory could be fetched, so the zero value is expected")
			assert.GreaterOrEqual(t, elapsed, noCacheWait-100*time.Millisecond)
			assert.Less(t, elapsed, 3*time.Second)
		})
	}
}

func TestCachedGetters_ReturnImmediatelyWhenNotDue(t *testing.T) {
	dm, tr, _ := newTestMgr(t, testDirURI, respondWith(http.StatusOK, fullDirectoryJSON))
	_ = dm.NewNonceURL(context.Background()) // load once
	require.Equal(t, int64(1), tr.hits.Load())

	// even if the server hangs from now on, a cached read is instant
	tr.setHandler(respondSlowly())

	for _, g := range cachedGetters() {
		start := time.Now()
		got := g.call(dm)

		assert.NotEmpty(t, got, g.name)
		assert.Less(t, time.Since(start), 100*time.Millisecond, g.name)
	}
	assert.Equal(t, int64(1), tr.hits.Load())
}

func TestProfilesCached(t *testing.T) {
	t.Run("published", func(t *testing.T) {
		dm, _, _ := newTestMgr(t, testDirURI, respondWith(http.StatusOK, fullDirectoryJSON))

		assert.Equal(t, map[string]string{
			"classic":    "The default profile",
			"shortlived": "A short lived profile",
		}, dm.ProfilesCached())
	})

	t.Run("extension not implemented returns nil", func(t *testing.T) {
		dm, _, _ := newTestMgr(t, testDirURI, respondWith(http.StatusOK, requiredOnlyDirectoryJSON))

		assert.Nil(t, dm.ProfilesCached())
	})
}

func TestProfileIsPublishedCached(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		profile string
		want    bool
	}{
		{"existing profile", fullDirectoryJSON, "classic", true},
		{"another existing profile", fullDirectoryJSON, "shortlived", true},
		{"missing profile", fullDirectoryJSON, "tlsserver", false},
		{"empty profile name", fullDirectoryJSON, "", false},
		{"case sensitive", fullDirectoryJSON, "Classic", false},
		{"matches keys, not descriptions", fullDirectoryJSON, "The default profile", false},
		{"no profiles map at all", requiredOnlyDirectoryJSON, "classic", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dm, _, _ := newTestMgr(t, testDirURI, respondWith(http.StatusOK, tt.body))

			assert.Equal(t, tt.want, dm.ProfileIsPublishedCached(tt.profile))
		})
	}
}

func TestRenewalInfo(t *testing.T) {
	tests := []struct {
		name         string
		body         string
		wantURL      string
		wantSupports bool
	}{
		{"published", fullDirectoryJSON, "https://acme.test/renewal-info", true},
		{"key absent", requiredOnlyDirectoryJSON, "", false},
		{"key present but empty", emptyRenewalInfoDirectoryJSON, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dm, _, _ := newTestMgr(t, testDirURI, respondWith(http.StatusOK, tt.body))

			assert.Equal(t, tt.wantURL, dm.RenewalInfoURL(context.Background()))
			assert.Equal(t, tt.wantSupports, dm.SupportsRenewalInfoCached())
		})
	}
}

func TestRequiresEABCached(t *testing.T) {
	t.Run("required", func(t *testing.T) {
		dm, _, _ := newTestMgr(t, testDirURI, respondWith(http.StatusOK, fullDirectoryJSON))

		assert.True(t, dm.RequiresEABCached())
	})

	t.Run("not required", func(t *testing.T) {
		dm, _, _ := newTestMgr(t, testDirURI, respondWith(http.StatusOK, requiredOnlyDirectoryJSON))

		assert.False(t, dm.RequiresEABCached())
	})
}

func TestRawResponseCached(t *testing.T) {
	dm, _, _ := newTestMgr(t, testDirURI, respondWith(http.StatusOK, fullDirectoryJSON))

	// stored exactly as the server sent it, including unknown fields and whitespace
	assert.Equal(t, fullDirectoryJSON, string(dm.RawResponseCached()))
}

// ---------------------------------------------------------------------------
// Concurrency (run with -race)
// ---------------------------------------------------------------------------

func TestGetters_ConcurrentReadsOfLoadedDirectory(t *testing.T) {
	dm, tr, _ := newTestMgr(t, testDirURI, respondWith(http.StatusOK, fullDirectoryJSON))

	// load the directory first so the readers below only ever see a cached,
	// not-due directory
	require.Equal(t, "https://acme.test/new-nonce", dm.NewNonceURL(context.Background()))

	ctx := context.Background()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			assert.Equal(t, "https://acme.test/new-nonce", dm.NewNonceURL(ctx))
			assert.Equal(t, "https://acme.test/new-order", dm.NewOrderURL(ctx))
			assert.Equal(t, "https://acme.test/renewal-info", dm.RenewalInfoURL(ctx))
			assert.True(t, dm.RequiresEABCached())
			assert.True(t, dm.ProfileIsPublishedCached("classic"))
			assert.True(t, dm.SupportsRenewalInfoCached())
			assert.NotEmpty(t, dm.RawResponseCached())
			assert.NotEmpty(t, dm.ProfilesCached())
		}()
	}
	wg.Wait()

	assert.Equal(t, int64(1), tr.hits.Load())
}
