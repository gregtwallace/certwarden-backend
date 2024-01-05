package providers

import (
	"legocerthub-backend/pkg/output"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

type providersResponse struct {
	output.JsonResponse
	Providers []provider `json:"providers"`
}

// GetAllProviders returns all of the providers in manager
func (mgr *Manager) GetAllProviders(w http.ResponseWriter, r *http.Request) *output.Error {
	mgr.mu.RLock()
	defer mgr.mu.RUnlock()

	// read all providers
	var allProviders []provider
	for _, p := range mgr.providers {
		allProviders = append(allProviders, *p)
	}

	// write response
	response := &providersResponse{}
	response.StatusCode = http.StatusOK
	response.Message = "ok"
	response.Providers = allProviders

	err := mgr.output.WriteJSON(w, response)
	if err != nil {
		mgr.logger.Errorf("failed to write json (%s)", err)
		return output.ErrWriteJsonError
	}

	return nil
}

// type domainsResponse struct {
// 	output.JsonResponse
// 	Domains map[string]int `json:"domains"`
// }

// // GetAllDomains returns all domains in manager and the ID of the provider that
// // services them.
// func (mgr *Manager) GetAllDomains(w http.ResponseWriter, r *http.Request) *output.Error {
// 	mgr.mu.RLock()
// 	defer mgr.mu.RUnlock()

// 	// read all domains and associated provider IDs
// 	allDomains := make(map[string]int)
// 	for d, p := range mgr.dP {
// 		allDomains[d] = p.ID
// 	}

// 	// write response
// 	response := &domainsResponse{}
// 	response.StatusCode = http.StatusOK
// 	response.Message = "ok"
// 	response.Domains = allDomains

// 	err := mgr.output.WriteJSON(w, response)
// 	if err != nil {
// 		mgr.logger.Errorf("failed to write json (%s)", err)
// 		return output.ErrWriteJsonError
// 	}

// 	return nil
// }

type providerResponse struct {
	output.JsonResponse
	Provider *provider `json:"provider"`
}

// GetOneProvider a provider from manager based on its ID param
func (mgr *Manager) GetOneProvider(w http.ResponseWriter, r *http.Request) *output.Error {
	mgr.mu.RLock()
	defer mgr.mu.RUnlock()

	// params
	idParam := httprouter.ParamsFromContext(r.Context()).ByName("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		mgr.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// get the provider
	var p *provider
	for _, oneP := range mgr.providers {
		if oneP.ID == id {
			p = oneP
			break
		}
	}
	if p == nil {
		mgr.logger.Debug(errBadID(id))
		return output.ErrValidationFailed
	}

	// write response
	response := &providerResponse{}
	response.StatusCode = http.StatusOK
	response.Message = "ok"
	response.Provider = p

	// return response to client
	err = mgr.output.WriteJSON(w, response)
	if err != nil {
		mgr.logger.Errorf("failed to write json (%s)", err)
		return output.ErrWriteJsonError
	}

	return nil
}
