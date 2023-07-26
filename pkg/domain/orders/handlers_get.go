package orders

import (
	"fmt"
	"legocerthub-backend/pkg/output"
	"legocerthub-backend/pkg/pagination_sort"
	"legocerthub-backend/pkg/storage"
	"legocerthub-backend/pkg/validation"
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
		response.Orders = append(response.Orders, orders[i].summaryResponse(service))
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
	orders, totalOrders, err := service.storage.GetAllValidCurrentOrders(query)
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
		response.Orders = append(response.Orders, orders[i].summaryResponse(service))
	}

	// return response to client
	_, err = service.output.WriteJSON(w, http.StatusOK, response, "valid_current_orders")
	if err != nil {
		service.logger.Error(err)
		return output.ErrWriteJsonFailed
	}
	return nil
}

// DownloadCertNewestOrder returns the pem from the cert's newest valid order to the client
func (service *Service) DownloadCertNewestOrder(w http.ResponseWriter, r *http.Request) (err error) {
	// insecure okay, cert pem is not private

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
		if err == storage.ErrNoRecord {
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
	_, err = service.output.WritePem(w, fmt.Sprintf("%s.cert.pem", order.Certificate.Name), *order.Pem)
	if err != nil {
		service.logger.Error(err)
		return output.ErrWritePemFailed
	}

	return nil
}

// DownloadOneOrder returns the pem for a single cert to the client
func (service *Service) DownloadOneOrder(w http.ResponseWriter, r *http.Request) (err error) {
	// insecure okay, cert pem is not private

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
		if err == storage.ErrNoRecord {
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
	_, err = service.output.WritePem(w, fmt.Sprintf("%s.cert.pem", order.Certificate.Name), *order.Pem)
	if err != nil {
		service.logger.Error(err)
		return output.ErrWritePemFailed
	}

	return nil
}
