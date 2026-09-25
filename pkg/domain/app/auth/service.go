package auth

import (
	"certwarden-backend/pkg/datatypes/safemap"
	"certwarden-backend/pkg/domain/app/auth/session_manager"
	"certwarden-backend/pkg/output"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

var errServiceComponent = errors.New("necessary auth service component is missing")

// constant for bcrypt cost value
const BcryptCost = 12

// App interface is for connecting to the main app
type App interface {
	IsHttps() bool
	CORSPermittedCrossOrigins() []string
	FrontendURLPath() string
	APIURLPath() string
	GetHttpClient() *http.Client
	GetLogger() *zap.SugaredLogger
	GetOutputter() *output.Service
	GetAuthStorage() Storage
	GetShutdownContext() context.Context
	GetShutdownWaitGroup() *sync.WaitGroup
}

type User struct {
	ID           int
	Username     string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Storage interface {
	GetOneUserByUsername(username string) (*User, error)
	PutUserPasswordHash(username string, passwordHash string, updatedAt time.Time) (userId int, err error)
}

type Config struct {
	Local struct {
		Enabled *bool `yaml:"enabled"`
	} `yaml:"local"`
	OIDC struct {
		IssuerURL      string `yaml:"issuer_url"`
		ClientID       string `yaml:"client_id"`
		ClientSecret   string `yaml:"client_secret"`
		APIRedirectURI string `yaml:"api_redirect_uri"`
		// Scopes overrides the scopes requested from, and required of, the
		// IdP. Empty = the upstream default (which includes the custom
		// certwarden:superadmin scope). IdPs that cannot issue custom scopes
		// (e.g. Cloudflare Access) need this, and then AUTHORIZATION no
		// longer comes from that scope: use AllowedEmails and/or the IdP's
		// own access policy.
		Scopes []string `yaml:"scopes"`
		// AllowedEmails, if non-empty, restricts OIDC login to ID tokens whose
		// verified email claim is in this list (case-insensitive).
		AllowedEmails []string `yaml:"allowed_emails"`
	} `yaml:"oidc"`
}

// service struct
type Service struct {
	logger                    *zap.SugaredLogger
	output                    *output.Service
	corsPermittedCrossOrigins []string
	frontendURLPath           string
	apiURLPath                string
	sessionManager            *session_manager.SessionManager
	local                     struct {
		storage Storage
	}
	oidc struct {
		ctxWithHttpClient context.Context
		pendingSessions   *safemap.SafeMap[*oidcPendingSession]
		oauth2Config      *oauth2.Config
		idTokenVerifier   *oidc.IDTokenVerifier
		requiredScopes    []string
		allowedEmails     map[string]struct{}
	}
}

// NewService creates a new users service
func NewService(app App, cfg *Config) (*Service, error) {
	service := new(Service)

	// logger
	service.logger = app.GetLogger()
	if service.logger == nil {
		return nil, errServiceComponent
	}

	// output service
	service.output = app.GetOutputter()
	if service.output == nil {
		return nil, errServiceComponent
	}

	// other misc
	service.corsPermittedCrossOrigins = app.CORSPermittedCrossOrigins()
	service.frontendURLPath = app.FrontendURLPath()
	service.apiURLPath = app.APIURLPath()

	// create session manager
	service.sessionManager = session_manager.NewSessionManager(app.IsHttps(), len(app.CORSPermittedCrossOrigins()) > 0, service.logger)
	// start cleaner
	service.sessionManager.StartCleanerService(app.GetShutdownContext(), app.GetShutdownWaitGroup())

	// storage
	if cfg.Local.Enabled != nil && *cfg.Local.Enabled {
		service.local.storage = app.GetAuthStorage()
		if service.local.storage == nil {
			return nil, errServiceComponent
		}
	}

	// OIDC (optional)
	if cfg.OIDC.IssuerURL != "" {
		// context to use CW's http Client
		oidcCtxWithHttpClient := oidc.ClientContext(app.GetShutdownContext(), app.GetHttpClient())

		// oidc provider
		var err error
		provider, err := oidc.NewProvider(oidcCtxWithHttpClient, cfg.OIDC.IssuerURL)
		if err != nil {
			// failed to make provider, log error but continue without oidc
			service.logger.Errorf("auth: failed to create oidc provider (%s), oidc will not be enabled", err)
		} else {
			// provider created okay, finish OIDC configuration
			// store ctx for use in other OIDC/Oauth2 calls
			service.oidc.ctxWithHttpClient = oidcCtxWithHttpClient

			// manage pending OIDC states
			service.oidc.pendingSessions = safemap.NewSafeMap[*oidcPendingSession]()

			// verify the rest of the config is populated
			if cfg.OIDC.ClientID == "" || cfg.OIDC.ClientSecret == "" || cfg.OIDC.APIRedirectURI == "" {
				return nil, errors.New("auth: when using OIDC, config must speficy client id, client secret, and api redirect uri")
			}

			// oidc oauth2 config
			service.oidc.oauth2Config = &oauth2.Config{
				ClientID:     cfg.OIDC.ClientID,
				ClientSecret: cfg.OIDC.ClientSecret,
				RedirectURL:  cfg.OIDC.APIRedirectURI,

				Endpoint: provider.Endpoint(),
				Scopes:   oidcRequiredScopes,
			}

			// scope override (see Config.OIDC.Scopes)
			service.oidc.requiredScopes = oidcRequiredScopes
			if len(cfg.OIDC.Scopes) > 0 {
				service.oidc.requiredScopes = cfg.OIDC.Scopes
				service.oidc.oauth2Config.Scopes = cfg.OIDC.Scopes
				service.logger.Infof("auth: oidc scopes overridden by config: %s", strings.Join(cfg.OIDC.Scopes, " "))
			}

			// optional email allow-list (see Config.OIDC.AllowedEmails)
			if len(cfg.OIDC.AllowedEmails) > 0 {
				service.oidc.allowedEmails = make(map[string]struct{}, len(cfg.OIDC.AllowedEmails))
				for _, e := range cfg.OIDC.AllowedEmails {
					service.oidc.allowedEmails[strings.ToLower(strings.TrimSpace(e))] = struct{}{}
				}
			} else if len(cfg.OIDC.Scopes) > 0 {
				service.logger.Warn("auth: oidc scopes overridden without allowed_emails; authorization relies entirely on the idp's access policy")
			}

			// ensure redirect parses
			_, err = url.Parse(service.oidc.oauth2Config.RedirectURL)
			if err != nil {
				err = fmt.Errorf("auth: oidc cfg url failed to parse (%w), fix the config", err)
				service.logger.Error(err)
				return nil, err
			}

			// oidc id token verifier
			service.oidc.idTokenVerifier = provider.Verifier(&oidc.Config{ClientID: cfg.OIDC.ClientID})

			// clean stale pending sessions
			service.startOidcCleanerService(service.oidc.ctxWithHttpClient, app.GetShutdownWaitGroup())
		}
	}

	return service, nil
}

// make ValidateAuthHeader available to App
func (service *Service) ValidateAuthHeader(r *http.Request, w http.ResponseWriter, logTaskName string) error {
	_, err := service.sessionManager.ValidateAuthHeader(r, w, logTaskName)
	return err
}

// auth method enabled checks
func (service *Service) methodLocalEnabled() bool {
	// storage is an interface - may need to re-evaluate this
	return service.local.storage != nil
}

func (service *Service) methodOIDCEnabled() bool {
	return service.oidc.pendingSessions != nil
}
