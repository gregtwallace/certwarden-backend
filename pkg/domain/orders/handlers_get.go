package orders

import (
	"certwarden-backend/pkg/output"
	"certwarden-backend/pkg/pagination_sort"
	"certwarden-backend/pkg/storage"
	"errors"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// allOrdersResponse provides the json response struct
// to answer a query for a portion of a cert's orders
type allOrdersResponse struct {
	output.JsonResponse
	TotalOrders int                    `json:"total_records"`
	Orders      []orderSummaryResponse `json:"orders"`
}

// GetCertOrders is an http handler that returns all of the orders for a specified cert id
func (service *Service) GetCertOrders(w http.ResponseWriter, r *http.Request) *output.JsonError {
	// parse pagination and sorting
	query := pagination_sort.ParseRequestToQuery(r)

	// convert id param to an integer
	certIdParam := httprouter.ParamsFromContext(r.Context()).ByName("certid")
	certId, err := strconv.Atoi(certIdParam)
	if err != nil {
		service.logger.Debug(err)
		return output.JsonErrValidationFailed(err)
	}

	// validate certificate ID
	_, outErr := service.certificates.GetCertificate(certId)
	if outErr != nil {
		return outErr
	}

	// get orders from storage
	orders, totalRows, err := service.storage.GetOrdersByCert(certId, query)
	if err != nil {
		// special error case for no record found
		if errors.Is(err, storage.ErrNoRecord) {
			service.logger.Debug(err)
			return output.JsonErrNotFound(err)
		} else {
			service.logger.Error(err)
			return output.JsonErrStorageGeneric(err)
		}
	}

	// populate order summaries for output
	outputOrders := []orderSummaryResponse{}
	for i := range orders {
		outputOrders = append(outputOrders, orders[i].summaryResponse(service))
	}

	// write response
	response := &allOrdersResponse{}
	response.StatusCode = http.StatusOK
	response.Message = "ok"
	response.TotalOrders = totalRows
	response.Orders = outputOrders

	err = service.output.WriteJSON(w, response)
	if err != nil {
		service.logger.Errorf("orders: failed to write json (%s)", err)
		return output.JsonErrWriteJsonError(err)
	}

	return nil
}

// GetAllValidCurrentOrders fetches each cert's most recent valid order (essentially this
// is a list of the certificates that are currently being hosted via API key)
func (service *Service) GetAllValidCurrentOrders(w http.ResponseWriter, r *http.Request) *output.JsonError {
	// parse pagination and sorting
	query := pagination_sort.ParseRequestToQuery(r)

	// get from storage
	orders, totalRows, err := service.storage.GetAllValidCurrentOrders(query)
	if err != nil {
		// special error case for no record found
		if errors.Is(err, storage.ErrNoRecord) {
			service.logger.Debug(err)
			return output.JsonErrNotFound(err)
		} else {
			service.logger.Error(err)
			return output.JsonErrStorageGeneric(err)
		}
	}

	// populate order summaries for output
	outputOrders := []orderSummaryResponse{}
	for i := range orders {
		outputOrders = append(outputOrders, orders[i].summaryResponse(service))
	}

	// write response
	response := &allOrdersResponse{}
	response.StatusCode = http.StatusOK
	response.Message = "ok"
	response.TotalOrders = totalRows
	response.Orders = outputOrders

	err = service.output.WriteJSON(w, response)
	if err != nil {
		service.logger.Errorf("orders: failed to write json (%s)", err)
		return output.JsonErrWriteJsonError(err)
	}

	return nil
}
