package challenges

import (
	"certwarden-backend/pkg/datatypes/safemap"
	"certwarden-backend/pkg/output"
	"certwarden-backend/pkg/validation"
	"encoding/json"
	"fmt"
	"net/http"
)

type domainAliasesPayload struct {
	DomainAliases []domainAliasJson `json:"domain_aliases"`
}

// PostDomainAliases updates the domain aliases map with the map provided by the client
func (service *Service) PostDomainAliases(w http.ResponseWriter, r *http.Request) *output.JsonError {
	var payload domainAliasesPayload

	// decode body into payload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		service.logger.Debug(err)
		return output.JsonErrValidationFailed(err)
	}

	// validate all values passed (only non-wildcard domains are acceptable)
	m := make(map[string]string)
	for _, alias := range payload.DomainAliases {
		// check dns identifier domain
		if !validation.DomainValid(alias.ChallengeDomain, false) {
			err = fmt.Errorf("identifier `%s` is not a valid domain name", alias.ChallengeDomain)
			service.logger.Debug(err)
			return output.JsonErrValidationFailed(err)
		}

		// check dns alias (provisioning domain)
		if !validation.DomainValid(alias.ProvisionDomain, false) {
			err = fmt.Errorf("alias `%s` is not a valid domain name", alias.ProvisionDomain)
			service.logger.Debug(err)
			return output.JsonErrValidationFailed(err)
		}

		// error if duplicate identifier / challenge domain
		_, exists := m[alias.ChallengeDomain]
		if exists {
			err = fmt.Errorf("identifier `%s` specified more than once", alias.ChallengeDomain)
			service.logger.Debug(err)
			return output.JsonErrValidationFailed(err)
		}

		// add alias to map
		m[alias.ChallengeDomain] = alias.ProvisionDomain
	}

	// update service map & write config
	service.dnsIDtoDomain = safemap.NewSafeMapFrom(m)
	service.writeAliasConfig()

	// write response
	response := &domainAliasesResponse{}
	response.StatusCode = http.StatusOK
	response.Message = "updated domain aliases"
	response.DomainAliases = service.domainAliases()

	err = service.output.WriteJSON(w, response)
	if err != nil {
		service.logger.Errorf("failed to write json (%s)", err)
		return output.JsonErrWriteJsonError(err)
	}

	return nil
}
