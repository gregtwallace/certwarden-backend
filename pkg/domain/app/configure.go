package app

import (
	"certwarden-backend/pkg/challenges"
	"certwarden-backend/pkg/challenges/dns_checker"
	"certwarden-backend/pkg/challenges/providers"
	"certwarden-backend/pkg/challenges/providers/http01internal"
	"certwarden-backend/pkg/domain/app/backup"
	"certwarden-backend/pkg/domain/app/updater"
	"certwarden-backend/pkg/domain/orders"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"

	"gopkg.in/yaml.v3"
)

// path to the config file
const configFile = "config.yaml"
const configFilenameWithPath = dataStorageAppDataPath + "/" + configFile

const configFolderMode = 0700
const configFileMode = 0600

func (app *Application) GetConfigFilenameWithPath() string {
	return configFilenameWithPath
}

func (app *Application) GetFileBackupFolderMode() fs.FileMode {
	return configFolderMode
}

// config is the configuration structure for app (and subsequently services)
type config struct {
	ConfigVersion             *int              `yaml:"config_version"`
	BindAddress               *string           `yaml:"bind_address"`
	HttpsPort                 *int              `yaml:"https_port"`
	HttpPort                  *int              `yaml:"http_port"`
	EnableHttpRedirect        *bool             `yaml:"enable_http_redirect"`
	FrontendServe             *bool             `yaml:"serve_frontend"`
	FrontendShowDebugInfo     *bool             `yaml:"frontend_show_debug_info"`
	CORSPermittedCrossOrigins []string          `yaml:"cors_permitted_crossorigins"`
	CertificateName           *string           `yaml:"certificate_name"`
	DisableHSTS               *bool             `yaml:"disable_hsts"`
	LogLevel                  *string           `yaml:"log_level"`
	EnablePprof               *bool             `yaml:"enable_pprof"`
	PprofHttpsPort            *int              `yaml:"pprof_https_port"`
	PprofHttpPort             *int              `yaml:"pprof_http_port"`
	Backup                    backup.Config     `yaml:"backup"`
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

// pprofHttpServAddress() returns formatted pprof http server address string
func (c config) pprofHttpServAddress() string {
	return fmt.Sprintf("%s:%d", *c.BindAddress, *c.PprofHttpPort)
}

// pprofHttpsServAddress() returns formatted pprof https server address string
func (c config) pprofHttpsServAddress() string {
	return fmt.Sprintf("%s:%d", *c.BindAddress, *c.PprofHttpsPort)
}

// loadConfigFile parses the config yaml file. It also sets default config
// for any unspecified options. If there is no config file, a blank one with
// the current config version is created. It will also try to upgrade old
// config schemas, if possible.
func (app *Application) loadConfigFile() (err error) {
	// check if config file exists
	if _, err := os.Stat(configFilenameWithPath); errors.Is(err, os.ErrNotExist) {
		// if doesn't exist, check old location and move if it exists in old location
		if _, err := os.Stat(dataStorageRootPath + "/" + configFile); err == nil {
			// exists at old location, move it
			err = os.Rename(dataStorageRootPath+"/"+configFile, configFilenameWithPath)
			if err != nil {
				app.logger.Errorf("failed to move config file from old location to new location (%s)", err)
				return err
			}
			app.logger.Infof("config file moved from %s to %s", dataStorageRootPath+"/"+configFile, configFilenameWithPath)
		} else {
			// config doesn't exist at old location either
			app.logger.Warn("config file does not exist, creating one")

			// new config file content
			newCfgFile := fmt.Sprintf("\"config_version\": %d\n", appConfigVersion)

			// write file
			err := os.WriteFile(configFilenameWithPath, []byte(newCfgFile), configFileMode)
			if err != nil {
				return fmt.Errorf("failed to create new config file (%s)", err)
			}
		}
	}
	// ignore any other Stat error, should error out below when opening

	// open config file
	cfgFile, err := os.Open(configFilenameWithPath)
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

	// upgrade if schema 1
	if cfgVer == 1 {
		cfgVer, err = configMigrateV1toV2(cfgFileYamlObj)
		if err != nil {
			return err
		}
	}

	// upgrade if schema 2
	if cfgVer == 2 {
		cfgVer, err = configMigrateV2toV3(cfgFileYamlObj)
		if err != nil {
			return err
		}
	}

	// fail if still not correct
	if cfgVer != appConfigVersion {
		return fmt.Errorf("config schema version is %d (expected %d) and cannot be fixed automatically; fix the config file", cfgVer, appConfigVersion)
	}

	// if cfg version changed, backup old config and write config
	if cfgVer != origCfgVer {
		// backup
		err = app.CreateBackupOnDisk()
		if err != nil {
			return fmt.Errorf("failed to backup data before writing config schema migration (%s)", err)
		}

		// update config bytes with new cfg
		cfgFileData, err = yaml.Marshal(cfgFileYamlObj)
		if err != nil {
			return fmt.Errorf("failed to marshal new config file for schema version migration (%s)", err)
		}

		// write new config
		err = os.WriteFile(configFilenameWithPath, cfgFileData, configFileMode)
		if err != nil {
			return fmt.Errorf("could not write schema version migrated config file (%s)", err)
		}
		app.logger.Infof("config schema version migrated from %d to %d", origCfgVer, cfgVer)
	} else {
		app.logger.Debugf("config schema version is current (%d)", appConfigVersion)
	}

	// decode config
	app.config = new(config)
	err = yaml.Unmarshal(cfgFileData, app.config)
	if err != nil {
		return fmt.Errorf("failed to unmarshal config file into config struct (%s)", err)
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
	if app.config.FrontendServe == nil {
		app.config.FrontendServe = new(bool)
		*app.config.FrontendServe = true
	}
	if app.config.FrontendShowDebugInfo == nil {
		app.config.FrontendShowDebugInfo = new(bool)
		*app.config.FrontendShowDebugInfo = false
	}
	if app.config.CertificateName == nil {
		app.config.CertificateName = new(string)
		*app.config.CertificateName = "serverdefault"
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
	if app.config.PprofHttpPort == nil {
		app.config.PprofHttpPort = new(int)
		*app.config.PprofHttpPort = 4065
	}
	if app.config.PprofHttpsPort == nil {
		app.config.PprofHttpsPort = new(int)
		*app.config.PprofHttpsPort = 4070
	}

	// backup
	if app.config.Backup.Enabled == nil {
		app.config.Backup.Enabled = new(bool)
		*app.config.Backup.Enabled = backup.DefaultBackupEnabled
	}
	if app.config.Backup.IntervalDays == nil {
		app.config.Backup.IntervalDays = new(int)
		*app.config.Backup.IntervalDays = backup.DefaultBackupDays
	}
	if app.config.Backup.Retention.MaxDays == nil {
		app.config.Backup.Retention.MaxDays = new(int)
		*app.config.Backup.Retention.MaxDays = backup.DefaultBackupRetentionDays
	}
	if app.config.Backup.Retention.MaxCount == nil {
		app.config.Backup.Retention.MaxCount = new(int)
		*app.config.Backup.Retention.MaxCount = backup.DefaultBackupRetentionCount
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

		app.config.Challenges.ProviderConfigs.Http01InternalConfigs = []providers.ConfigManagerHttp01Internal{{
			Domains: []string{"*"},
			Config: &http01internal.Config{
				Port: http01Port,
			},
		}}
	}
}
