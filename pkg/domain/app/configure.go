package app

import (
	"fmt"
	"legocerthub-backend/pkg/challenges"
	"legocerthub-backend/pkg/challenges/providers/dns01cloudflare"
	"legocerthub-backend/pkg/challenges/providers/http01internal"
	"legocerthub-backend/pkg/domain/orders"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

// path to the config file
const configFile = "./config.yaml"

// config is the configuration structure for app (and subsequently services)
type config struct {
	Hostname           *string           `yaml:"hostname"`
	HttpsPort          *int              `yaml:"https_port"`
	HttpPort           *int              `yaml:"http_port"`
	EnableHttpRedirect *bool             `yaml:"enable_http_redirect"`
	LogLevel           *string           `yaml:"log_level"`
	ServeFrontend      *bool             `yaml:"serve_frontend"`
	PrivateKeyName     *string           `yaml:"private_key_name"`
	CertificateName    *string           `yaml:"certificate_name"`
	DevMode            *bool             `yaml:"dev_mode"`
	Orders             orders.Config     `yaml:"orders"`
	Challenges         challenges.Config `yaml:"challenge_providers"`
}

// httpAddress() returns formatted http server address string
func (c config) httpDomainAndPort() string {
	return fmt.Sprintf("%s:%d", *c.Hostname, *c.HttpPort)
}

// httpsAddress() returns formatted https server address string
func (c config) httpsDomainAndPort() string {
	return fmt.Sprintf("%s:%d", *c.Hostname, *c.HttpsPort)
}

// readConfigFile parses the config yaml file. It also sets default config
// for any unspecified options
func readConfigFile() (cfg config) {
	// load default config options
	cfg = defaultConfig()

	// open config file, if exists
	file, err := os.Open(configFile)
	if err != nil {
		log.Printf("warn: config file error: %s", err)
		return cfg
	}
	defer file.Close()

	// decode config over default config
	// this will overwrite default values, but only for options that exist
	// in the config file
	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&cfg)
	if err != nil {
		log.Printf("warn: config file error: %s", err)
		return cfg
	}

	return cfg
}

// defaultConfig generates the configuration using defaults
// config.default.yaml should be updated if this func is updated
func defaultConfig() (cfg config) {
	cfg = config{
		Hostname:           new(string),
		HttpsPort:          new(int),
		HttpPort:           new(int),
		EnableHttpRedirect: new(bool),
		LogLevel:           new(string),
		ServeFrontend:      new(bool),
		PrivateKeyName:     new(string),
		CertificateName:    new(string),
		DevMode:            new(bool),
		Orders: orders.Config{
			AutomaticOrderingEnable:     new(bool),
			ValidRemainingDaysThreshold: new(int),
			RefreshTimeHour:             new(int),
			RefreshTimeMinute:           new(int),
		},
		Challenges: challenges.Config{
			Http01InternalConfig: http01internal.Config{
				Enable: new(bool),
				Port:   new(int),
			},
			Dns01CloudflareConfig: dns01cloudflare.Config{
				Enable: new(bool),
			},
		},
	}

	// set default values
	// http/s server
	*cfg.Hostname = "localhost"
	*cfg.HttpsPort = 4055
	*cfg.HttpPort = 4050

	*cfg.EnableHttpRedirect = false

	*cfg.LogLevel = defaultLogLevel.String()
	*cfg.ServeFrontend = true

	// key/cert
	*cfg.PrivateKeyName = "legocerthub"
	*cfg.CertificateName = "legocerthub"

	// dev mode
	*cfg.DevMode = false

	// orders
	*cfg.Orders.AutomaticOrderingEnable = true
	*cfg.Orders.ValidRemainingDaysThreshold = 40
	*cfg.Orders.RefreshTimeHour = 3
	*cfg.Orders.RefreshTimeMinute = 12

	// challenge providers
	// http-01-internal
	*cfg.Challenges.Http01InternalConfig.Enable = true
	*cfg.Challenges.Http01InternalConfig.Port = 4060

	// dns-01-cloudflare
	*cfg.Challenges.Dns01CloudflareConfig.Enable = false

	// end challenge providers

	return cfg
}
