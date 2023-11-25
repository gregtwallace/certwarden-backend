package providers

import (
	"encoding/json"
	"fmt"
	"legocerthub-backend/pkg/output"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// deletePayload is the needed payload to delete a provider
type deletePayload struct {
	ID  int    `json:"-"`
	Tag string `json:"tag"`
}

// DeleteProvider deletes the provider specified by the ID from manager also freeing
// up the domains previously mapped to it. If the tag is not specified or is incorrect
// deleting fails.
func (mgr *Manager) DeleteProvider(w http.ResponseWriter, r *http.Request) *output.Error {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	// decode body into payload
	var payload deletePayload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		mgr.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// params
	idParam := httprouter.ParamsFromContext(r.Context()).ByName("id")
	payload.ID, err = strconv.Atoi(idParam)
	if err != nil {
		mgr.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// if manager only has 1 provider, delete will never be allowed
	if len(mgr.pD) <= 1 {
		mgr.logger.Debug("cannot delete provider if there is only 1 provider available")
		return output.ErrValidationFailed
	}

	// find provider
	p := (*provider)(nil)
	for oneP := range mgr.pD {
		if oneP.ID == payload.ID {

			// once found, verify tag is correct
			if oneP.Tag == payload.Tag {
				p = oneP
				break
			} else {
				mgr.logger.Debug(errWrongTag)
				return output.ErrValidationFailed
			}
		}
	}

	// didn't find id
	if p == nil {
		mgr.logger.Debug(errBadID(payload.ID))
		return output.ErrValidationFailed
	}

	// call provider stop func before deleting
	err = p.Stop()
	if err != nil {
		// if error just log it
		// if app is unstable, rely on provider service to call Fatal
		mgr.logger.Errorf("failed to stop provider being deleted (%s)", err)
	}

	// actually do deletion
	mgr.unsafeDeleteProvider(p)

	// update config file
	err = mgr.unsafeWriteProvidersConfig()
	if err != nil {
		mgr.logger.Errorf("failed to save config file after providers update (%s)", err)
		return output.ErrInternal
	}

	// write response
	response := &output.JsonResponse{
		StatusCode: http.StatusOK,
		Message:    fmt.Sprintf("deleted provider (id: %d)", payload.ID),
	}

	err = mgr.output.WriteJSON(w, response)
	if err != nil {
		mgr.logger.Errorf("failed to write json (%s)", err)
		return output.ErrWriteJsonError
	}

	return nil
}
