package dns01cloudflare

import (
	"certwarden-backend/pkg/output"
	"errors"
	"fmt"
	"time"

	"github.com/cloudflare/cloudflare-go"
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

// redactedIdentifier selects the correct identifier field and then returns the identifier
// in its redacted form
func (cfg *Config) redactedIdentifier() string {
	// if token specified
	if cfg.ApiToken != nil {
		return output.RedactString(*cfg.ApiToken)
	}

	// if global api key
	if cfg.Account.GlobalApiKey != nil {
		id := output.RedactString(*cfg.Account.GlobalApiKey)
		if cfg.Account.Email != nil {
			id = id + " - " + *cfg.Account.Email
		}
		return id
	}

	return "unknown"
}

// configureCloudflareAPI configures the service to use the API Tokens
// and Accounts specified within the config.
func (service *Service) configureCloudflareAPI(cfg *Config) (err error) {
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

	// if using apiToken
	if cfg.ApiToken != nil {
		// make api for the token
		service.cloudflareApi, err = cloudflare.NewWithAPIToken(*cfg.ApiToken, service.httpClient.AsCloudflareOptions()...)
		// defer to common err check
	} else if cfg.Account != nil && cfg.Account.Email != nil && cfg.Account.GlobalApiKey != nil {
		// else if using Account
		service.cloudflareApi, err = cloudflare.New(*cfg.Account.GlobalApiKey, *cfg.Account.Email, service.httpClient.AsCloudflareOptions()...)
		// defer to common err check
	} else {
		// else incomplete config
		return errMissingConfigInfo
	}

	// common err check
	if err != nil {
		err = fmt.Errorf("failed to create api instance %s (%s)", cfg.redactedIdentifier(), err)
		service.logger.Error(err)
		return err
	}

	return nil
}
