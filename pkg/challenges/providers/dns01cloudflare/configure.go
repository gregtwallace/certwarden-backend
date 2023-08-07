package dns01cloudflare

import (
	"context"
	"fmt"

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
	service.knownDomainZones = make(map[string]zone)

	// configure accounts
	// create API objects for each account and map all known zones
	for i := range config.Accounts {
		// make api for account
		apiInstance, err := cloudflare.New(config.Accounts[i].GlobalApiKey, config.Accounts[i].Email, service.httpClient.AsCloudflareOptions()...)
		if err != nil {
			err = fmt.Errorf("failed to create api instance %s (%s)", redactIdentifier(config.Accounts[i].GlobalApiKey), err)
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
		apiInstance, err := cloudflare.NewWithAPIToken(config.ApiTokens[i].APIToken, service.httpClient.AsCloudflareOptions()...)
		if err != nil {
			err = fmt.Errorf("failed to create api instance %s (%s)", redactIdentifier(config.Accounts[i].GlobalApiKey), err)
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
		err = fmt.Errorf("api instance %s failed to list zones (%s)", redactedApiIdentifier(cfApi), err)
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

				// may get overwritten if multiple instances support same zone, but doesn't matter
				service.knownDomainZones[zoneList[i].Name] = z

				// break once proper perm found
				break
			}
		}

		// log error if a zone exists without the needed permission
		if !editDnsFound {
			service.logger.Warnf("api instance %s does not have #dns_records:edit permission for defined zone %s", redactedApiIdentifier(cfApi), zoneList[i].Name)
		}
	}

	return nil
}

// redactedIdentifier selects either the APIKey, APIUserServiceKey, or APIToken
// (depending on which is in use for the API instance) and then redacts it to return
// the first and last characters of the key separated with asterisks. This is useful
// for logging issues without saving the full credential to logs.
func redactedApiIdentifier(cfApi *cloudflare.API) string {
	identifier := ""

	// select whichever is present
	if len(cfApi.APIToken) > 0 {
		identifier = cfApi.APIToken
	} else if len(cfApi.APIKey) > 0 {
		identifier = cfApi.APIKey
	} else if len(cfApi.APIUserServiceKey) > 0 {
		identifier = cfApi.APIUserServiceKey
	} else {
		// none present, return unknown
		return "unknown"
	}

	// return redacted
	return redactIdentifier(identifier)
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
