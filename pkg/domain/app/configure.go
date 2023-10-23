package app

import (
	"errors"
	"fmt"
	"io"
	"legocerthub-backend/pkg/challenges"
	"legocerthub-backend/pkg/challenges/dns_checker"
	"legocerthub-backend/pkg/challenges/providers/http01internal"
	"legocerthub-backend/pkg/domain/app/updater"
	"legocerthub-backend/pkg/domain/orders"
	"os"

	"gopkg.in/yaml.v3"
)

// path to the config file
const configFilePath = dataStoragePath + "/config.yaml"

func (app *Application) GetConfigFilename() string {
	return configFilePath
}

// config is the configuration structure for app (and subsequently services)
type config struct {
	ConfigVersion             *int              `yaml:"config_version"`
	BindAddress               *string           `yaml:"bind_address"`
	HttpsPort                 *int              `yaml:"https_port"`
	HttpPort                  *int              `yaml:"http_port"`
	EnableHttpRedirect        *bool             `yaml:"enable_http_redirect"`
	ServeFrontend             *bool             `yaml:"serve_frontend"`
	CORSPermittedCrossOrigins []string          `yaml:"cors_permitted_crossorigins"`
	CertificateName           *string           `yaml:"certificate_name"`
	DisableHSTS               *bool             `yaml:"disable_hsts"`
	LogLevel                  *string           `yaml:"log_level"`
	EnablePprof               *bool             `yaml:"enable_pprof"`
	PprofPort                 *int              `yaml:"pprof_port"`
	Updater                   updater.Config    `yaml:"updater"`
	Orders                    orders.Config     `yaml:"orders"`
	Challenges                challenges.Config `yaml:"challenges"`
}

// httpAddress() returns formatted http server address string
func (c config) httpServAddress() string {
	return fmt.Sprintf("%s:%d", *c.BindAddress, *c.HttpPort)
}

// httpsAddress() returns formatted https server address string
func (c config) httpsServAddress() string {
	return fmt.Sprintf("%s:%d", *c.BindAddress, *c.HttpsPort)
}

// pprofAddress() returns formatted pprov server address string
func (c config) pprofServAddress() string {
	return fmt.Sprintf("%s:%d", *c.BindAddress, *c.PprofPort)
}

// loadConfigFile parses the config yaml file. It also sets default config
// for any unspecified options. If there is no config file, a blank one with
// the current config version is created. It will also try to upgrade old
// config schemas, if possible.
func (app *Application) loadConfigFile() (err error) {
	// check if config file exists
	if _, err := os.Stat(configFilePath); errors.Is(err, os.ErrNotExist) {
		app.logger.Warn("LeGo config file does not exist, creating one")
		// create config file
		cfgFile, err := os.Create(configFilePath)
		if err != nil {
			return fmt.Errorf("failed to create new LeGo config file (%s)", err)
		}

		// write new config with config version
		newCfgFile := fmt.Sprintf("\"config_version\": %d\n", appConfigVersion)
		cfgFile.WriteString(newCfgFile)

		cfgFile.Close()
	}

	// open config file
	cfgFile, err := os.Open(configFilePath)
	if err != nil {
		return fmt.Errorf("failed to open config file (%s)", err)
	}
	defer cfgFile.Close()

	// read in config file
	cfgFileData, err := io.ReadAll(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to read config file (%s)", err)
	}

	// unmarshal into yaml object
	cfgFileYamlObj := make(map[string]any)
	err = yaml.Unmarshal(cfgFileData, cfgFileYamlObj)
	if err != nil {
		return fmt.Errorf("failed to parse config file for version migration (%s)", err)
	}

	// get current config version
	cfgVerVal, ok := cfgFileYamlObj["config_version"]
	if !ok {
		return fmt.Errorf("config version is missing (expected %d); fix the config file", appConfigVersion)
	}
	origCfgVer, ok := cfgVerVal.(int)
	if !ok {
		return fmt.Errorf("config version is not an integer (%s); fix the config file", cfgVerVal)
	}

	// check config version, do auto schema upgrades if possible
	cfgVer := origCfgVer
	for cfgVer != appConfigVersion && err == nil {
		// take incremental migration action, if possible
		switch cfgVer {
		case 1:
			cfgVer, err = configMigrateV1toV2(cfgFileYamlObj)

		case appConfigVersion:
			// no-op, loop will end due to version ==

		default:
			err = fmt.Errorf("config version is %d (expected %d) and cannot be fixed automatically; fix the config file", *app.config.ConfigVersion, appConfigVersion)
		}
	}
	// err check from upgrade for loop
	if err != nil {
		return err
	}

	// if cfg version changed, write config
	if cfgVer != origCfgVer {
		// update config bytes with new cfg
		cfgFileData, err = yaml.Marshal(cfgFileYamlObj)
		if err != nil {
			return fmt.Errorf("failed to marshal new config file for version migration (%s)", err)
		}

		// write new config
		err = os.WriteFile(configFilePath, cfgFileData, 0600)
		if err != nil {
			return fmt.Errorf("could not write version migrated config file (%s)", err)
		}
		app.logger.Infof("config version migrated from %d to %d", origCfgVer, cfgVer)
	} else {
		app.logger.Debugf("config version is current (%d)", appConfigVersion)
	}

	// decode config
	app.config = new(config)
	err = yaml.Unmarshal(cfgFileData, app.config)
	if err != nil {
		return fmt.Errorf("failed to unmarshal config file into lego config struct (%s)", err)
	}

	// set defaults on anything that wasn't specified
	app.setDefaultConfigValues()

	// success
	return nil
}

