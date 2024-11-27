package orders

import (
	"certwarden-backend/pkg/output"
	"certwarden-backend/pkg/storage"
	"certwarden-backend/pkg/validation"
	"errors"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// DownloadCertNewestOrder returns the pem from the cert's newest valid order to the client
func (service *Service) DownloadCertNewestOrder(w http.ResponseWriter, r *http.Request) *output.JsonError {
	// get id from param
	idParam := httprouter.ParamsFromContext(r.Context()).ByName("certid")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		service.logger.Debug(err)
		return output.JsonErrValidationFailed(err)
	}

	// get from storage
	order, err := service.storage.GetCertNewestValidOrderById(id)
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

	// nil check of pem
	if order.Pem == nil || *order.Pem == "" {
		service.logger.Debug(errNoPemContent)
		return output.JsonErrNotFound(errNoPemContent)
	}

	// return pem file to client
	service.output.WritePem(w, r, order)

	return nil
}

// DownloadOneOrder returns the pem for a single cert to the client
func (service *Service) DownloadOneOrder(w http.ResponseWriter, r *http.Request) *output.JsonError {
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

	// if user requests an order without pem content, fail
	if order.Pem == nil || *order.Pem == "" {
		service.logger.Debug(errNoPemContent)
		return output.JsonErrNotFound(errNoPemContent)
	}

	// return pem file to client
	service.output.WritePem(w, r, order)

	return nil
}
