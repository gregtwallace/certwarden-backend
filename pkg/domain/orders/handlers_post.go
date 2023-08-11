package orders

import (
	"encoding/json"
	"legocerthub-backend/pkg/acme"
	"legocerthub-backend/pkg/output"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// NewOrder sends the account information to the ACME new-order endpoint
// which creates a new order for the certificate. If an order already exists
// ACME may send back the existing order instead of creating a new one
// endpoint: /api/v1/certificates/:id/orders
func (service *Service) NewOrder(w http.ResponseWriter, r *http.Request) (err error) {
	// certId param
	certIdParam := httprouter.ParamsFromContext(r.Context()).ByName("certid")
	certId, err := strconv.Atoi(certIdParam)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// get certificate (validate exists)
	_, err = service.certificates.GetCertificate(certId)
	if err != nil {
		return err
	}

	// place order and kickoff high-priority fulfillment
	orderId, err := service.placeNewOrderAndFulfill(certId, true)
	if err != nil {
		return err
	}

	// return response to client
	response := output.JsonResponse{
		Status:  http.StatusCreated,
		Message: "order created", // TODO?
		ID:      orderId,
	}

	err = service.output.WriteJSON(w, response.Status, response, "response")
	if err != nil {
		return err
	}

	return nil
}

// FulfillExistingOrder is a handler that attempts to fulfill an existing order (i.e.
// move it to the 'valid' state)
func (service *Service) FulfillExistingOrder(w http.ResponseWriter, r *http.Request) (err error) {
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
	err = service.isOrderRetryable(certId, orderId)
	if err != nil {
		return err
	}
	// end validation

	// kickoff order fulfillment (async)
	err = service.orderFromAcme(orderId, true)
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

	err = service.output.WriteJSON(w, response.Status, response, "response")
	if err != nil {
		return err
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
func (service *Service) RevokeOrder(w http.ResponseWriter, r *http.Request) (err error) {
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
		return err
	}
	// order
	order, err := service.getOrderForRevocation(certId, orderId)
	if err != nil {
		return err
	}
	// end validation

	// get account key
	key, err := order.Certificate.CertificateAccount.AcmeAccountKey()
	if err != nil {
		service.logger.Error(err)
		return // done, failed
	}

	// revoke the certificate with ACME
	acmeService, err := service.acmeServerService.AcmeService(order.Certificate.CertificateAccount.AcmeServer.ID)
	if err != nil {
		service.logger.Error(err)
		return // done, failed
	}
	err = acmeService.RevokeCertificate(*order.Pem, payload.Reason, key)
	// defer err check **

	// if no error, or error is already revoked, update db
	acmeErr, isAcmeErr := err.(acme.Error)
	if err == nil || (isAcmeErr && acmeErr.Type == "urn:ietf:params:acme:error:alreadyRevoked") {
		err = service.storage.RevokeOrder(orderId)
		// defer err check **
	}
	// any other error from ACME Revoke step** OR if that step was no error, this checks for error from db update step
	if err != nil {
		service.logger.Error(err)
		return output.ErrInternal
	}

	// update certificate timestamp
	err = service.storage.UpdateCertUpdatedTime(certId)
	if err != nil {
		service.logger.Error(err)
	}

	// return response to client
	response := output.JsonResponse{
		Status:  http.StatusOK,
		Message: "certificate revoked",
		ID:      orderId,
	}

	err = service.output.WriteJSON(w, response.Status, response, "response")
	if err != nil {
		return err
	}

	return nil
}
