package challenges

import (
	"net/http"
)

// GetProviders returns all of the currently configured providers configs
func (service *Service) GetProvidersConfigs(w http.ResponseWriter, r *http.Request) (err error) {
	err = service.output.WriteJSON(w, http.StatusOK, service.providers.configs(), "providers")
	if err != nil {
		return err
	}
	return nil
}
