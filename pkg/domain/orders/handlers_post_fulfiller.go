package orders

import (
	"errors"
	"fmt"
	"legocerthub-backend/pkg/output"
	"legocerthub-backend/pkg/storage"
	"legocerthub-backend/pkg/validation"
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
)

// PostProcessOrder executes the certificate's post processing on the specified
// order. Post processing is not run if the order isn't valid.
func (service *Service) PostProcessOrder(w http.ResponseWriter, r *http.Request) *output.Error {
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

	// verify valid, not known revoked, not past validTo, and finalized key isn't deleted; else don't post process it
	if order.Status != "valid" || order.KnownRevoked || order.ValidTo == nil || *order.ValidTo <= int(time.Now().Unix()) || order.FinalizedKey == nil {
		// avoid nil
		validTo := 0
		if order.ValidTo != nil {
			validTo = *order.ValidTo
		}

		// avoid nil
		finalKeyName := "[deleted]"
		if order.FinalizedKey != nil {
			finalKeyName = order.FinalizedKey.Name
		}

		service.logger.Debug(fmt.Errorf("cant post process order %d (status: %s, knownrevoked: %t, final key name %s, validTo: %d, now: %d )", orderId, order.Status, order.KnownRevoked, finalKeyName, validTo, time.Now().Unix()))
		return output.ErrValidationFailed
	}

	// go routing for post processing (use routine to avoid timeout on api call return)
	go func() {
		service.orderFulfiller.executePostProcessing(order)
	}()

	// write response
	response := &output.JsonResponse{}
	response.StatusCode = http.StatusOK
	response.Message = fmt.Sprintf("post processing of order id %d executing", orderId)

	err = service.output.WriteJSON(w, response)
	if err != nil {
		service.logger.Errorf("failed to write json (%s)", err)
		return output.ErrWriteJsonError
	}

	return nil
}
