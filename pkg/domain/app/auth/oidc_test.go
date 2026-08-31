package auth

import (
	"net/url"
	"testing"

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
	const customScope = "api://bb61c7e1-3af1-4b35-b0d2-1cbe60913ba2/certwarden:superadmin"

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

func TestOIDCMissingRequiredScopeUsesConfiguredSuperadminScope(t *testing.T) {
	const customScope = "api://bb61c7e1-3af1-4b35-b0d2-1cbe60913ba2/certwarden:superadmin"
	cfg := new(Config)
	cfg.OIDC.SuperadminScope = customScope
	oauth2Cfg := newOIDCOAuth2Config(cfg, oauth2.Endpoint{})

	tests := []struct {
		name          string
		grantedScopes string
		wantMissing   string
	}{
		{
			name:          "custom scope granted",
			grantedScopes: "openid offline_access profile " + customScope,
			wantMissing:   "",
		},
		{
			name:          "bare scope does not satisfy custom scope",
			grantedScopes: "openid offline_access profile certwarden:superadmin",
			wantMissing:   customScope,
		},
		{
			name:          "standard scope remains required",
			grantedScopes: "openid offline_access " + customScope,
			wantMissing:   "profile",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := oidcMissingRequiredScope(test.grantedScopes, oauth2Cfg.Scopes); got != test.wantMissing {
				t.Errorf("expected missing scope %q, got %q", test.wantMissing, got)
			}
		})
	}
}
