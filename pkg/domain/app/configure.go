package app

import (
	"fmt"
	"legocerthub-backend/pkg/challenges"
	"legocerthub-backend/pkg/challenges/dns_checker"
	"legocerthub-backend/pkg/challenges/providers/http01internal"
	"legocerthub-backend/pkg/domain/app/updater"
	"legocerthub-backend/pkg/domain/orders"
	"os"

	"gopkg.in/yaml.v3"
)

// path to the config file
const configFile = dataStoragePath + "/config.yaml"

func (app *Application) GetConfigFilename() string {
	return configFile
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

// readConfigFile parses the config yaml file. It also sets default config
// for any unspecified options
func (app *Application) readConfigFile() (err error) {
	// open config file, if exists
	file, err := os.Open(configFile)
	if err != nil {
		app.logger.Warnf("can't open config file, using defaults (%s)", err)
		app.setDefaultConfigValues()
		return nil
	}
	// only needed if file actually opened
	defer file.Close()

	// decode config
	app.config = new(config)
	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(app.config)
	if err != nil {
		app.logger.Errorf("failed to read config file (%s)", err)
		return err
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

	// default config version is always invalid to ensure error if doesn't
	// exist in config file
	if app.config.ConfigVersion == nil {
		app.config.ConfigVersion = new(int)
		*app.config.ConfigVersion = -1
	}

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
