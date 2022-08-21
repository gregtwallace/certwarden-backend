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
	Hostname *string `yaml:"hostname"`
	Port     *int    `yaml:"port"`
	DevMode  *bool   `yaml:"dev_mode"`
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
		Hostname: new(string),
		Port:     new(int),
		DevMode:  new(bool),
	}

	// http server
	*cfg.Hostname = "localhost"
	*cfg.Port = 4050

	// dev mode
	*cfg.DevMode = false

	return cfg
}
