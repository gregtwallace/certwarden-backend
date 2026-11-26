package validation_test

import (
	"certwarden-backend/pkg/validation"
	"fmt"
	"testing"
)

func TestValidation_IsValidHTTPSURL(t *testing.T) {
	tcs := []struct {
		maybeValidURL string
		expectedOK    bool
	}{
		// empty
		{"", false},

		// some valid https formats
		{"https://example.com", true},
		{"https://example.com/path", true},
		{"https://example.com:8080", true},
		{"https://example.com?query=param", true},
		{"https://example.com#fragment", true},

		// other invalid formats
		{"http://example.com", false},
		{"ftp://example.com", false},
		{"https://", false},
		{"not a url", false},
	}

	// run tests
	for i := range tcs {
		t.Run(fmt.Sprintf("%d: %q", i, tcs[i].maybeValidURL), func(t *testing.T) {
			actualOK := validation.IsValidHTTPSURL(tcs[i].maybeValidURL)

			if actualOK != tcs[i].expectedOK {
				t.Errorf("string value '%s' expected ok '%t' but got '%t'", tcs[i].maybeValidURL, tcs[i].expectedOK, actualOK)
			}
		})
	}
}