// setDefaultConfigValues checks each field of the config that has a default
// value and if the value is not set it sets the default
func (app *Application) setDefaultConfigValues() {
	if app.config == nil {
		app.config = new(config)
	}

	// no default config version

	// http/s server
	if app.config.BindAddress == nil {
		app.config.BindAddress = new(string)
		*app.config.BindAddress = ""
	}
	if app.config.HttpsPort == nil {
		app.config.HttpsPort = new(int)
		*app.config.HttpsPort = 4055
	}
	if app.config.HttpPort == nil {
		app.config.HttpPort = new(int)
		*app.config.HttpPort = 4050
	}
	if app.config.EnableHttpRedirect == nil {
		app.config.EnableHttpRedirect = new(bool)
		*app.config.EnableHttpRedirect = true
	}
	if app.config.ServeFrontend == nil {
		app.config.ServeFrontend = new(bool)
		*app.config.ServeFrontend = true
	}
	if app.config.CertificateName == nil {
		app.config.CertificateName = new(string)
		*app.config.CertificateName = "legocerthub"
	}
	if app.config.DisableHSTS == nil {
		app.config.DisableHSTS = new(bool)
		*app.config.DisableHSTS = false
	}

	// debug and dev stuff
	if app.config.LogLevel == nil {
		app.config.LogLevel = new(string)
		*app.config.LogLevel = defaultLogLevel.String()
	}
	if app.config.EnablePprof == nil {
		app.config.EnablePprof = new(bool)
		*app.config.EnablePprof = false
	}
	if app.config.PprofPort == nil {
		app.config.PprofPort = new(int)
		*app.config.PprofPort = 4065
	}

	// updater
	if app.config.Updater.AutoCheck == nil {
		app.config.Updater.AutoCheck = new(bool)
		*app.config.Updater.AutoCheck = true
	}
	if app.config.Updater.Channel == nil {
		app.config.Updater.Channel = new(updater.Channel)
		*app.config.Updater.Channel = updater.ChannelBeta
	}

	// orders
	if app.config.Orders.AutomaticOrderingEnable == nil {
		app.config.Orders.AutomaticOrderingEnable = new(bool)
		*app.config.Orders.AutomaticOrderingEnable = true
	}
	if app.config.Orders.ValidRemainingDaysThreshold == nil {
		app.config.Orders.ValidRemainingDaysThreshold = new(int)
		*app.config.Orders.ValidRemainingDaysThreshold = 40
	}
	if app.config.Orders.RefreshTimeHour == nil {
		app.config.Orders.RefreshTimeHour = new(int)
		*app.config.Orders.RefreshTimeHour = 3
	}
	if app.config.Orders.RefreshTimeMinute == nil {
		app.config.Orders.RefreshTimeMinute = new(int)
		*app.config.Orders.RefreshTimeMinute = 12
	}

	// challenge dns checker services
	if app.config.Challenges.DnsCheckerConfig.DnsServices == nil || len(app.config.Challenges.DnsCheckerConfig.DnsServices) <= 0 {
		app.config.Challenges.DnsCheckerConfig.DnsServices = []dns_checker.DnsServiceIPPair{
			// Cloudflare
			{
				Primary:   "1.1.1.1",
				Secondary: "1.0.0.1",
			},
			// Quad9
			{
				Primary:   "9.9.9.9",
				Secondary: "149.112.112.112",
			},
			// Google
			{
				Primary:   "8.8.8.8",
				Secondary: "8.8.4.4",
			},
		}
	}

	// challenge provider
	if app.config.Challenges.ProviderConfigs.Len() <= 0 {
		http01Port := new(int)
		*http01Port = 4060

		app.config.Challenges.ProviderConfigs.Http01InternalConfigs = []*http01internal.Config{{
			Doms: []string{"*"},
			Port: http01Port,
		}}
	}
}
