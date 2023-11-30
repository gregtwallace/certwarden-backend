package app

import "fmt"

// CHANGES v2 to v3:
// - pprof_port renamed to pprof_http_port
// - pprof_https_port added

// configMigrateV2toV3 modifies the unmarhsalled yaml of the config file
// to migrate the config from version 2 to version 3. an error is returned
// if the migration cannot be performed.
func configMigrateV2toV3(cfgFileYamlObj map[string]any) (newCfgVer int, err error) {
	currentSchemaVersion := 2
	newSchemaVersion := 3

	if cfgFileYamlObj["config_version"] != currentSchemaVersion {
		return -1, fmt.Errorf("cannot update schema, current version %d (expected %d)", cfgFileYamlObj["config_version"], currentSchemaVersion)
	}

	// set config version
	cfgFileYamlObj["config_version"] = newSchemaVersion

	// if old has pprof_port, set new pprof_http_port from old & delete old
	pprofPort, ok := cfgFileYamlObj["pprof_port"]
	if ok {
		cfgFileYamlObj["pprof_http_port"] = pprofPort
		delete(cfgFileYamlObj, "pprof_port")
	}

	return newSchemaVersion, nil
}
