package app

import (
	"fmt"
	"legocerthub-backend/pkg/challenges"
	"legocerthub-backend/pkg/challenges/dns_checker"
	"legocerthub-backend/pkg/challenges/providers/dns01acmedns"
	"legocerthub-backend/pkg/challenges/providers/dns01acmesh"
	"legocerthub-backend/pkg/challenges/providers/dns01cloudflare"
	"legocerthub-backend/pkg/challenges/providers/dns01manual"
	"legocerthub-backend/pkg/challenges/providers/http01internal"
	"legocerthub-backend/pkg/domain/app/updater"
	"legocerthub-backend/pkg/domain/orders"
	"log"
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
	PrivateKeyName       *string           `yaml:"private_key_name"`
	CertificateName      *string           `yaml:"certificate_name"`
	DevMode              *bool             `yaml:"dev_mode"`
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

// readConfigFile parses the config yaml file. It also sets default config
// for any unspecified options
func readConfigFile() (cfg *config, err error) {
	// load default config options
	cfg = defaultConfig()

	// open config file, if exists
	file, err := os.Open(configFile)
	if err != nil {
		log.Printf("warn: can't open config file, using defaults (%s)", err)
		return cfg, nil
	}
	defer file.Close()

	// decode config over default config
	// this will overwrite default values, but only for options that exist
	// in the config file
	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
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
		PrivateKeyName:     new(string),
		CertificateName:    new(string),
		DevMode:            new(bool),
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
		Challenges: challenges.Config{
			DnsCheckerConfig: dns_checker.Config{
				// skip_check_wait_seconds defaults to nil
				// servers are a slice, no need to call new()
			},
			ProviderConfigs: challenges.ConfigProviders{
				Http01InternalConfig: http01internal.Config{
					Enable: new(bool),
					Port:   new(int),
				},
				Dns01ManualConfig: dns01manual.Config{
					Enable: new(bool),
					// script paths don't have a default
				},
				Dns01AcmeDnsConfig: dns01acmedns.Config{
					Enable:      new(bool),
					HostAddress: new(string),
				},
				Dns01AcmeShConfig: dns01acmesh.Config{
					Enable: new(bool),
				},
				Dns01CloudflareConfig: dns01cloudflare.Config{
					Enable: new(bool),
				},
			},
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
	cfg.CORSPermittedOrigins = []string{}

	// key/cert
	*cfg.PrivateKeyName = "legocerthub"
	*cfg.CertificateName = "legocerthub"

	// dev mode
	*cfg.DevMode = false

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

	// challenge providers
	// http-01-internal
	*cfg.Challenges.ProviderConfigs.Http01InternalConfig.Enable = true
	*cfg.Challenges.ProviderConfigs.Http01InternalConfig.Port = 4060

	// dns-01-manual
	*cfg.Challenges.ProviderConfigs.Dns01ManualConfig.Enable = false

	// dns-01-cloudflare
	*cfg.Challenges.ProviderConfigs.Dns01CloudflareConfig.Enable = false

	// dns-01-acmedns
	*cfg.Challenges.ProviderConfigs.Dns01AcmeDnsConfig.Enable = false

	// end challenge providers

	return cfg
}
