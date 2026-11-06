package noncemanager

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

func TestDefaultTilNext(t *testing.T) {
	// test 100 random values
	for i := 0; i < 100; i++ {
		d := defaultTilNext()
		if d < 1*time.Second || d > 1*time.Second+500*time.Millisecond {
			t.Errorf("defaultTilNext returned %v, but expected value between 1s and 1.5s", d)
		}
	}
}

// roundTripFunc lets tests supply an http.RoundTripper as a plain function.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func strPtr(s string) *string { return &s }

func TestFetchNonceOnce_NilURL(t *testing.T) {
	nonce, next := fetchNonceOnce(context.Background(), http.DefaultClient, nil)
	if nonce != "" {
		t.Errorf("nonce = %q, want empty", nonce)
	}
	if next <= 0 {
		t.Errorf("nextTry = %v, want > 0", next)
	}
}

func TestFetchNonceOnce_EmptyURL(t *testing.T) {
	nonce, next := fetchNonceOnce(context.Background(), http.DefaultClient, strPtr(""))
	if nonce != "" {
		t.Errorf("nonce = %q, want empty", nonce)
	}
	if next <= 0 {
		t.Errorf("nextTry = %v, want > 0", next)
	}
}

func TestFetchNonceOnce_InvalidURL(t *testing.T) {
	// a control character makes http.NewRequestWithContext fail
	nonce, next := fetchNonceOnce(context.Background(), http.DefaultClient, strPtr("http://example.com"))
	if nonce != "" {
		t.Errorf("nonce = %q, want empty", nonce)
	}
	if next <= 0 {
		t.Errorf("nextTry = %v, want > 0", next)
	}
}

func TestFetchNonceOnce_ClientError(t *testing.T) {
	cli := &http.Client{
		Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return nil, errors.New("boom")
		}),
	}
	nonce, next := fetchNonceOnce(context.Background(), cli, strPtr("https://example.invalid/nonce"))
	if nonce != "" {
		t.Errorf("nonce = %q, want empty", nonce)
	}
	if next <= 0 {
		t.Errorf("nextTry = %v, want > 0", next)
	}
}

func TestFetchNonceOnce_ServerUnreachable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	url := srv.URL
	srv.Close() // closed, so connection will be refused

	nonce, next := fetchNonceOnce(context.Background(), srv.Client(), &url)
	if nonce != "" {
		t.Errorf("nonce = %q, want empty", nonce)
	}
	if next <= 0 {
		t.Errorf("nextTry = %v, want > 0", next)
	}
}

func TestFetchNonceOnce_ContextCanceled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Replay-Nonce", "should-not-be-seen")
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	url := srv.URL
	nonce, next := fetchNonceOnce(ctx, srv.Client(), &url)
	if nonce != "" {
		t.Errorf("nonce = %q, want empty", nonce)
	}
	if next <= 0 {
		t.Errorf("nextTry = %v, want > 0", next)
	}
}

func TestFetchNonceOnce_AppliesTimeout(t *testing.T) {
	var deadline time.Time
	var hasDeadline bool

	cli := &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			deadline, hasDeadline = r.Context().Deadline()
			return nil, errors.New("stop here")
		}),
	}

	start := time.Now()
	fetchNonceOnce(context.Background(), cli, strPtr("https://example.invalid/nonce"))

	if !hasDeadline {
		t.Fatal("request context has no deadline")
	}
	if d := deadline.Sub(start); d <= 0 || d > 5*time.Second+500*time.Millisecond {
		t.Errorf("deadline is %v after start, want <= ~5s", d)
	}
}

func TestFetchNonceOnce_Success(t *testing.T) {
	var gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		w.Header().Set("Replay-Nonce", "abc123")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	url := srv.URL
	nonce, next := fetchNonceOnce(context.Background(), srv.Client(), &url)
	if nonce != "abc123" {
		t.Errorf("nonce = %q, want %q", nonce, "abc123")
	}
	if next != 0 {
		t.Errorf("nextTry = %v, want 0", next)
	}
	if gotMethod != http.MethodHead {
		t.Errorf("method = %q, want %q", gotMethod, http.MethodHead)
	}
}

