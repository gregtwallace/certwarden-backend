package app

import (
	"fmt"
	"net/http"
	"runtime"
	"time"
)

// httpCWRoundTripper implements RoundTrip with some headers for CertWarden
type httpCWRoundTripper struct {
	userAgent string
}

func (t *httpCWRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// always override user-agent
	req.Header.Set("User-Agent", t.userAgent)

	// set preferred language (SHOULD do this per RFC 8555, 6.1)
	// TODO: Implement user choice?
	req.Header.Set("Accept-Language", "en-US, en;q=0.8")

	return http.DefaultTransport.RoundTrip(req)
}

func makeHttpClient() (client *http.Client) {
	t := &httpCWRoundTripper{
		userAgent: fmt.Sprintf("CertWarden/%s (%s; %s)", appVersion, runtime.GOOS, runtime.GOARCH),
	}

	return &http.Client{
		// set client timeout
		Timeout:   30 * time.Second,
		Transport: t,
	}
}
