package orders

import (
	"legocerthub-backend/pkg/output"
	"legocerthub-backend/pkg/storage"
	"legocerthub-backend/pkg/validation"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// GetCertOrders is an http handler that returns all of the orders for a specified cert id
func (service *Service) GetCertOrders(w http.ResponseWriter, r *http.Request) (err error) {
	// convert id param to an integer
	certIdParam := httprouter.ParamsFromContext(r.Context()).ByName("id")
	certId, err := strconv.Atoi(certIdParam)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// if id < 0 it is definitely not valid
	err = validation.IsIdExisting(&certId)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// get from storage
	orders, err := service.storage.GetCertOrders(certId)
	if err != nil {
		// special error case for no record found
		if err == storage.ErrNoRecord {
			service.logger.Debug(err)
			return output.ErrNotFound
		} else {
			service.logger.Error(err)
			return output.ErrStorageGeneric
		}
	}

	// return response to client
	_, err = service.output.WriteJSON(w, http.StatusOK, orders, "orders")
	if err != nil {
		service.logger.Error(err)
		return output.ErrWriteJsonFailed
	}

	return nil
}
