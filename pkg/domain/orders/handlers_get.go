package orders

import (
	"legocerthub-backend/pkg/output"
	"legocerthub-backend/pkg/storage"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// GetCertOrders is an http handler that returns all of the orders for a specified cert id
func (service *Service) GetCertOrders(w http.ResponseWriter, r *http.Request) (err error) {
	// convert id param to an integer
	certIdParam := httprouter.ParamsFromContext(r.Context()).ByName("certid")
	certId, err := strconv.Atoi(certIdParam)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// validate certificate ID
	_, err = service.certificates.GetCertificate(certId)
	if err != nil {
		return err
	}

	// get orders from storage
	orders, err := service.storage.GetOrdersByCert(certId)
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

	// response
	var response []orderSummaryResponse
	for i := range orders {
		response = append(response, orders[i].summaryResponse())
	}

	// return response to client
	_, err = service.output.WriteJSON(w, http.StatusOK, response, "orders")
	if err != nil {
		service.logger.Error(err)
		return output.ErrWriteJsonFailed
	}

	return nil
}

// GetAllValidCurrentOrders fetches each cert's most recent valid order (essentially this
// is a list of the certificates that are currently being hosted via API key)
func (service *Service) GetAllValidCurrentOrders(w http.ResponseWriter, r *http.Request) (err error) {
	// get from storage
	orders, err := service.storage.GetAllValidCurrentOrders()
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

	// response
	var response []orderSummaryResponse
	for i := range orders {
		response = append(response, orders[i].summaryResponse())
	}

	// return response to client
	_, err = service.output.WriteJSON(w, http.StatusOK, response, "orders")
	if err != nil {
		service.logger.Error(err)
		return output.ErrWriteJsonFailed
	}
	return nil
}
