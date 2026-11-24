package app

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"unicode"
)

// TODO: Change package name to `app_test`
// TODO: Test other misc handlers

func TestAppHttpToHttpsRedirectHandler(t *testing.T) {
	// test cases
	tc := []struct {
		requestTarget string
		expectSuccess bool
	}{
		// should work
		{"http://something.example.com:80/someplace/index.html?somethingelse=52", true},
		{"http://something.example.com/someplace/index.html?somethingelse=52", true},
		{"http://something.example.com:4050", true},
		{"http://notexplicit.example.com", true},
		{"http://notexplicit.example.com:5222", true},
		{"http://10.10.0.101", true},
		{"http://10.10.0.101:8080", true},
		{"http://[2001:db8::1234]", true},
		{"http://[2001:db8::1234]:8080", true},

		// should not work
		{"", false},
		{"http://localhost:80/certwarden/health", false},
		{"http://localhost:80/certwarden/api/v1/app/auth/status", false},
		{"http://something.examplex.com:80/someplace/index.html?somethingelse=52", false},
		{"http://something.examplex.com/someplace/index.html?somethingelse=52", false},
		{"http://192.168.100.1:80/someplace/index.html?somethingelse=52", false},
		{"http://notexplicit.examplex.com", false},
		{"http://notexplicit.examplex.com:5222", false},
		{"http://192.168.100.1", false},
		{"http://192.168.100.1:8080", false},
		{"http://[2001:db8::5555]", false},
		{"http://[2001:db8::5555]:8080", false},
	}

	// make dummy self-signed cert
	sc, err := NewTestSafecert([]string{"example.com", "something.example.com", "*.example.com"},
		[]string{"10.10.0.101", "2001:db8::1234"})
	if err != nil {
		t.Fatalf("failed to make test cert (%s)", err)
	}

	// bare minimum app to make function work
	a := Application{
		config: &config{
			HttpsPort: new(4055),
		},
		httpsCert: sc,
	}

	// run tests
	for i := range tc {
		t.Run(fmt.Sprintf("redirect %d: %s", i, tc[i].requestTarget), func(t *testing.T) {
			// do request/write
			w := httptest.NewRecorder()
			var r *http.Request
			if tc[i].requestTarget != "" {
				r = httptest.NewRequest(http.MethodGet, tc[i].requestTarget, http.NoBody)
			} else {
				// special case if trying to set Host to ""
				r = httptest.NewRequest(http.MethodGet, "http://www.example.com", http.NoBody)
				r.Host = ""
			}

			a.httpToHttpsRedirectHandler(w, r)

			// check result
			res := w.Result()
			defer res.Body.Close()
			resDat, err := io.ReadAll(res.Body)
			if err != nil {
				t.Errorf("expected error to be nil got %v", err)
			}
			// trim off trailing white space
			resDat = bytes.TrimRightFunc(resDat, unicode.IsSpace)

			// check: response code
			if !tc[i].expectSuccess && res.StatusCode != http.StatusBadRequest {
				t.Errorf("expected status 400 bad request but got '%d'", res.StatusCode)
			}

			if tc[i].expectSuccess && res.StatusCode != http.StatusTemporaryRedirect {
				t.Errorf("expected status 307 temporary redirect but got '%d'", res.StatusCode)
			}

			// check: body
			hostName, _, err := net.SplitHostPort(r.Host)
			if err != nil {
				// if it failed, assume there is no port and use raw r.Host
				hostName = r.Host
			}
			var expectedBody []byte

			// expect error
			if !tc[i].expectSuccess {
				expectedBody = fmt.Appendf(expectedBody, "redirect failed: https hostname '%s' is not a part of the server's certificate", hostName)
			} else {
				// expect ok
				newAddr := "https://" + hostName + ":" + strconv.Itoa(*a.config.HttpsPort)
				expectedBody = fmt.Appendf(expectedBody, "<a href=%q>Temporary Redirect</a>.", newAddr)
			}

			if !bytes.Equal(resDat, expectedBody) {
				t.Errorf("expected body '%s' but got '%s'", expectedBody, resDat)
			}
		})
	}

	// special case for health endpoint
	// test cases
	tcHealth := []struct {
		requestTarget string
	}{
		// should work
		{"http://localhost:80/certwarden/api/health"},
		{"http://127.0.0.1:80/certwarden/api/health"},
		{"http://9.9.9.9:80/certwarden/api/health"},
		{"http://[::1]:80/certwarden/api/health"},
		{"http://localhost:4050/certwarden/api/health"},
		{"http://127.0.0.1:4050/certwarden/api/health"},
		{"http://9.9.9.9:4050/certwarden/api/health"},
		{"http://[::1]:4050/certwarden/api/health"},
	}

	for i := range tcHealth {
		t.Run(fmt.Sprintf("health GET %d: %s", i, tcHealth[i].requestTarget), func(t *testing.T) {
			// do request/write
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, tcHealth[i].requestTarget, http.NoBody)

			a.httpToHttpsRedirectHandler(w, r)

			// check result
			res := w.Result()
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			if err != nil {
				t.Errorf("expected error to be nil got %v", err)
			}
			if !bytes.Equal(resBody, []byte{}) {
				t.Errorf("expected body '%s' but got '%s'", []byte{}, resBody)
			}

			// check: response code
			if res.StatusCode != http.StatusNoContent {
				t.Errorf("expected status 204 no content but got '%d'", res.StatusCode)
			}
		})
	}

	// do it again but with HEAD method
	for i := range tcHealth {
		t.Run(fmt.Sprintf("health HEAD %d: %s", i, tcHealth[i].requestTarget), func(t *testing.T) {
			// do request/write
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodHead, tcHealth[i].requestTarget, http.NoBody)

			a.httpToHttpsRedirectHandler(w, r)

			// check result
			res := w.Result()
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			if err != nil {
				t.Errorf("expected error to be nil got %v", err)
			}
			if !bytes.Equal(resBody, []byte{}) {
				t.Errorf("expected body '%s' but got '%s'", []byte{}, resBody)
			}

			// check: response code
			if res.StatusCode != http.StatusNoContent {
				t.Errorf("expected status 204 no content but got '%d'", res.StatusCode)
			}
		})
	}
}