func TestFetchNonceOnce_NonceReturnedOnNon200(t *testing.T) {
	// the function only looks at headers, not status code
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Replay-Nonce", "from-error-response")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	url := srv.URL
	nonce, next := fetchNonceOnce(context.Background(), srv.Client(), &url)
	if nonce != "from-error-response" {
		t.Errorf("nonce = %q, want %q", nonce, "from-error-response")
	}
	if next != 0 {
		t.Errorf("nextTry = %v, want 0", next)
	}
}

func TestFetchNonceOnce_NonceTakesPrecedenceOverRetryAfter(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Replay-Nonce", "n1")
		w.Header().Set("Retry-After", "120")
	}))
	defer srv.Close()

	url := srv.URL
	nonce, next := fetchNonceOnce(context.Background(), srv.Client(), &url)
	if nonce != "n1" {
		t.Errorf("nonce = %q, want %q", nonce, "n1")
	}
	if next != 0 {
		t.Errorf("nextTry = %v, want 0", next)
	}
}

func TestFetchNonceOnce_NoNonceNoRetryAfter(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	defer srv.Close()

	url := srv.URL
	nonce, next := fetchNonceOnce(context.Background(), srv.Client(), &url)
	if nonce != "" {
		t.Errorf("nonce = %q, want empty", nonce)
	}
	if next <= 0 {
		t.Errorf("nextTry = %v, want > 0 (default)", next)
	}
}

func TestFetchNonceOnce_RetryAfterSeconds(t *testing.T) {
	const seconds = 120
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Retry-After", strconv.Itoa(seconds))
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	url := srv.URL
	nonce, next := fetchNonceOnce(context.Background(), srv.Client(), &url)
	if nonce != "" {
		t.Errorf("nonce = %q, want empty", nonce)
	}

	want := seconds * time.Second
	tolerance := 5 * time.Second
	if next < want-tolerance || next > want+tolerance {
		t.Errorf("nextTry = %v, want ~%v", next, want)
	}
}

func TestFetchNonceOnce_RetryAfterHTTPDate(t *testing.T) {
	retryAt := time.Now().Add(90 * time.Second).UTC()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Retry-After", retryAt.Format(http.TimeFormat))
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	url := srv.URL
	nonce, next := fetchNonceOnce(context.Background(), srv.Client(), &url)
	if nonce != "" {
		t.Errorf("nonce = %q, want empty", nonce)
	}

	// http.TimeFormat has 1s resolution, so allow some slack
	want := 90 * time.Second
	tolerance := 5 * time.Second
	if next < want-tolerance || next > want+tolerance {
		t.Errorf("nextTry = %v, want ~%v", next, want)
	}
}

func TestFetchNonceOnce_InvalidRetryAfter(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Retry-After", "not-a-valid-value")
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	url := srv.URL
	nonce, next := fetchNonceOnce(context.Background(), srv.Client(), &url)
	if nonce != "" {
		t.Errorf("nonce = %q, want empty", nonce)
	}
	if next <= 0 {
		t.Errorf("nextTry = %v, want > 0 (default)", next)
	}
}

func TestFetchNonceOnce_ReadsAndClosesBody(t *testing.T) {
	var closed bool
	body := &trackingBody{onClose: func() { closed = true }}

	cli := &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			h := http.Header{}
			h.Set("Replay-Nonce", "xyz")
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     h,
				Body:       body,
				Request:    r,
			}, nil
		}),
	}

	nonce, _ := fetchNonceOnce(context.Background(), cli, strPtr("https://example.invalid/nonce"))
	if nonce != "xyz" {
		t.Errorf("nonce = %q, want %q", nonce, "xyz")
	}
	if !body.drained {
		t.Error("response body was not read to EOF")
	}
	if !closed {
		t.Error("response body was not closed")
	}
}

// trackingBody is an io.ReadCloser that records whether it was drained and closed.
type trackingBody struct {
	data    int // number of bytes handed out so far
	drained bool
	onClose func()
}

func (b *trackingBody) Read(p []byte) (int, error) {
	const payload = 64
	if b.data >= payload {
		b.drained = true
		return 0, io.EOF
	}
	n := len(p)
	if n > payload-b.data {
		n = payload - b.data
	}
	b.data += n
	return n, nil
}

func (b *trackingBody) Close() error {
	if b.onClose != nil {
		b.onClose()
	}
	return nil
}
