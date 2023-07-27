package updater

import (
	"legocerthub-backend/pkg/output"
	"legocerthub-backend/pkg/storage/sqlite"
	"net/http"
)

// allKeysResponse provides the json response struct
// to answer a query for a portion of the keys
type getNewVersionInfoResponse struct {
	LastCheckedUnixTime  int          `json:"last_checked_time"`
	NewVersionAvailable  bool         `json:"available"`
	ConfigVersionMatches bool         `json:"config_version_matches"`
	DbVersionMatches     bool         `json:"database_version_matches"`
	NewVersionInfo       *versionInfo `json:"info,omitempty"`
}

// GetNewVersionInfo returns if there is a newer known version and if there is
// it returns detailed information about that new version.
func (service *Service) GetNewVersionInfo(w http.ResponseWriter, r *http.Request) (err error) {
	service.mu.RLock()
	defer service.mu.RUnlock()

	// does new config version match? if blank, false
	configMatch := false
	if service.newVersionInfo != nil {
		configMatch = service.currentConfigVersion == service.newVersionInfo.ConfigVersion
	}

	// does new db version match? if blank, false
	dbMatch := false
	if service.newVersionInfo != nil {
		configMatch = sqlite.DbCurrentUserVersion == service.newVersionInfo.DatabaseVersion
	}

	// new version or not?
	response := getNewVersionInfoResponse{
		// last checked time -62135596800 (default time.Time value) means never checked
		LastCheckedUnixTime:  int(service.newVersionLastCheck.Unix()),
		NewVersionAvailable:  service.newVersionAvailable,
		ConfigVersionMatches: configMatch,
		DbVersionMatches:     dbMatch,
		NewVersionInfo:       service.newVersionInfo,
	}

	// return response to client
	_, err = service.output.WriteJSON(w, http.StatusOK, response, "new_version")
	if err != nil {
		return err
	}

	return nil
}

// CheckForNewVersion causes the backend to query the remote update information
// and then return info about any new version.
func (service *Service) CheckForNewVersion(w http.ResponseWriter, r *http.Request) (err error) {
	// update version info from remote
	err = service.fetchNewVersion()
	if err != nil {
		service.logger.Error(err)
		return output.ErrInternal
	}

	// return new version info (same as GET new version)
	return service.GetNewVersionInfo(w, r)
}
