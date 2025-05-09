package app

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/cloudfoundry/jibber_jabber"
	"golang.org/x/time/rate"
)

// rate limit params -- 5 / 10 are the most conservative values published by
// Let's Encrypt at: https://letsencrypt.org/docs/rate-limits/#overall-requests-limit
const (
	limitEventsPerSecond = 5
	limitBurstAtMost     = 10
)

// default Accept-Language header values
const (
	httpClientDefaultLocale   = "en-US"
	httpClientDefaultLanguage = "en"
)

// httpCWRoundTripper implements RoundTrip with some headers for CertWarden
type httpCWRoundTripper struct {
	userAgent      string
	acceptLanguage string
	rateLimiter    *rate.Limiter
}

func (rt *httpCWRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// always override user-agent
	req.Header.Set("User-Agent", rt.userAgent)

	// set preferred language (SHOULD do this per RFC 8555, 6.1)
	req.Header.Set("Accept-Language", rt.acceptLanguage)

	// impose rate limit
	err := rt.rateLimiter.Wait(req.Context())
	if err != nil {
		return nil, err
	}

	return http.DefaultTransport.RoundTrip(req)
}

// makeHttpClient returns an http.Client with a custom transport to ensure certain headers
// are added to all requests
func makeHttpClient() (client *http.Client) {
	// craft Accept-Language value (defaults added last)
	acceptLanguages := []string{}

	// try to get system values for locale and langauge
	userLocale, err := jibber_jabber.DetectIETF()
	// Note: on NON-error
	if err == nil {
		acceptLanguages = append(acceptLanguages, userLocale)
	}

	userLanguage, err := jibber_jabber.DetectLanguage()
	if err == nil {
		acceptLanguages = append(acceptLanguages, userLanguage)
	}

	// append defaults if they're not the system returned values
	if userLocale != httpClientDefaultLocale {
		acceptLanguages = append(acceptLanguages, httpClientDefaultLocale)
	}

	if userLanguage != httpClientDefaultLanguage {
		acceptLanguages = append(acceptLanguages, httpClientDefaultLanguage)
	}

	// assemble header value (acquired vals, then default vals, then final wildcard)
	sb := strings.Builder{}
	for indx, langVal := range acceptLanguages {
		_, _ = sb.WriteString(langVal)
		if indx != 0 {
			_, _ = sb.WriteString(fmt.Sprintf(";q=%.1f", 1-.1*float32(indx)))
		}
		_, _ = sb.WriteString(", ")
	}
	_, _ = sb.WriteString("*;q=0.5")

	// create rate limiter
	rl := rate.NewLimiter(rate.Every(time.Second/limitEventsPerSecond), limitBurstAtMost)

	// make round tripper
	t := &httpCWRoundTripper{
		userAgent:      fmt.Sprintf("CertWarden/%s (%s; %s)", appVersion, runtime.GOOS, runtime.GOARCH),
		acceptLanguage: sb.String(),
		rateLimiter:    rl,
	}

	return &http.Client{
		// set client timeout
		Timeout:   30 * time.Second,
		Transport: t,
	}
}
