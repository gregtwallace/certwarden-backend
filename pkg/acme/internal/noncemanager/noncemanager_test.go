package noncemanager

import (
	"certwarden-backend/pkg/datatypes/ringbuffer"
	"certwarden-backend/pkg/helpers_test"
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync/atomic"
	"testing"
)

func TestNew(t *testing.T) {
	httpClient := http.DefaultClient
	nonceURL := "https://example.com/acme/new-nonce"

	mgr := New(httpClient, &nonceURL)

	if mgr.httpClient != httpClient {
		t.Errorf("httpClient not set correctly, expected %v, got %v", httpClient, mgr.httpClient)
	}

	if mgr.newNonceUrl != &nonceURL {
		t.Errorf("newNonceUrl not set correctly, expected %v, got %v", nonceURL, mgr.newNonceUrl)
	}

	if mgr.nonceValues == nil {
		t.Errorf("nonceValues not set correctly, expected non-nil, got nil")
	}
}

func TestPush(t *testing.T) {
	// manually set struct values that are needed for lenUnsafe
	testCases := []struct {
		pushVal     string
		expectedErr error
		expectedVal string
	}{
		// limited tests since the buffer is tested separately
		{"", errPushEmptyNonce, ""}, // empty string
		{"abcdef", nil, "abcdef"},   // valid string
	}

	httpClient := http.DefaultClient
	nonceURL := "https://example.com/acme/new-nonce"

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("#%d", i), func(t *testing.T) {
			mgr := New(httpClient, &nonceURL)

			// push new data
			err := mgr.PushNonce(tc.pushVal)

			// check results
			if !helpers_test.ErrorsIs(err, tc.expectedErr) {
				t.Errorf("expected error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedErr), helpers_test.ErrorToVal(err))
			} else {
				// check that the value was pushed to the buffer
				val, _ := mgr.nonceValues.Pop()
				if val != tc.expectedVal {
					t.Errorf("expected value '%s' but got '%s'", tc.expectedVal, val)
				}
			}
		})
	}
}

// PopNonce tests

// newPopTestManager returns a Manager whose HTTP client counts requests and answers
// with fetchedNonce (or no nonce at all if fetchedNonce is empty), and whose nonce
// buffer is empty.
func newPopTestManager(fetchedNonce string, requests *atomic.Int32) *Manager {
	manager := newTestManager(roundTripFunc(func(r *http.Request) (*http.Response, error) {
		requests.Add(1)
		if fetchedNonce == "" {
			// no nonce, ask for immediate retry so a failing fetch is fast
			return testResponse(r, map[string]string{"Retry-After": "0"}), nil
		}
		return testResponse(r, map[string]string{"Replay-Nonce": fetchedNonce}), nil
	}))

	manager.nonceValues = ringbuffer.New(10)
	return manager
}

// seedNonces puts nonces into the manager's buffer.
func seedNonces(t *testing.T, manager *Manager, nonces ...string) {
	t.Helper()
	for _, n := range nonces {
		manager.nonceValues.Push(n)
	}
}

