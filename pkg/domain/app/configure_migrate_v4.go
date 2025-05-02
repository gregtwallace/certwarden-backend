package app

import (
	"errors"
	"fmt"
)

// CHANGES v3 to v4:
// - Add `precheck_wait` and `postcheck_wait` to provider configs
// - Automatically populate values with defaults that roughly mirror the previous behavior

// configMigrateV3toV4 modifies the unmarhsalled yaml of the config file
// to migrate the config from version 3 to version 4. an error is returned
// if the migration cannot be performed.
func configMigrateV3toV4(cfgFileYamlObj map[string]any) (newCfgVer int, err error) {
	currentSchemaVersion := 3
	newSchemaVersion := 4

	if cfgFileYamlObj["config_version"] != currentSchemaVersion {
		return -1, fmt.Errorf("cannot update cfg schema, current version %d (expected %d)", cfgFileYamlObj["config_version"], currentSchemaVersion)
	}

	// set config version
	cfgFileYamlObj["config_version"] = newSchemaVersion

	// drill down to each provider and update with sane defaults
	challenges, ok := cfgFileYamlObj["challenges"]
	if !ok {
		// nothing to update
		return newSchemaVersion, nil
	}

	challengesMap, ok := challenges.(map[string]any)
	if !ok {
		return -1, errors.New("cannot update cfg schema, `challenges` in config file is wrong type")
	}

	providers, ok := challengesMap["providers"]
	if !ok {
		// nothing to update
		return newSchemaVersion, nil
	}

	providersMap, ok := providers.(map[string]any)
	if !ok {
		return -1, errors.New("cannot update cfg schema, `providers` in config file is wrong type")
	}

	// range through keys of each provider type
	for key := range providersMap {
		providerOneTypeArray, ok := providersMap[key].([]any)
		if !ok {
			return -1, fmt.Errorf("cannot update cfg schema, `%s` in config file is wrong type", key)
		}

		// range through each individual provider
		for i := range providerOneTypeArray {
			provider, ok := providerOneTypeArray[i].(map[string]any)
			if !ok {
				return -1, fmt.Errorf("cannot update cfg schema, provider index %d of type %s in config file is wrong type", i, key)
			}

			// default
			provider["precheck_wait"] = 3 * 60
			provider["postcheck_wait"] = 0

			// if internal http, special default
			if key == "http_01_internal" {
				provider["precheck_wait"] = 0
				// provider["postcheck_wait"] = 0
			}

			// update tree
			providerOneTypeArray[i] = provider
		}
		providersMap[key] = providerOneTypeArray
	}
	challengesMap["providers"] = providersMap
	cfgFileYamlObj["challenges"] = challengesMap

	return newSchemaVersion, nil
}
