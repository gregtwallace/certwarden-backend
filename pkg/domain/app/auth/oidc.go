package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

const oidcPendingSessionMinExp = 5 * time.Minute

var oidcCertWardenScope string
var oidcRequiredScopes []string

// setOIDCScopes sets the appropriate scopes based on the issuer URL
func setOIDCScopes(issuerURL string, appURI string) {
	// Default scope format for all providers
	oidcCertWardenScope = "certwarden:superadmin"

	// For Microsoft Entra ID, the scope format requires the full application URI prefix
    if strings.Contains(issuerURL, "login.microsoftonline.com") && appURI != "" {
        oidcCertWardenScope = fmt.Sprintf("%s/certwarden:superadmin", appURI)
    }

	// Set the required scopes array
	oidcRequiredScopes = []string{oidc.ScopeOpenID, oidc.ScopeOfflineAccess, "profile", oidcCertWardenScope}
}

// oidcPendingSession tracks various bits of information across the different steps of the OIDC
// autentication and authorization flow
type oidcPendingSession struct {
	callerRedirectUrl *url.URL
	codeVerifierHex   string
	createdAt         time.Time
	oauth2Token       *oauth2.Token
	oidcIDToken       *oidc.IDToken
	idpCode           string
}

// oidcErrorURL copies a URL and removes the OIDC param values from the
// returned copy and then sets the OIDC error param on the returned copy
// instead
func oidcErrorURL(u *url.URL, err error) *url.URL {
	newU, _ := url.Parse(u.String())

	queryValues := newU.Query()

	queryValues.Del("redirect_uri")
	queryValues.Del("state")
	queryValues.Del("code")

	queryValues.Set("oidc_error", url.QueryEscape(err.Error()))

	newU.RawQuery = queryValues.Encode()

	return newU
}

// oidcUnauthorizedErrorURL copies a URL and removes the OIDC param values from the
// returned copy and then sets the OIDC error param on the returned copy
// instead
func oidcUnauthorizedErrorURL(u *url.URL) *url.URL {
	err := errors.New("oidc failed (unauthorized)")
	return oidcErrorURL(u, err)
}

// expectedToken contains oauth2 Token the values used by this application
type expectedToken struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token"`
	Scope        string `json:"scope"`
}

// oidcExtraFuncs implements session manager's extraFuncs interface
type oidcExtraFuncs struct {
	ctxWithHttpClient context.Context
	cfg               *oauth2.Config
	idTokenVerifier   *oidc.IDTokenVerifier
	token             *expectedToken

	mu sync.Mutex
}

// RefreshCheck for oidc performs a token refresh with the Idp; this is done manually instead
// of with the OIDC package because that pkg doesn't appear to have a force refresh option
func (oef *oidcExtraFuncs) RefreshCheck() error {
	oef.mu.Lock()
	defer oef.mu.Unlock()

	// make OAuth2 refresh payload
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("client_id", oef.cfg.ClientID)
	data.Set("client_secret", oef.cfg.ClientSecret)
	data.Set("refresh_token", oef.token.RefreshToken)

	// make http request
	req, err := http.NewRequest("POST", oef.cfg.Endpoint.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Add("content-type", "application/x-www-form-urlencoded")

	// do request with ctx http client
	httpClient, found := oef.ctxWithHttpClient.Value(oauth2.HTTPClient).(*http.Client)
	if !found {
		return fmt.Errorf("oidc refresh failed, http client is missing")
	}
	res, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// see: https://datatracker.ietf.org/doc/html/rfc6749#section-5.1
	// success must return 200
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("oidc refresh failed, status %d", res.StatusCode)
	}

	// unmarshal the token
	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("oidc refresh failed, failed to read body (%s)", err)
	}

	var t expectedToken
	err = json.Unmarshal(bodyBytes, &t)
	if err != nil {
		return fmt.Errorf("oidc refresh failed, failed to unmarshal new token (%s)", err)
	}

	// validate token values
	if t.AccessToken == "" {
		return errors.New("oidc refresh failed, new access token empty")
	}

	if t.RefreshToken == "" {
		return errors.New("oidc refresh failed, new refresh token empty")
	}

	_, err = oef.idTokenVerifier.Verify(oef.ctxWithHttpClient, t.IDToken)
	if err != nil {
		return fmt.Errorf("oidc refresh failed, id token failed verification (%s)", err)
	}

	// Validate the required scopes were granted
	if t.Scope != "" {
		responseScopes := strings.Split(t.Scope, " ")
		for _, requiredScope := range oidcRequiredScopes {
			found := false
			for _, responseScope := range responseScopes {
				if requiredScope == responseScope {
					found = true
					break
				}
			}

			if !found {
				return fmt.Errorf("oidc refresh failed, required scope '%s' was not granted", requiredScope)
			}
		}
	}

	// set new token & return ok
	oef.token = &t
	return nil
}

// startOidcCleanerService starts a goroutine to remove pending sessions that were abandoned;
// the expiration limit is not a hard fail, it just provides an eventual backstop to purge
// pending sessions (as opposed to a hard fail if the time limit is broached by even 1 second)
func (service *Service) startOidcCleanerService(ctx context.Context, wg *sync.WaitGroup) {
	// log start and update wg
	service.logger.Info("starting oidc panding session cleaner service")

	// delete func that checks values for expired session
	deleteFunc := func(_ string, v *oidcPendingSession) bool {
		// is now past min expiration?
		return time.Now().After(v.createdAt.Add(oidcPendingSessionMinExp))
	}

	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			// wait time is based on min expiration
			delayTimer := time.NewTimer(2 * oidcPendingSessionMinExp)

			select {
			case <-ctx.Done():
				// ensure timer releases resources
				if !delayTimer.Stop() {
					<-delayTimer.C
				}

				// exit
				service.logger.Info("oidc panding session cleaner service shutdown complete")
				return

			case <-delayTimer.C:
				// continue and run
			}

			// run delete func against sessions map
			_ = service.oidc.pendingSessions.DeleteFunc(deleteFunc)
		}
	}()
}