func TestManagerPopNonce_FromBuffer(t *testing.T) {
	var requests atomic.Int32
	manager := newPopTestManager("fetched", &requests)
	seedNonces(t, manager, "buffered")

	nonce, err := manager.PopNonce(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if nonce != "buffered" {
		t.Errorf("nonce = %q, want %q", nonce, "buffered")
	}
	if got := requests.Load(); got != 0 {
		t.Errorf("requests = %d, want 0 (buffer should satisfy the pop)", got)
	}
}

func TestManagerPopNonce_EmptyBufferFetches(t *testing.T) {
	var requests atomic.Int32
	manager := newPopTestManager("fetched", &requests)

	nonce, err := manager.PopNonce(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if nonce != "fetched" {
		t.Errorf("nonce = %q, want %q", nonce, "fetched")
	}
	if got := requests.Load(); got != 1 {
		t.Errorf("requests = %d, want 1", got)
	}
}

func TestManagerPopNonce_EmptyBufferFetchFails(t *testing.T) {
	var requests atomic.Int32
	manager := newPopTestManager("", &requests) // server never returns a nonce

	nonce, err := manager.PopNonce(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if nonce != "" {
		t.Errorf("nonce = %q, want empty", nonce)
	}
	if got := requests.Load(); got != nonceFetchMaxRetries {
		t.Errorf("requests = %d, want %d (all retries used)", got, nonceFetchMaxRetries)
	}
}

func TestManagerPopNonce_EmptyBufferContextCanceled(t *testing.T) {
	var requests atomic.Int32
	manager := newPopTestManager("fetched", &requests)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	nonce, err := manager.PopNonce(ctx)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if nonce != "" {
		t.Errorf("nonce = %q, want empty", nonce)
	}
	if !errors.Is(err, errContextCanceled) {
		t.Errorf("error = %q, want it to mention ctx canceled", err.Error())
	}
}

func TestManagerPopNonce_BufferThenFetch(t *testing.T) {
	var requests atomic.Int32
	manager := newPopTestManager("fetched", &requests)
	seedNonces(t, manager, "buffered")

	// first pop drains the buffer
	first, err := manager.PopNonce(context.Background())
	if err != nil {
		t.Fatalf("first pop: unexpected error: %v", err)
	}
	if first != "buffered" {
		t.Errorf("first nonce = %q, want %q", first, "buffered")
	}
	if got := requests.Load(); got != 0 {
		t.Fatalf("requests after first pop = %d, want 0", got)
	}

	// second pop finds the buffer empty and fetches
	second, err := manager.PopNonce(context.Background())
	if err != nil {
		t.Fatalf("second pop: unexpected error: %v", err)
	}
	if second != "fetched" {
		t.Errorf("second nonce = %q, want %q", second, "fetched")
	}
	if got := requests.Load(); got != 1 {
		t.Errorf("requests after second pop = %d, want 1", got)
	}
}

func TestManagerPopNonce_BufferedNoncesReturnedInFIFOOrder(t *testing.T) {
	var requests atomic.Int32
	manager := newPopTestManager("fetched", &requests)
	want := []string{"n1", "n2", "n3"}
	seedNonces(t, manager, want...)

	for i, w := range want {
		nonce, err := manager.PopNonce(context.Background())
		if err != nil {
			t.Fatalf("pop %d: unexpected error: %v", i, err)
		}
		if nonce != w {
			t.Errorf("pop %d: nonce = %q, want %q", i, nonce, w)
		}
	}
	if got := requests.Load(); got != 0 {
		t.Errorf("requests = %d, want 0 (all pops served from buffer)", got)
	}

	// buffer is now drained, so the next pop must fetch
	nonce, err := manager.PopNonce(context.Background())
	if err != nil {
		t.Fatalf("final pop: unexpected error: %v", err)
	}
	if nonce != "fetched" {
		t.Errorf("final nonce = %q, want %q", nonce, "fetched")
	}
	if got := requests.Load(); got != 1 {
		t.Errorf("requests after drain = %d, want 1", got)
	}
}

func TestManagerPopNonce_OverflowEvictsOldest(t *testing.T) {
	var requests atomic.Int32
	manager := newPopTestManager("fetched", &requests)
	manager.nonceValues = ringbuffer.New(2)

	// buffer holds 2, so "a" is evicted when "c" is pushed
	seedNonces(t, manager, "a", "b", "c")

	for i, w := range []string{"b", "c"} {
		nonce, err := manager.PopNonce(context.Background())
		if err != nil {
			t.Fatalf("pop %d: unexpected error: %v", i, err)
		}
		if nonce != w {
			t.Errorf("pop %d: nonce = %q, want %q", i, nonce, w)
		}
	}
	if got := requests.Load(); got != 0 {
		t.Errorf("requests = %d, want 0", got)
	}
}
