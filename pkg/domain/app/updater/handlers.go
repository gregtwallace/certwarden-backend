package updater

import (
	"legocerthub-backend/pkg/output"
	"net/http"
)

// allKeysResponse provides the json response struct
// to answer a query for a portion of the keys
type getNewVersionInfoResponse struct {
	NewVersionAvailable bool         `json:"available"`
	NewVersionInfo      *versionInfo `json:"info,omitempty"`
}

// GetNewVersionInfo returns if there is a newer known version and if there is
// it returns detailed information about that new version.
func (service *Service) GetNewVersionInfo(w http.ResponseWriter, r *http.Request) (err error) {
	service.mu.RLock()
	defer service.mu.RUnlock()

	// new version or not?
	response := getNewVersionInfoResponse{
		NewVersionAvailable: service.newVersionAvailable,
		NewVersionInfo:      service.newVersionInfo,
	}

	// return response to client
	_, err = service.output.WriteJSON(w, http.StatusOK, response, "new_version")
	if err != nil {
		service.logger.Error(err)
		return output.ErrWriteJsonFailed
	}

	return nil
}