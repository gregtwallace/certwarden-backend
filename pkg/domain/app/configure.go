package app

import (
	"fmt"
	"legocerthub-backend/pkg/challenges"
	"legocerthub-backend/pkg/challenges/dns_checker"
	"legocerthub-backend/pkg/domain/app/updater"
	"legocerthub-backend/pkg/domain/orders"
	"os"

	"gopkg.in/yaml.v3"
)

// path to the config file
const configFile = dataStoragePath + "/config.yaml"

// config is the configuration structure for app (and subsequently services)
type config struct {
	ConfigVersion        int               `yaml:"config_version"`
	BindAddress          *string           `yaml:"bind_address"`
	HttpsPort            *int              `yaml:"https_port"`
	HttpPort             *int              `yaml:"http_port"`
	EnableHttpRedirect   *bool             `yaml:"enable_http_redirect"`
	LogLevel             *string           `yaml:"log_level"`
	ServeFrontend        *bool             `yaml:"serve_frontend"`
	CORSPermittedOrigins []string          `yaml:"cors_permitted_origins"`
	CertificateName      *string           `yaml:"certificate_name"`
	DevMode              *bool             `yaml:"dev_mode"`
	EnablePprof          *bool             `yaml:"enable_pprof"`
	PprofPort            *int              `yaml:"pprof_port"`
	Updater              updater.Config    `yaml:"updater"`
	Orders               orders.Config     `yaml:"orders"`
	Challenges           challenges.Config `yaml:"challenges"`
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
	// load default config options
	app.config = defaultConfig()

	// open config file, if exists
	file, err := os.Open(configFile)
	if err != nil {
		app.logger.Warnf("can't open config file, using defaults (%s)", err)
		return nil
	}
	// only needed if file actually opened
	defer file.Close()

	// decode config over default config
	// this will overwrite default values, but only for options that exist
	// in the config file
	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(app.config)
	if err != nil {
		return err
	}

	// success
	return nil
}

// defaultConfig generates the configuration using defaults
// config.default.yaml should be updated if this func is updated
func defaultConfig() (cfg *config) {
	cfg = &config{
		BindAddress:        new(string),
		HttpsPort:          new(int),
		HttpPort:           new(int),
		EnableHttpRedirect: new(bool),
		LogLevel:           new(string),
		ServeFrontend:      new(bool),
		CertificateName:    new(string),
		DevMode:            new(bool),
		EnablePprof:        new(bool),
		PprofPort:          new(int),
		Updater: updater.Config{
			AutoCheck: new(bool),
			Channel:   new(updater.Channel),
		},
		Orders: orders.Config{
			AutomaticOrderingEnable:     new(bool),
			ValidRemainingDaysThreshold: new(int),
			RefreshTimeHour:             new(int),
			RefreshTimeMinute:           new(int),
		},
	}

	// set default values
	// default config version is always invalid to ensure error if doesn't
	// exist in config file
	cfg.ConfigVersion = -1

	// http/s server
	*cfg.BindAddress = ""
	*cfg.HttpsPort = 4055
	*cfg.HttpPort = 4050

	*cfg.EnableHttpRedirect = true
	*cfg.LogLevel = defaultLogLevel.String()
	*cfg.ServeFrontend = true

	// LeGo https certificate name
	*cfg.CertificateName = "legocerthub"

	// dev mode
	*cfg.DevMode = false
	*cfg.EnablePprof = false
	*cfg.PprofPort = 4065

	// updater
	*cfg.Updater.AutoCheck = true
	*cfg.Updater.Channel = updater.ChannelBeta

	// orders
	*cfg.Orders.AutomaticOrderingEnable = true
	*cfg.Orders.ValidRemainingDaysThreshold = 40
	*cfg.Orders.RefreshTimeHour = 3
	*cfg.Orders.RefreshTimeMinute = 12

	// challenge dns checker services
	cfg.Challenges.DnsCheckerConfig.DnsServices = []dns_checker.DnsServiceIPPair{
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

	return cfg
}
