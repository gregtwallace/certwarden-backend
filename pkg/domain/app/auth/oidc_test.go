package auth

import (
	"certwarden-backend/pkg/datatypes/safemap"
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v3"
)

func TestNewOIDCOAuth2ConfigUsesDefaultSuperadminScope(t *testing.T) {
	cfg := new(Config)
	oauth2Cfg := newOIDCOAuth2Config(cfg, oauth2.Endpoint{
		AuthURL: "https://idp.example.com/authorize",
	})

	authURL, err := url.Parse(oauth2Cfg.AuthCodeURL("state"))
	if err != nil {
		t.Fatalf("failed to parse authorization URL: %v", err)
	}

	const want = "openid offline_access profile certwarden:superadmin"
	if got := authURL.Query().Get("scope"); got != want {
		t.Errorf("expected scope %q, got %q", want, got)
	}
}

func TestNewOIDCOAuth2ConfigUsesConfiguredSuperadminScope(t *testing.T) {
	const customScope = "api://example-app/certwarden:superadmin"

	cfg := new(Config)
	err := yaml.Unmarshal([]byte("oidc:\n  superadmin_scope: \""+customScope+"\"\n"), cfg)
	if err != nil {
		t.Fatalf("failed to parse OIDC configuration: %v", err)
	}

	oauth2Cfg := newOIDCOAuth2Config(cfg, oauth2.Endpoint{
		AuthURL: "https://idp.example.com/authorize",
	})
	authURL, err := url.Parse(oauth2Cfg.AuthCodeURL("state"))
	if err != nil {
		t.Fatalf("failed to parse authorization URL: %v", err)
	}

	const want = "openid offline_access profile " + customScope
	if got := authURL.Query().Get("scope"); got != want {
		t.Errorf("expected scope %q, got %q", want, got)
	}
}

func TestOIDCRequiredAuthorizationScopes(t *testing.T) {
	const customScope = "api://example-app/certwarden:superadmin"
	tests := []struct {
		name            string
		configuredScope string
		grantedScopes   string
		wantMissing     string
	}{
		{
			name:            "openid omitted",
			configuredScope: customScope,
			grantedScopes:   "profile offline_access " + customScope,
			wantMissing:     "",
		},
		{
			name:            "profile omitted",
			configuredScope: customScope,
			grantedScopes:   "openid offline_access " + customScope,
			wantMissing:     "",
		},
		{
			name:            "offline access omitted",
			configuredScope: customScope,
			grantedScopes:   "openid profile " + customScope,
			wantMissing:     "",
		},
		{
			name:            "configured superadmin scope omitted",
			configuredScope: customScope,
			grantedScopes:   "openid profile offline_access",
			wantMissing:     customScope,
		},
		{
			name:          "default legacy superadmin scope granted",
			grantedScopes: "certwarden:superadmin",
			wantMissing:   "",
		},
		{
			name:            "custom fully qualified scope granted",
			configuredScope: customScope,
			grantedScopes:   customScope,
			wantMissing:     "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			requiredScopes := oidcRequiredAuthorizationScopes(test.configuredScope)
			if got := oidcMissingRequiredScope(test.grantedScopes, requiredScopes); got != test.wantMissing {
				t.Errorf("expected missing scope %q, got %q", test.wantMissing, got)
			}
		})
	}
}

func TestOIDCTokenResponseScope(t *testing.T) {
	tests := []struct {
		name        string
		extra       map[string]any
		wantScope   string
		wantPresent bool
		wantValid   bool
	}{
		{
			name:        "scope returned",
			extra:       map[string]any{"scope": "certwarden:superadmin"},
			wantScope:   "certwarden:superadmin",
			wantPresent: true,
			wantValid:   true,
		},
		{
			name:      "scope omitted",
			extra:     map[string]any{},
			wantValid: true,
		},
		{
			name:        "scope returned empty",
			extra:       map[string]any{"scope": ""},
			wantPresent: true,
			wantValid:   true,
		},
		{
			name:        "scope malformed",
			extra:       map[string]any{"scope": []string{"certwarden:superadmin"}},
			wantPresent: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			token := new(oauth2.Token).WithExtra(test.extra)
			gotScope, gotPresent, gotValid := oidcTokenResponseScope(token)
			if gotScope != test.wantScope {
				t.Errorf("expected scope %q, got %q", test.wantScope, gotScope)
			}
			if gotPresent != test.wantPresent {
				t.Errorf("expected presence %t, got %t", test.wantPresent, gotPresent)
			}
			if gotValid != test.wantValid {
				t.Errorf("expected validity %t, got %t", test.wantValid, gotValid)
			}
		})
	}
}

