package app

// CHANGES v1 to v2:
// - cors_permitted_origins renamed to cors_permitted_crossorigins

// configMigrateV1toV2 modifies the unmarhsalled yaml of the config file
// to migrate the config from version 1 to version 2. an error is returned
// if the migration cannot be performed.
func configMigrateV1toV2(cfgFileYamlObj map[string]any) (newCfgVer int, err error) {
	// set config version
	cfgFileYamlObj["config_version"] = 2

	// if old has cors_permitted_origins, set new cors_permitted_crossorigins from old & delete old
	corsPermittedOrigins, ok := cfgFileYamlObj["cors_permitted_origins"]
	if ok {
		cfgFileYamlObj["cors_permitted_crossorigins"] = corsPermittedOrigins
		delete(cfgFileYamlObj, "cors_permitted_origins")
	}

	return 2, nil
}
