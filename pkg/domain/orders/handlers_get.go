package orders

import (
	"legocerthub-backend/pkg/output"
	"legocerthub-backend/pkg/pagination_sort"
	"legocerthub-backend/pkg/storage"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// allOrdersResponse provides the json response struct
// to answer a query for a portion of the cert's orders
type allOrdersResponse struct {
	Orders      []orderSummaryResponse `json:"orders"`
	TotalOrders int                    `json:"total_records"`
}

// GetCertOrders is an http handler that returns all of the orders for a specified cert id
func (service *Service) GetCertOrders(w http.ResponseWriter, r *http.Request) (err error) {
	// parse pagination and sorting
	query := pagination_sort.ParseRequestToQuery(r)

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
	orders, totalRows, err := service.storage.GetOrdersByCert(certId, query)
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
	response := allOrdersResponse{
		TotalOrders: totalRows,
	}

	// populate order summaries for output
	for i := range orders {
		response.Orders = append(response.Orders, orders[i].summaryResponse())
	}

	// return response to client
	_, err = service.output.WriteJSON(w, http.StatusOK, response, "all_orders")
	if err != nil {
		service.logger.Error(err)
		return output.ErrWriteJsonFailed
	}

	return nil
}

// validCurrentResponse is the API response for this query
type validCurrentResponse struct {
	Orders      []orderSummaryResponse `json:"orders"`
	TotalOrders int                    `json:"total_orders"`
}

// GetAllValidCurrentOrders fetches each cert's most recent valid order (essentially this
// is a list of the certificates that are currently being hosted via API key)
func (service *Service) GetAllValidCurrentOrders(w http.ResponseWriter, r *http.Request) (err error) {
	// parse pagination and sorting
	query := pagination_sort.ParseRequestToQuery(r)

	// get from storage
	orders, totalOrders, err := service.storage.GetAllValidCurrentOrders(query, nil)
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
	response := validCurrentResponse{
		TotalOrders: totalOrders,
	}
	for i := range orders {
		response.Orders = append(response.Orders, orders[i].summaryResponse())
	}

	// return response to client
	_, err = service.output.WriteJSON(w, http.StatusOK, response, "valid_current_orders")
	if err != nil {
		service.logger.Error(err)
		return output.ErrWriteJsonFailed
	}
	return nil
}
