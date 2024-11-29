package challenges

import (
	"certwarden-backend/pkg/output"
	"net/http"
)

type domainAliasesResponse struct {
	output.JsonResponse
	DomainAliases map[string]string `json:"domain_aliases"` // DNS Identifier Domain: Domain To Provision On
}

// GetAllProviders returns all of the providers in manager
func (service *Service) GetDomainAliases(w http.ResponseWriter, r *http.Request) *output.JsonError {
	// no validation needed

	// acquire regular map from service
	m := make(map[string]string)
	service.dnsIDtoDomain.CopyToMap(m)

	// write response
	response := &domainAliasesResponse{}
	response.StatusCode = http.StatusOK
	response.Message = "ok"
	response.DomainAliases = m

	err := service.output.WriteJSON(w, response)
	if err != nil {
		service.logger.Errorf("failed to write json (%s)", err)
		return output.JsonErrWriteJsonError(err)
	}

	return nil
}