func TestOIDCCallbackValidatesOnlySuperadminAuthorizationScope(t *testing.T) {
	const customScope = "api://example-app/certwarden:superadmin"

	tests := []struct {
		name            string
		configuredScope string
		responseScope   any
		wantAuthorized  bool
	}{
		{
			name:            "openid omitted",
			configuredScope: customScope,
			responseScope:   "profile offline_access " + customScope,
			wantAuthorized:  true,
		},
		{
			name:            "profile omitted",
			configuredScope: customScope,
			responseScope:   "openid offline_access " + customScope,
			wantAuthorized:  true,
		},
		{
			name:            "offline access omitted",
			configuredScope: customScope,
			responseScope:   "openid profile " + customScope,
			wantAuthorized:  true,
		},
		{
			name:            "configured superadmin scope omitted",
			configuredScope: customScope,
			responseScope:   "openid profile offline_access",
		},
		{
			name:           "default legacy superadmin scope",
			responseScope:  oidcDefaultSuperadminScope,
			wantAuthorized: true,
		},
		{
			name:            "custom fully qualified scope",
			configuredScope: customScope,
			responseScope:   customScope,
			wantAuthorized:  true,
		},
		{
			name:            "scope response omitted",
			configuredScope: customScope,
			wantAuthorized:  true,
		},
		{
			name:            "scope response empty",
			configuredScope: customScope,
			responseScope:   "",
		},
		{
			name:            "scope response malformed",
			configuredScope: customScope,
			responseScope:   []string{customScope},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service, state, tokenServer := newOIDCCallbackTestService(t, test.configuredScope, test.responseScope)
			defer tokenServer.Close()

			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/callback?state="+state+"&code=provider-code", nil)
			if errJSON := service.OIDCGetCallback(response, request); errJSON != nil {
				t.Fatalf("callback returned JSON error: %v", errJSON)
			}

			_, authorized := service.oidc.pendingSessions.Read(state)
			if authorized != test.wantAuthorized {
				t.Errorf("expected authorization result %t, got %t", test.wantAuthorized, authorized)
			}
		})
	}
}

func TestOIDCRefreshValidatesOnlySuperadminAuthorizationScope(t *testing.T) {
	const customScope = "api://example-app/certwarden:superadmin"

	tests := []struct {
		name          string
		responseScope any
		wantError     bool
	}{
		{
			name:          "custom superadmin scope only",
			responseScope: customScope,
		},
		{
			name:          "custom superadmin scope omitted from grant",
			responseScope: "openid profile offline_access",
			wantError:     true,
		},
		{
			name: "scope response omitted",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := oidcTestTokenResponse(t, test.responseScope)
			tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if err := json.NewEncoder(w).Encode(response); err != nil {
					t.Errorf("failed to encode token response: %v", err)
				}
			}))
			defer tokenServer.Close()

			ctx := oidc.ClientContext(context.Background(), tokenServer.Client())
			extraFuncs := &oidcExtraFuncs{
				ctxWithHttpClient: ctx,
				cfg: &oauth2.Config{
					ClientID:     "test-client",
					ClientSecret: "test-secret",
					Endpoint:     oauth2.Endpoint{TokenURL: tokenServer.URL},
				},
				requiredAuthorizationScopes: oidcRequiredAuthorizationScopes(customScope),
				idTokenVerifier: oidc.NewVerifier(
					"https://idp.example.com",
					nil,
					&oidc.Config{ClientID: "test-client", InsecureSkipSignatureCheck: true},
				),
				token: &expectedToken{RefreshToken: "old-refresh-token"},
			}

			err := extraFuncs.RefreshCheck()
			if (err != nil) != test.wantError {
				t.Fatalf("expected error %t, got %v", test.wantError, err)
			}
		})
	}
}

func newOIDCCallbackTestService(t *testing.T, configuredScope string, responseScope any) (*Service, string, *httptest.Server) {
	t.Helper()

	tokenResponse := oidcTestTokenResponse(t, responseScope)
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(tokenResponse); err != nil {
			t.Errorf("failed to encode token response: %v", err)
		}
	}))

	service := new(Service)
	service.logger = zap.NewNop().Sugar()
	service.oidc.ctxWithHttpClient = oidc.ClientContext(context.Background(), tokenServer.Client())
	service.oidc.pendingSessions = safemap.NewSafeMap[*oidcPendingSession]()
	service.oidc.oauth2Config = &oauth2.Config{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		RedirectURL:  "https://certwarden.example.com/callback",
		Endpoint:     oauth2.Endpoint{TokenURL: tokenServer.URL},
	}
	service.oidc.requiredAuthorizationScopes = oidcRequiredAuthorizationScopes(configuredScope)
	service.oidc.idTokenVerifier = oidc.NewVerifier(
		"https://idp.example.com",
		nil,
		&oidc.Config{ClientID: "test-client", InsecureSkipSignatureCheck: true},
	)

	const state = "test-state"
	callerRedirectURL, err := url.Parse("https://certwarden.example.com/certwarden/")
	if err != nil {
		t.Fatalf("failed to parse caller redirect URL: %v", err)
	}
	service.oidc.pendingSessions.Add(state, &oidcPendingSession{
		callerRedirectUrl: callerRedirectURL,
		codeVerifierHex:   "test-code-verifier",
		createdAt:         time.Now(),
	})

	return service, state, tokenServer
}

func oidcTestTokenResponse(t *testing.T, responseScope any) map[string]any {
	t.Helper()

	response := map[string]any{
		"access_token":  "test-access-token",
		"refresh_token": "test-refresh-token",
		"token_type":    "Bearer",
		"expires_in":    3600,
		"id_token": oidcTestUnsignedIDToken(t, map[string]any{
			"iss": "https://idp.example.com",
			"sub": "test-subject",
			"aud": "test-client",
			"exp": time.Now().Add(time.Hour).Unix(),
		}),
	}
	if responseScope != nil {
		response["scope"] = responseScope
	}

	return response
}

func oidcTestUnsignedIDToken(t *testing.T, claims map[string]any) string {
	t.Helper()

	headerJSON, err := json.Marshal(map[string]string{"alg": "none", "typ": "JWT"})
	if err != nil {
		t.Fatalf("failed to marshal ID token header: %v", err)
	}
	payloadJSON, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("failed to marshal ID token claims: %v", err)
	}

	return base64.RawURLEncoding.EncodeToString(headerJSON) + "." +
		base64.RawURLEncoding.EncodeToString(payloadJSON) + "."
}
