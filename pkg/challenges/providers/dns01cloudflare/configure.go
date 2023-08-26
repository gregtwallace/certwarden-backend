package dns01cloudflare

import (
	"context"
	"errors"
	"fmt"

	"github.com/cloudflare/cloudflare-go"
)

var (
	errAccountAndTokenSpecified = errors.New("cloudflare provider config should have either an account or an api token, not both")
	errMissingConfigInfo        = errors.New("cloudflare config missing an account (email and global key) or api token")
)

// Configuration options for an instance of Cloudflare provider
type Config struct {
	Domains []string `yaml:"domains"`
	// Account
	Account struct {
		Email        *string `yaml:"email"`
		GlobalApiKey *string `yaml:"global_api_key"`
	} `yaml:"account"`
	// -- OR --
	// Token
	ApiToken *string `yaml:"api_token"`
}

// zoneValid checks for the proper Cloudflare permission to edit dns on the
// specified zone
func zoneValid(z *cloudflare.Zone) bool {
	for _, permission := range z.Permissions {
		if permission == "#dns_records:edit" {
			return true
		}
	}
	return false
}

// configureCloudflareAPI configures the service to use the API Tokens
// and Accounts specified within the config. If any of the config does
// not work, configuration is aborted and an error is returned.
func (service *Service) configureCloudflareAPI(cfg *Config) (err error) {
	// if both account and token are specified, error
	if (cfg.Account.Email != nil || cfg.Account.GlobalApiKey != nil) && cfg.ApiToken != nil {
		return errAccountAndTokenSpecified
	}

	// if using apiToken
	if cfg.ApiToken != nil {
		// make api for the token
		service.cloudflareApi, err = cloudflare.NewWithAPIToken(*cfg.ApiToken, service.httpClient.AsCloudflareOptions()...)
		if err != nil {
			err = fmt.Errorf("failed to create api instance %s (%s)", redactIdentifier(*cfg.ApiToken), err)
			service.logger.Error(err)
			return err
		}

	} else if cfg.Account.Email != nil && cfg.Account.GlobalApiKey != nil {
		// else if using Account
		service.cloudflareApi, err = cloudflare.New(*cfg.Account.GlobalApiKey, *cfg.Account.Email, service.httpClient.AsCloudflareOptions()...)
		if err != nil {
			err = fmt.Errorf("failed to create api instance %s - %s (%s)", redactIdentifier(*cfg.Account.GlobalApiKey), *cfg.Account.Email, err)
			service.logger.Error(err)
			return err
		}

	} else {
		// else incomplete config
		return errMissingConfigInfo
	}

	// fetch list of zones
	availableZones, err := service.cloudflareApi.ListZones(context.Background())
	if err != nil {
		err = fmt.Errorf("api instance %s failed to list zones (%s)", service.redactedApiIdentifier(), err)
		service.logger.Error(err)
		return err
	}

	// add all available zones, even if not being used (configured in cfg.Domains)
	allZoneNames := []string{}
	for i := range availableZones {
		// verify proper permission
		if zoneValid(&availableZones[i]) {
			allZoneNames = append(allZoneNames, availableZones[i].Name)
			service.domainIDs[availableZones[i].Name] = availableZones[i].ID
		}
	}
	service.logger.Debugf("cloudflare instance %s all available zones: %s", service.redactedApiIdentifier(), allZoneNames)

	// verify all domains in cfg.Domains are available zones (if not wildcard provider)
	if !(len(cfg.Domains) == 1 && cfg.Domains[0] == "*") {
		for i := range cfg.Domains {
			found := false
			for serviceDomain := range service.domainIDs {
				if cfg.Domains[i] == serviceDomain {
					found = true
					break
				}
			}
			if !found {
				// if wildcard domain, error is different
				if cfg.Domains[i] == "*" {
					return errors.New("when using wildcard domain * it must be the only specified domain on the provider")
				}
				return fmt.Errorf("cloudflare domain %s is either not available or missing the proper permission using the specified api credential", cfg.Domains[i])
			}
		}
	}

	return nil
}

// redactIdentifier removes the middle portion of a string and returns only the first and last
// characters separated by asterisks. if the key is less than or equal to 12 chars only
// asterisks are returned
func redactIdentifier(id string) string {
	// if the identifier is less than 12 chars in length, return fully redacted
	// this should never happen but just in case to prevent credential logging
	if len(id) <= 12 {
		return "************"
	}

	// return first 3 + asterisks + last 3
	return string(id[:3]) + "************" + string(id[len(id)-3:])
}

// redactedIdentifier selects either the APIKey, APIUserServiceKey, or APIToken
// (depending on which is in use for the API instance) and then redacts it to return
// the first and last characters of the key separated with asterisks. This is useful
// for logging issues without saving the full credential to logs.
func (service *Service) redactedApiIdentifier() string {
	identifier := ""

	// select whichever is present
	if len(service.cloudflareApi.APIToken) > 0 {
		identifier = service.cloudflareApi.APIToken
	} else if len(service.cloudflareApi.APIKey) > 0 {
		identifier = service.cloudflareApi.APIKey
	} else if len(service.cloudflareApi.APIUserServiceKey) > 0 {
		identifier = service.cloudflareApi.APIUserServiceKey
	} else {
		// none present, return unknown
		return "unknown"
	}

	// return redacted
	return redactIdentifier(identifier)
}
