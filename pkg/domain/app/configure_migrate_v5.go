package app

import (
	"errors"
	"fmt"
)

// CHANGES v4 to v5:
// -

// configMigrateV4toV5 modifies the unmarhsalled yaml of the config file
// to migrate the config from version 4 to version 5. an error is returned
// if the migration cannot be performed.
func configMigrateV4toV5(cfgFileYamlObj map[string]any) (newCfgVer int, err error) {
	currentSchemaVersion := 4
	newSchemaVersion := 5

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

			// calculate value for wait time (add old pre and post)
			waitVal := 0
			preWait, preOk := provider["precheck_wait"].(int)
			postWait, postOk := provider["postcheck_wait"].(int)
			if preOk && postOk {
				waitVal = preWait + postWait
			}

			// floor wait vals
			if key != "http_01_internal" && waitVal < 5*60 {
				waitVal = 5 * 60
			} else if waitVal < 5 {
				waitVal = 5
			}

			// set new key
			provider["post_resource_provision_wait"] = waitVal

			// delete old pre / post keys
			delete(provider, "precheck_wait")
			delete(provider, "postcheck_wait")

			// update tree
			providerOneTypeArray[i] = provider
		}
		providersMap[key] = providerOneTypeArray
	}
	challengesMap["providers"] = providersMap
	cfgFileYamlObj["challenges"] = challengesMap

	return newSchemaVersion, nil
}
