package challenges

import (
	"legocerthub-backend/pkg/output"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// GetProviders returns all of the currently configured providers
func (service *Service) GetProviders(w http.ResponseWriter, r *http.Request) (err error) {
	ps, err := service.providers.Providers()
	if err != nil {
		service.logger.Debug(err)
		return output.ErrInternal
	}

	err = service.output.WriteJSON(w, http.StatusOK, ps, "providers")
	if err != nil {
		return err
	}
	return nil
}

// GetProvider returns the provider with the specified id number
func (service *Service) GetProvider(w http.ResponseWriter, r *http.Request) (err error) {
	// params
	idParam := httprouter.ParamsFromContext(r.Context()).ByName("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// // if id is new, provide some info
	// if validation.IsIdNew(id) {
	// 	return service.xxx?(w, r)
	// }

	// get the key from storage (and validate id)
	p, err := service.providers.Provider(id)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// return response to client
	err = service.output.WriteJSON(w, http.StatusOK, p, "provider")
	if err != nil {
		return err
	}
	return nil
}
