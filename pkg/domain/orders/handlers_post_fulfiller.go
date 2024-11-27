package orders

import (
	"certwarden-backend/pkg/output"
	"certwarden-backend/pkg/storage"
	"certwarden-backend/pkg/validation"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
)

// PostProcessOrder executes the certificate's post processing on the specified
// order. Post processing is not run if the order isn't valid.
func (service *Service) PostProcessOrder(w http.ResponseWriter, r *http.Request) *output.JsonError {
	// get params
	params := httprouter.ParamsFromContext(r.Context())

	certIdParam := params.ByName("certid")
	certId, err := strconv.Atoi(certIdParam)
	if err != nil {
		service.logger.Debug(err)
		return output.JsonErrValidationFailed(err)
	}

	orderIdParam := params.ByName("orderid")
	orderId, err := strconv.Atoi(orderIdParam)
	if err != nil {
		service.logger.Debug(err)
		return output.JsonErrValidationFailed(err)
	}

	// basic check
	if !validation.IsIdExistingValidRange(certId) {
		service.logger.Debug(errCertIdBad)
		return output.JsonErrValidationFailed(errCertIdBad)
	}
	if !validation.IsIdExistingValidRange(orderId) {
		service.logger.Debug(errOrderIdBad)
		return output.JsonErrValidationFailed(errOrderIdBad)
	}

	// get from storage
	order, err := service.storage.GetOneOrder(orderId)
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

	// verify cert id matches the order
	if order.Certificate.ID != certId {
		service.logger.Debug(errIdMismatch)
		return output.JsonErrNotFound(errIdMismatch)
	}

	// verify valid, not known revoked, not past validTo, and finalized key isn't deleted; else don't post process it
	if order.Status != "valid" || order.KnownRevoked || order.ValidTo == nil || order.ValidTo.Before(time.Now()) || order.FinalizedKey == nil {
		// avoid nil
		finalKeyName := "[deleted]"
		if order.FinalizedKey != nil {
			finalKeyName = order.FinalizedKey.Name
		}

		err = fmt.Errorf("orders: cant post process order %d (status: %s, knownrevoked: %t, final key name: %s, validTo: %s)", orderId, order.Status, order.KnownRevoked, finalKeyName, order.ValidTo)
		service.logger.Debug(err)
		return output.JsonErrValidationFailed(err)
	}

	// add to post processing
	err = service.postProcess(order.ID, true)
	if err != nil {
		service.logger.Error(err)
		return output.JsonErrInternal(err)
	}

	// write response
	response := &output.JsonResponse{}
	response.StatusCode = http.StatusOK
	response.Message = fmt.Sprintf("orders: post processing of order id %d executing", orderId)

	err = service.output.WriteJSON(w, response)
	if err != nil {
		service.logger.Errorf("orders: failed to write json (%s)", err)
		return output.JsonErrWriteJsonError(err)
	}

	return nil
}
