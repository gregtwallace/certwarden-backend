package orders

import (
	"legocerthub-backend/pkg/acme"
	"legocerthub-backend/pkg/output"
	"legocerthub-backend/pkg/validation"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// NewOrder sends the account information to the ACME new-order endpoint
// which creates a new order for the certificate. If an order already exists
// ACME may send back the existing order instead of creating a new one
// endpoint: /api/v1/certificates/:id/orders
func (service *Service) NewOrder(w http.ResponseWriter, r *http.Request) (err error) {
	certIdStr := httprouter.ParamsFromContext(r.Context()).ByName("certid")

	// convert id param to an integer
	certId, err := strconv.Atoi(certIdStr)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// fetch the relevant cert
	cert, err := service.storage.GetOneCertById(certId, true)
	if err != nil {
		service.logger.Error(err)
		return output.ErrStorageGeneric
	}

	// no need to validate, can try to order any cert in storage

	// get account key
	key, err := cert.AcmeAccount.AccountKey()
	if err != nil {
		service.logger.Error(err)
		return output.ErrInternal
	}

	// send the new-order to ACME
	var acmeResponse acme.Order
	if *cert.AcmeAccount.IsStaging {
		acmeResponse, err = service.acmeStaging.NewOrder(cert.NewOrderPayload(), key)
	} else {
		acmeResponse, err = service.acmeProd.NewOrder(cert.NewOrderPayload(), key)
	}
	if err != nil {
		service.logger.Error(err)
		return output.ErrInternal
	}
	service.logger.Debugf("new order location: %s", acmeResponse.Location)

	// save ACME response to order storage
	orderId, err := service.storage.PostNewOrder(cert, acmeResponse)
	if err != nil {
		service.logger.Error(err)
		return output.ErrStorageGeneric
	}

	// kickoff order fulfillment (async)
	_ = service.orderFromAcme(orderId)
	// if already being worked, no need to indicate that since this is a creation
	// task

	// return response to client
	response := output.JsonResponse{
		Status: http.StatusCreated,
		// Message: "order created", // TODO?
		Message: acmeResponse,
		ID:      orderId,
	}

	_, err = service.output.WriteJSON(w, response.Status, response, "response")
	if err != nil {
		service.logger.Error(err)
		return output.ErrWriteJsonFailed
	}

	return nil
}

// FulfillExistingOrder is a handler that attempts to fulfill an existing order (i.e.
// move it to the 'valid' state)
func (service *Service) FulfillExistingOrder(w http.ResponseWriter, r *http.Request) (err error) {
	params := httprouter.ParamsFromContext(r.Context())
	certIdStr := params.ByName("certid")
	orderIdStr := params.ByName("orderid")

	// convert id params to integers
	certId, err := strconv.Atoi(certIdStr)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}
	orderId, err := strconv.Atoi(orderIdStr)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	/// validation
	// ids (ensure cert and order match)
	// also confirm order isn't invalid
	err = service.isOrderRetryable(certId, orderId)
	if err == validation.ErrOrderInvalid {
		service.logger.Debug(err)
		return output.ErrOrderInvalid
	} else if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}
	///

	// kickoff order fulfillement (async)
	err = service.orderFromAcme(orderId)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrOrderCantFulfill
	}

	// return response to client
	response := output.JsonResponse{
		Status:  http.StatusOK,
		Message: "attempting to fulfill",
		ID:      orderId,
	}

	_, err = service.output.WriteJSON(w, response.Status, response, "response")
	if err != nil {
		service.logger.Error(err)
		return output.ErrWriteJsonFailed
	}

	return nil
}
