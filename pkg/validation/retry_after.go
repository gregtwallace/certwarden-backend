package validation

import (
	"errors"
	"net/http"
	"strconv"
	"time"
)

var (
	ErrHTTPRetryAfterNegativeSeconds = errors.New("retry-after seconds were an invalid negative value")
	ErrHTTPRetryAfterInvalidFormat   = errors.New("retry-after value was not validly formatted")
)

// parseRetryAfter is an internal implementation
// DO NOT EXPORT
func parseRetryAfter(retryAfter string, nowFunc func() time.Time) (time.Time, error) {
	now := nowFunc().Round(time.Second)

	// is value a number of seconds?
	secs, err := strconv.Atoi(retryAfter)
	if err == nil {
		// negative is invalid
		if secs < 0 {
			return time.Time{}, ErrHTTPRetryAfterNegativeSeconds
		}

		return now.Add(time.Duration(secs) * time.Second), nil
	}

	// is value an HTTP date?
	t, err := http.ParseTime(retryAfter)
	if err == nil {
		// negative is invalid
		until := t.Sub(now)
		if until < 0 {
			return time.Time{}, ErrHTTPRetryAfterNegativeSeconds
		}

		return t, nil
	}

	// neither valid format was found
	return time.Time{}, ErrHTTPRetryAfterInvalidFormat
}

// ParseRetryAfter parses the string value of Retry-After header value string; if
// the value is invalid, an error is returned.
// see: https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Retry-After
func ParseRetryAfter(retryAfter string) (time.Time, error) {
	return parseRetryAfter(retryAfter, time.Now)
}
