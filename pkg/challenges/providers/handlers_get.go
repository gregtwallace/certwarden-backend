package providers

import (
	"fmt"
	"legocerthub-backend/pkg/output"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

var errBadID = func(id int) error { return fmt.Errorf("no provider exists with id %d", id) }

// GetAllProviders returns all of the providers in manager
func (mgr *Manager) GetAllProviders(w http.ResponseWriter, r *http.Request) (err error) {
	mgr.mu.RLock()
	defer mgr.mu.RUnlock()

	var allProviders []*provider

	for p := range mgr.pD {
		allProviders = append(allProviders, p)
	}

	err = mgr.output.WriteJSON(w, http.StatusOK, allProviders, "providers")
	if err != nil {
		return err
	}
	return nil
}

// GetAllDomains returns all domains in manager and the ID of the provider that
// services them.
func (mgr *Manager) GetAllDomains(w http.ResponseWriter, r *http.Request) (err error) {
	mgr.mu.RLock()
	defer mgr.mu.RUnlock()

	allDomains := make(map[string]int)

	for d, p := range mgr.dP {
		allDomains[d] = p.ID
	}

	err = mgr.output.WriteJSON(w, http.StatusOK, allDomains, "domains")
	if err != nil {
		return err
	}
	return nil
}

// GetOneProvider a provider from manager based on its ID param
func (mgr *Manager) GetOneProvider(w http.ResponseWriter, r *http.Request) (err error) {
	mgr.mu.RLock()
	defer mgr.mu.RUnlock()

	// params
	idParam := httprouter.ParamsFromContext(r.Context()).ByName("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		mgr.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// // if id is new, provide some info
	// if validation.IsIdNew(id) {
	// 	return service.xxx?(w, r)
	// }

	// get the provider
	var p *provider
	for oneP := range mgr.pD {
		if oneP.ID == id {
			p = oneP
			break
		}
	}
	if p == nil {
		mgr.logger.Debug(errBadID(id))
		return output.ErrValidationFailed
	}

	// return response to client
	err = mgr.output.WriteJSON(w, http.StatusOK, p, "provider")
	if err != nil {
		return err
	}
	return nil
}
