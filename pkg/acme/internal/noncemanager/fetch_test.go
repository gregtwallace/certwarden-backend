package noncemanager

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// To keep retries fast, "failing" responses carry "Retry-After: 0", which makes
// fetchNonceOnce return a ~0 wait instead of defaultTilNext(). This assumes
// validation.ParseRetryAfter accepts "0"; if it doesn't, use a past HTTP date.

// testResponse builds a minimal HEAD-style response with the given headers.
func testResponse(r *http.Request, headers map[string]string) *http.Response {
	h := http.Header{}
	for k, v := range headers {
		h.Set(k, v)
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     h,
		Body:       io.NopCloser(strings.NewReader("")),
		Request:    r,
	}
}

func newTestManager(rt http.RoundTripper) *Manager {
	return &Manager{
		httpClient:  &http.Client{Transport: rt},
		newNonceUrl: strPtr("https://example.invalid/new-nonce"),
	}
}

func TestManagerFetchNonce_SuccessFirstTry(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.Header().Set("Replay-Nonce", "first-nonce")
	}))
	defer srv.Close()

	url := srv.URL
	manager := &Manager{httpClient: srv.Client(), newNonceUrl: &url}

	nonce, err := manager.fetchNonce(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if nonce != "first-nonce" {
		t.Errorf("nonce = %q, want %q", nonce, "first-nonce")
	}
	if got := calls.Load(); got != 1 {
		t.Errorf("requests = %d, want 1", got)
	}
}

func TestManagerFetchNonce_SuccessAfterRetries(t *testing.T) {
	var calls atomic.Int32
	manager := newTestManager(roundTripFunc(func(r *http.Request) (*http.Response, error) {
		n := calls.Add(1)
		if n < nonceFetchMaxRetries {
			// no nonce, but ask for an immediate retry
			return testResponse(r, map[string]string{"Retry-After": "0"}), nil
		}
		return testResponse(r, map[string]string{"Replay-Nonce": "late-nonce"}), nil
	}))

	nonce, err := manager.fetchNonce(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if nonce != "late-nonce" {
		t.Errorf("nonce = %q, want %q", nonce, "late-nonce")
	}
	if got := calls.Load(); got != nonceFetchMaxRetries {
		t.Errorf("requests = %d, want %d", got, nonceFetchMaxRetries)
	}
}

func TestManagerFetchNonce_ExhaustsRetries(t *testing.T) {
	var calls atomic.Int32
	manager := newTestManager(roundTripFunc(func(r *http.Request) (*http.Response, error) {
		calls.Add(1)
		return testResponse(r, map[string]string{"Retry-After": "0"}), nil
	}))

	nonce, err := manager.fetchNonce(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if nonce != "" {
		t.Errorf("nonce = %q, want empty", nonce)
	}
	if !errors.Is(err, errRetriesExhausted) {
		t.Errorf("error = %q, want it to mention exhausted retries", err.Error())
	}
	if got := calls.Load(); got != nonceFetchMaxRetries {
		t.Errorf("requests = %d, want exactly %d", got, nonceFetchMaxRetries)
	}
}

func TestManagerFetchNonce_ContextAlreadyCanceled(t *testing.T) {
	var calls atomic.Int32
	manager := newTestManager(roundTripFunc(func(r *http.Request) (*http.Response, error) {
		calls.Add(1)
		return testResponse(r, map[string]string{"Replay-Nonce": "unused"}), nil
	}))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	nonce, err := manager.fetchNonce(ctx)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if nonce != "" {
		t.Errorf("nonce = %q, want empty", nonce)
	}
	if !errors.Is(err, errContextCanceled) {
		t.Errorf("error = %q, want it to mention ctx canceled", err.Error())
	}
	// the first wait is 0, but a cancelled ctx should still abort before any request
	// is made (select may pick either ready case, so allow at most one request)
	if got := calls.Load(); got > 1 {
		t.Errorf("requests = %d, want at most 1", got)
	}
}

func TestManagerFetchNonce_ContextCanceledWhileWaiting(t *testing.T) {
	var calls atomic.Int32
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	manager := newTestManager(roundTripFunc(func(r *http.Request) (*http.Response, error) {
		calls.Add(1)
		// cancel during the first request and tell the caller to wait a long time,
		// so the next iteration must exit through ctx.Done() rather than the timer
		cancel()
		return testResponse(r, map[string]string{"Retry-After": "3600"}), nil
	}))

	start := time.Now()
	nonce, err := manager.fetchNonce(ctx)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if nonce != "" {
		t.Errorf("nonce = %q, want empty", nonce)
	}
	if !errors.Is(err, errContextCanceled) {
		t.Errorf("error = %q, want it to mention ctx canceled", err.Error())
	}
	if got := calls.Load(); got != 1 {
		t.Errorf("requests = %d, want 1", got)
	}
	if elapsed > 5*time.Second {
		t.Errorf("fetchNonce took %v, should have returned promptly after cancel", elapsed)
	}
}

func TestManagerFetchNonce_ErrorsAreDistinct(t *testing.T) {
	// exhausted-retries and ctx-canceled paths should produce different errors
	exhausted := newTestManager(roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return testResponse(r, map[string]string{"Retry-After": "0"}), nil
	}))
	_, errExhausted := exhausted.fetchNonce(context.Background())

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, errCanceled := exhausted.fetchNonce(ctx)

	if errExhausted == nil || errCanceled == nil {
		t.Fatalf("expected both errors, got exhausted=%v canceled=%v", errExhausted, errCanceled)
	}
	if errExhausted.Error() == errCanceled.Error() {
		t.Errorf("errors should differ, both were %q", errExhausted.Error())
	}
}
