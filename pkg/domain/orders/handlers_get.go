package orders

import (
	"errors"
	"legocerthub-backend/pkg/output"
	"legocerthub-backend/pkg/pagination_sort"
	"legocerthub-backend/pkg/storage"
	"legocerthub-backend/pkg/validation"
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
func (service *Service) GetCertOrders(w http.ResponseWriter, r *http.Request) *output.Error {
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
			return output.ErrNotFound
		} else {
			service.logger.Error(err)
			return output.ErrStorageGeneric
		}
	}

	// populate order summaries for output
	outputOrders := []orderSummaryResponse{}
	for i := range orders {
		outputOrders = append(outputOrders, orders[i].summaryResponse(service.orderFulfiller))
	}

	// write response
	response := &allOrdersResponse{}
	response.StatusCode = http.StatusOK
	response.Message = "ok"
	response.TotalOrders = totalRows
	response.Orders = outputOrders

	err = service.output.WriteJSON(w, response)
	if err != nil {
		service.logger.Errorf("failed to write json (%s)", err)
		return output.ErrWriteJsonError
	}

	return nil
}

// GetAllValidCurrentOrders fetches each cert's most recent valid order (essentially this
// is a list of the certificates that are currently being hosted via API key)
func (service *Service) GetAllValidCurrentOrders(w http.ResponseWriter, r *http.Request) *output.Error {
	// parse pagination and sorting
	query := pagination_sort.ParseRequestToQuery(r)

	// get from storage
	orders, totalRows, err := service.storage.GetAllValidCurrentOrders(query)
	if err != nil {
		// special error case for no record found
		if errors.Is(err, storage.ErrNoRecord) {
			service.logger.Debug(err)
			return output.ErrNotFound
		} else {
			service.logger.Error(err)
			return output.ErrStorageGeneric
		}
	}

	// populate order summaries for output
	outputOrders := []orderSummaryResponse{}
	for i := range orders {
		outputOrders = append(outputOrders, orders[i].summaryResponse(service.orderFulfiller))
	}

	// write response
	response := &allOrdersResponse{}
	response.StatusCode = http.StatusOK
	response.Message = "ok"
	response.TotalOrders = totalRows
	response.Orders = outputOrders

	err = service.output.WriteJSON(w, response)
	if err != nil {
		service.logger.Errorf("failed to write json (%s)", err)
		return output.ErrWriteJsonError
	}

	return nil
}

// DownloadCertNewestOrder returns the pem from the cert's newest valid order to the client
func (service *Service) DownloadCertNewestOrder(w http.ResponseWriter, r *http.Request) *output.Error {
	// get id from param
	idParam := httprouter.ParamsFromContext(r.Context()).ByName("certid")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// get from storage
	order, err := service.storage.GetCertNewestValidOrderById(id)
	if err != nil {
		// special error case for no record found
		if errors.Is(err, storage.ErrNoRecord) {
			service.logger.Debug(err)
			return output.ErrNotFound
		} else {
			service.logger.Error(err)
			return output.ErrStorageGeneric
		}
	}

	// nil check of pem
	if order.Pem == nil || *order.Pem == "" {
		service.logger.Debug(errNoPemContent)
		return output.ErrNotFound
	}

	// return pem file to client
	service.output.WritePem(w, r, order)

	return nil
}

// DownloadOneOrder returns the pem for a single cert to the client
func (service *Service) DownloadOneOrder(w http.ResponseWriter, r *http.Request) *output.Error {
	// get params
	params := httprouter.ParamsFromContext(r.Context())

	certIdParam := params.ByName("certid")
	certId, err := strconv.Atoi(certIdParam)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	orderIdParam := params.ByName("orderid")
	orderId, err := strconv.Atoi(orderIdParam)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// basic check
	if !validation.IsIdExistingValidRange(certId) {
		service.logger.Debug(errCertIdBad)
		return output.ErrValidationFailed
	}
	if !validation.IsIdExistingValidRange(orderId) {
		service.logger.Debug(errOrderIdBad)
		return output.ErrValidationFailed
	}

	// get from storage
	order, err := service.storage.GetOneOrder(orderId)
	if err != nil {
		// special error case for no record found
		if errors.Is(err, storage.ErrNoRecord) {
			service.logger.Debug(err)
			return output.ErrNotFound
		} else {
			service.logger.Error(err)
			return output.ErrStorageGeneric
		}
	}

	// verify cert id matches the order
	if order.Certificate.ID != certId {
		service.logger.Debug(errIdMismatch)
		return output.ErrNotFound
	}

	// if user requests an order without pem content, fail
	if order.Pem == nil || *order.Pem == "" {
		service.logger.Debug(errNoPemContent)
		return output.ErrNotFound
	}

	// return pem file to client
	service.output.WritePem(w, r, order)

	return nil
}

// GetAllWorkStatus returns all jobs/orders currently with fulfiller
func (service *Service) GetAllWorkStatus(w http.ResponseWriter, r *http.Request) *output.Error {
	err := service.output.WriteJSON(w, service.orderFulfiller.allWorkStatus())
	if err != nil {
		service.logger.Errorf("failed to write json (%s)", err)
		return output.ErrWriteJsonError
	}
	return nil
}
