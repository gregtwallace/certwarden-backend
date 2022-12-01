package dns01cloudflare

import (
	"context"
	"errors"
	"legocerthub-backend/pkg/datatypes"

	"github.com/cloudflare/cloudflare-go"
)

var ErrDuplicateZone = errors.New("duplicate domain (zone) configuration found")

// Configuration options
type Config struct {
	Enable   *bool `yaml:"enable"`
	Accounts []struct {
		Email        string `yaml:"email"`
		GlobalApiKey string `yaml:"global_api_key"`
	} `yaml:"accounts"`
	ApiTokens []struct {
		APIToken  string   `yaml:"api_token"`
		ZoneNames []string `yaml:"zone_names"`
	} `yaml:"api_tokens"`
}

// configureCloudflareAPI configures the service to use the API Tokens
// and Accounts specified within the config. If any of the config does
// not work, configuration is aborted and an error is returned.
func (service *Service) configureCloudflareAPI(config *Config) error {
	// make map to store known domain zones
	service.knownDomainZones = datatypes.NewSafeMap()

	// configure domains specified by tokens
	// for each token add the domain -> api mappings to known domains
	for i := range config.ApiTokens {

		// make api for the token
		apiInstance, err := cloudflare.NewWithAPIToken(config.ApiTokens[i].APIToken)
		if err != nil {
			service.logger.Error(err)
			return err
		}

		// add each domain name to known, specify the associated api / ZoneID
		// and confirm the config is valid for the zone
		for j := range config.ApiTokens[i].ZoneNames {
			// test the api token by getting zone ID for the name (domain)
			zoneID, err := apiInstance.ZoneIDByName(config.ApiTokens[i].ZoneNames[j])
			if err != nil {
				service.logger.Error(err)
				return err
			}

			// add the zone to the known list
			z := zone{
				id:  zoneID,
				api: apiInstance,
			}
			exists, _ := service.knownDomainZones.Add(config.ApiTokens[i].ZoneNames[j], z)
			// if the same domain/zone is configured more than once, error
			if exists {
				err = ErrDuplicateZone
				service.logger.Error(err)
				return err
			}
		}
	}
	// configure domains specified by tokens - END

	// configure accounts
	// create API objects for each account and map all known zones
	for i := range config.Accounts {
		// configure API
		apiInstance, err := cloudflare.New(config.Accounts[i].GlobalApiKey, config.Accounts[i].Email)
		if err != nil {
			service.logger.Error(err)
			return err
		}

		// fetch list of zones
		zoneList, err := apiInstance.ListZones(context.Background())
		if err != nil {
			service.logger.Error(err)
			return err
		}

		// add all to the known zone list
		for j := range zoneList {
			z := zone{
				id:  zoneList[j].ID,
				api: apiInstance,
			}
			_, _ = service.knownDomainZones.Add(zoneList[j].Name, z)
			// no error if exists, allow overlap with tokens
		}
	}
	// configure accounts - END

	return nil
}
