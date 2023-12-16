package orders

import (
	"errors"
	"legocerthub-backend/pkg/output"
	"legocerthub-backend/pkg/storage"
	"legocerthub-backend/pkg/validation"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

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
