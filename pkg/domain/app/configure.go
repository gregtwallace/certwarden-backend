package app

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

// path to the config file
const configFile = "./config.yaml"

// config is the configuration structure for app (and subsequently services)
type config struct {
	Hostname        *string `yaml:"hostname"`
	ServeFrontend   *bool   `yaml:"serve_frontend"`
	HttpsPort       *int    `yaml:"https_port"`
	HttpPort        *int    `yaml:"http_port"`
	PrivateKeyName  *string `yaml:"private_key_name"`
	CertificateName *string `yaml:"certificate_name"`
	DevMode         *bool   `yaml:"dev_mode"`
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
		Hostname:        new(string),
		ServeFrontend:   new(bool),
		HttpsPort:       new(int),
		HttpPort:        new(int),
		PrivateKeyName:  new(string),
		CertificateName: new(string),
		DevMode:         new(bool),
	}

	// http/s server
	*cfg.Hostname = "localhost"
	*cfg.ServeFrontend = true
	*cfg.HttpsPort = 4055
	*cfg.HttpPort = 4050

	// key/cert
	*cfg.PrivateKeyName = "legocerthub"
	*cfg.CertificateName = "legocerthub"

	// dev mode
	*cfg.DevMode = false

	return cfg
}
