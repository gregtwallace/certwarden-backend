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
	DomainAliases map[string]string `json:"domain_aliases"` // DNS Identifier Domain: Domain To Provision On
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
	for k, v := range payload.DomainAliases {
		// check dns identifier domain
		if !validation.DomainValid(k, false) {
			err = fmt.Errorf("identifier `%s` is not a valid domain name", k)
			service.logger.Debug(err)
			return output.JsonErrValidationFailed(err)
		}

		// check dns alias (provisioning domain)
		if !validation.DomainValid(v, false) {
			err = fmt.Errorf("alias `%s` is not a valid domain name", v)
			service.logger.Debug(err)
			return output.JsonErrValidationFailed(err)
		}
	}

	// update service map & write config
	service.dnsIDtoDomain = safemap.NewSafeMapFrom(payload.DomainAliases)
	service.writeAliasConfig()

	// write response
	// ensure map response isn't returned as null
	if payload.DomainAliases == nil {
		payload.DomainAliases = make(map[string]string)
	}

	response := &domainAliasesResponse{}
	response.StatusCode = http.StatusOK
	response.Message = "updated domain aliases"
	response.DomainAliases = payload.DomainAliases

	err = service.output.WriteJSON(w, response)
	if err != nil {
		service.logger.Errorf("failed to write json (%s)", err)
		return output.JsonErrWriteJsonError(err)
	}

	return nil
}
