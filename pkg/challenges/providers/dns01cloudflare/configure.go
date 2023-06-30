package dns01cloudflare

import (
	"context"
	"legocerthub-backend/pkg/datatypes"

	"github.com/cloudflare/cloudflare-go"
)

// Configuration options
type Config struct {
	Enable   *bool `yaml:"enable"`
	Accounts []struct {
		Email        string `yaml:"email"`
		GlobalApiKey string `yaml:"global_api_key"`
	} `yaml:"accounts"`
	// TODO: Simplify this to not be nested, however, config version number will need
	// to change as this will break configs.
	ApiTokens []struct {
		APIToken string `yaml:"api_token"`
	} `yaml:"api_tokens"`
}

// configureCloudflareAPI configures the service to use the API Tokens
// and Accounts specified within the config. If any of the config does
// not work, configuration is aborted and an error is returned.
func (service *Service) configureCloudflareAPI(config *Config) error {
	// make map to store known domain zones
	service.knownDomainZones = datatypes.NewSafeMap()

	// configure accounts
	// create API objects for each account and map all known zones
	for i := range config.Accounts {
		// make api for account
		apiInstance, err := cloudflare.New(config.Accounts[i].GlobalApiKey, config.Accounts[i].Email)
		if err != nil {
			service.logger.Error(err)
			return err
		}

		// add zones from api
		err = service.addZonesFromApiInstance(apiInstance)
		if err != nil {
			return err
		}
	}
	// configure accounts - END

	// configure domains specified by tokens
	// for each token add the domain -> api mappings to known domains
	for i := range config.ApiTokens {
		// make api for the token
		apiInstance, err := cloudflare.NewWithAPIToken(config.ApiTokens[i].APIToken)
		if err != nil {
			service.logger.Error(err)
			return err
		}

		// add zones from api
		err = service.addZonesFromApiInstance(apiInstance)
		if err != nil {
			return err
		}
	}
	// configure domains specified by tokens - END

	return nil
}

// addZonesFromApiInstance fetches available zones from a cloudflare API and
// then adds those zones to the Service available list
func (service *Service) addZonesFromApiInstance(cfApi *cloudflare.API) error {
	// fetch list of zones
	zoneList, err := cfApi.ListZones(context.Background())
	if err != nil {
		service.logger.Error(err)
		return err
	}

	// add all to the known zone list
	for i := range zoneList {
		// verify proper dns edit permission is present
		editDnsFound := false

		for j := range zoneList[i].Permissions {
			// only add the zone if the key has proper permissions to edit dns
			if zoneList[i].Permissions[j] == "#dns_records:edit" {

				editDnsFound = true

				// add zone
				z := zone{
					id:  zoneList[i].ID,
					api: cfApi,
				}
				_, _ = service.knownDomainZones.Add(zoneList[i].Name, z)
				// no error if exists -- will just use whichever token loaded in first

				// break once proper perm found
				break
			}
		}

		// log error if a zone exists without the needed permission
		if !editDnsFound {
			// redacted key depends on if this was an account or a token API instance
			if len(cfApi.APIToken) > 0 {
				service.logger.Warnf("api instance %s does not have #dns_records:edit permission for defined zone %s", redactKey(cfApi.APIToken), zoneList[i].Name)
			} else if len(cfApi.APIKey) > 0 {
				service.logger.Warnf("api instance %s does not have #dns_records:edit permission for defined zone %s", redactKey(cfApi.APIKey), zoneList[i].Name)
			} else {
				service.logger.Error("cloudflare api instance does not have an APIKey or APIToken -- this should not be possible")
			}
		}
	}

	return nil
}

// redactKey removes the middle portion of a string and returns only the first and last
// characters separated by asterisks. if the key is less than or equal to 10 chars only
// asterisks are returned
func redactKey(key string) string {
	if len(key) <= 10 {
		return "**********"
	}

	// return first 3 + asterisks + last 3
	return string(key[:3]) + "**********" + string(key[len(key)-3:])
}
