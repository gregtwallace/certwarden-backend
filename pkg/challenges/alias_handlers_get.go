package challenges

import (
	"certwarden-backend/pkg/output"
	"net/http"
)

type domainAliasesResponse struct {
	output.JsonResponse
	DomainAliases []domainAliasJson `json:"domain_aliases"`
}

// GetAllProviders returns all of the providers in manager
func (service *Service) GetDomainAliases(w http.ResponseWriter, r *http.Request) *output.JsonError {
	// no validation needed

	// write response
	response := &domainAliasesResponse{}
	response.StatusCode = http.StatusOK
	response.Message = "ok"
	response.DomainAliases = service.domainAliases()

	err := service.output.WriteJSON(w, response)
	if err != nil {
		service.logger.Errorf("failed to write json (%s)", err)
		return output.JsonErrWriteJsonError(err)
	}

	return nil
}
