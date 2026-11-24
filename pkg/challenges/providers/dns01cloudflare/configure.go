package dns01cloudflare

import (
	"errors"
	"time"

	"github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/option"
)

var (
	errAccountAndTokenSpecified = errors.New("cloudflare provider config should have either an account or an api token, not both")
	errMissingConfigInfo        = errors.New("cloudflare config missing an account (email and global key) or api token")
)

// timeout for api calls
const apiCallTimeout = 10 * time.Second

// Configuration options for an instance of Cloudflare provider
type Config struct {
	// Account
	Account *struct {
		Email        *string `yaml:"email" json:"email"`
		GlobalApiKey *string `yaml:"global_api_key" json:"global_api_key"`
	} `yaml:"account,omitempty" json:"account,omitempty"`
	// -- OR --
	// Token
	ApiToken *string `yaml:"api_token,omitempty" json:"api_token,omitempty"`
}

// configureCloudflareAPI configures the service to use the API Tokens
// and Accounts specified within the config.
func (service *Service) configureCloudflareClient(cfg *Config) (err error) {
	// if blank value, change to nil pointer (treat as omitted)
	if cfg.Account != nil && ((cfg.Account.Email != nil && *cfg.Account.Email == "") || (cfg.Account.GlobalApiKey != nil && *cfg.Account.GlobalApiKey == "")) {
		cfg.Account.Email = nil
		cfg.Account.GlobalApiKey = nil
	}
	if cfg.ApiToken != nil && *cfg.ApiToken == "" {
		cfg.ApiToken = nil
	}

	// if both account and token are specified, error
	if cfg.Account != nil && cfg.ApiToken != nil {
		return errAccountAndTokenSpecified
	}

	// make option to use the custom http.Client
	opts := []option.RequestOption{option.WithHTTPClient(service.httpClient)}

	// if using apiToken
	if cfg.ApiToken != nil {
		// make api for the token
		opts = append(opts, option.WithAPIToken(*cfg.ApiToken))
	} else if cfg.Account != nil && cfg.Account.Email != nil && cfg.Account.GlobalApiKey != nil {
		// else if using Account
		opts = append(opts, option.WithAPIEmail(*cfg.Account.Email))
		opts = append(opts, option.WithAPIKey(*cfg.Account.GlobalApiKey))
	} else {
		// else incomplete config
		return errMissingConfigInfo
	}

	service.cloudflareClient = cloudflare.NewClient(opts...)

	return nil
}
