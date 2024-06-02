package orders

import (
	"certwarden-backend/pkg/acme"
	"certwarden-backend/pkg/output"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

type orderResponse struct {
	output.JsonResponse
	Order orderSummaryResponse `json:"order"`
}

// NewOrder sends the account information to the ACME new-order endpoint
// which creates a new order for the certificate. If an order already exists
// ACME may send back the existing order instead of creating a new one
// endpoint: /api/v1/certificates/:id/orders
func (service *Service) NewOrder(w http.ResponseWriter, r *http.Request) *output.Error {
	// certId param
	certIdParam := httprouter.ParamsFromContext(r.Context()).ByName("certid")
	certId, err := strconv.Atoi(certIdParam)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// get certificate (validate exists)
	_, outErr := service.certificates.GetCertificate(certId)
	if outErr != nil {
		return outErr
	}

	// place order and kickoff high-priority fulfillment
	newOrder, outErr := service.placeNewOrderAndFulfill(certId, true)
	if outErr != nil {
		return outErr
	}

	// write response
	response := &orderResponse{}
	response.StatusCode = http.StatusCreated
	response.Message = "created order"
	response.Order = newOrder.summaryResponse(service)

	err = service.output.WriteJSON(w, response)
	if err != nil {
		service.logger.Errorf("orders: failed to write json (%s)", err)
		return output.ErrWriteJsonError
	}

	return nil
}

// FulfillExistingOrder is a handler that attempts to fulfill an existing order (i.e.
// move it to the 'valid' state)
func (service *Service) FulfillExistingOrder(w http.ResponseWriter, r *http.Request) *output.Error {
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

	// validate order is acceptable
	outErr := service.isOrderRetryable(certId, orderId)
	if outErr != nil {
		return outErr
	}
	// end validation

	// add to order fulfillment queue
	err = service.fulfillOrder(orderId, true)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// get order from db to return
	order, outErr := service.getOrder(certId, orderId)
	if outErr != nil {
		return outErr
	}

	// write response
	response := &orderResponse{}
	response.StatusCode = http.StatusCreated
	response.Message = "attempting to fulfill order"
	response.Order = order.summaryResponse(service)

	err = service.output.WriteJSON(w, response)
	if err != nil {
		service.logger.Errorf("failed to write json (%s)", err)
		return output.ErrWriteJsonError
	}

	return nil
}

// revokePayload allows clients to specify the revocation reason, it is not
// required
type revokePayload struct {
	Reason int `json:"reason"`
}

// RevokeOrder is a handler that will revoke an order if it is valid and not
// past its valid_to timestamp.
func (service *Service) RevokeOrder(w http.ResponseWriter, r *http.Request) *output.Error {
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

	// parse payload
	var payload revokePayload
	// decode body into payload
	_ = json.NewDecoder(r.Body).Decode(&payload)
	// no need to error check, default int val is 0, which is the
	// desired value if not specified

	// validation / get order
	// revocation reason (see: rfc5280 section-5.3.1)
	err = service.validRevocationReason(payload.Reason)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// order
	order, outErr := service.getOrderForRevocation(certId, orderId)
	if outErr != nil {
		return outErr
	}
	// end validation

	// get account key
	key, err := order.Certificate.CertificateAccount.AcmeAccountKey()
	if err != nil {
		service.logger.Error(err)
		return output.ErrInternal
	}

	// revoke the certificate with ACME
	acmeService, err := service.acmeServerService.AcmeService(order.Certificate.CertificateAccount.AcmeServer.ID)
	if err != nil {
		service.logger.Error(err)
		return output.ErrInternal
	}

	err = acmeService.RevokeCertificate(*order.Pem, payload.Reason, key)
	if err != nil {
		// fail on any non-ACME error OR fail on ACME error if it is not 'already revoked' error type
		acmeErr := new(acme.Error)
		if !errors.As(err, &acmeErr) || acmeErr.Type != "urn:ietf:params:acme:error:alreadyRevoked" {
			service.logger.Error(err)
			return output.ErrInternal
		}
	}

	// update db
	err = service.storage.RevokeOrder(orderId)
	if err != nil {
		service.logger.Error(err)
		return output.ErrStorageGeneric
	}

	// update certificate timestamp
	err = service.storage.UpdateCertUpdatedTime(certId)
	if err != nil {
		service.logger.Error(err)
		// no return
	}

	// get order from db to return
	order, outErr = service.getOrder(certId, orderId)
	if outErr != nil {
		return outErr
	}

	// write response
	response := &orderResponse{}
	response.StatusCode = http.StatusOK
	response.Message = "revoked order"
	response.Order = order.summaryResponse(service)

	err = service.output.WriteJSON(w, response)
	if err != nil {
		service.logger.Errorf("orders: failed to write json (%s)", err)
		return output.ErrWriteJsonError
	}

	return nil
}
