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
	certIdStr := httprouter.ParamsFromContext(r.Context()).ByName("certid")

	// convert id param to an integer
	certId, err := strconv.Atoi(certIdStr)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
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
	if !service.isOrderRetryable(certId, orderId) {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}
	///

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

	_, err = service.output.WriteJSON(w, response.Status, response, "response")
	if err != nil {
		service.logger.Error(err)
		return output.ErrWriteJsonFailed
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
	params := httprouter.ParamsFromContext(r.Context())
	certIdStr := params.ByName("certid")
	orderIdStr := params.ByName("orderid")

	// convert params to integers
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

	// parse payload
	var payload revokePayload
	// decode body into payload
	_ = json.NewDecoder(r.Body).Decode(&payload)
	// no need to error check, default int val is 0, which is the
	// desired value if not specified

	// validation
	// ids (ensure cert and order match)
	// also confirm order is revocable (valid and unexpired)
	if !service.isOrderRevocable(certId, orderId) {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}
	// end validation

	// revoke cert
	// fetch the relevant order
	order, err := service.storage.GetOneOrder(orderId)
	if err != nil {
		service.logger.Error(err)
		return // done, failed
	}

	// fetch the certificate with sensitive data and update the order object
	order.Certificate, err = service.storage.GetOneCertById(order.Certificate.ID)
	if err != nil {
		service.logger.Error(err)
		return // done, failed
	}

	// get account key
	key, err := order.Certificate.CertificateAccount.AcmeAccountKey()
	if err != nil {
		service.logger.Error(err)
		return // done, failed
	}

	// revoke the certificate with ACME
	if order.Certificate.CertificateAccount.IsStaging {
		err = service.acmeStaging.RevokeCertificate(*order.Pem, payload.Reason, key)
	} else {
		err = service.acmeProd.RevokeCertificate(*order.Pem, payload.Reason, key)
	}
	// if no error, or error is already revoked, update db
	acmeErr, isAcmeErr := err.(acme.Error)
	if err == nil || (isAcmeErr && acmeErr.Type == "urn:ietf:params:acme:error:alreadyRevoked") {
		err = service.storage.RevokeOrder(orderId)
	}
	// checks err for IsStaging, or update to storage (if that condition was met)
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

	_, err = service.output.WriteJSON(w, response.Status, response, "response")
	if err != nil {
		service.logger.Error(err)
		return output.ErrWriteJsonFailed
	}

	return nil
}
