package updater

import (
	"legocerthub-backend/pkg/output"
	"legocerthub-backend/pkg/storage/sqlite"
	"net/http"
)

// getNewVersionInfoResponse
type getNewVersionInfoResponse struct {
	output.JsonResponse
	NewVersion struct {
		LastCheckedUnixTime  int          `json:"last_checked_time"`
		Available            bool         `json:"available"`
		ConfigVersionMatches bool         `json:"config_version_matches"`
		DbVersionMatches     bool         `json:"database_version_matches"`
		Info                 *versionInfo `json:"info,omitempty"`
	} `json:"new_version"`
}

// GetNewVersionInfo returns if there is a newer known version and if there is
// it returns detailed information about that new version.
func (service *Service) GetNewVersionInfo(w http.ResponseWriter, r *http.Request) *output.Error {
	service.newVersion.mu.RLock()
	defer service.newVersion.mu.RUnlock()

	// does new config version match? if blank, false
	configMatch := false
	if service.newVersion.info != nil {
		configMatch = service.currentConfigVersion == service.newVersion.info.ConfigVersion
	}

	// does new db version match? if blank, false
	dbMatch := false
	if service.newVersion.info != nil {
		dbMatch = sqlite.DbCurrentUserVersion == service.newVersion.info.DatabaseVersion
	}

	// new version or not?
	response := &getNewVersionInfoResponse{}
	response.StatusCode = http.StatusOK
	response.Message = "ok"
	// last checked time -62135596800 (default time.Time value) means never checked
	response.NewVersion.LastCheckedUnixTime = int(service.newVersion.lastCheck.Unix())
	response.NewVersion.Available = service.newVersion.available
	response.NewVersion.ConfigVersionMatches = configMatch
	response.NewVersion.DbVersionMatches = dbMatch
	response.NewVersion.Info = service.newVersion.info

	// write response
	err := service.output.WriteJSON(w, response)
	if err != nil {
		service.logger.Errorf("failed to write json (%s)", err)
		return output.ErrWriteJsonError
	}

	return nil
}

// CheckForNewVersion causes the backend to query the remote update information
// and then return info about any new version.
func (service *Service) CheckForNewVersion(w http.ResponseWriter, r *http.Request) *output.Error {
	// update version info from remote
	err := service.fetchNewVersion()
	if err != nil {
		service.logger.Error(err)
		return output.ErrInternal
	}

	// return new version info (same as GET new version)
	return service.GetNewVersionInfo(w, r)
}
