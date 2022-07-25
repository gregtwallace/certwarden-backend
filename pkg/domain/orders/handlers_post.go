package orders

import (
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
	idParamStr := httprouter.ParamsFromContext(r.Context()).ByName("id")

	// convert id param to an integer
	idParam, err := strconv.Atoi(idParamStr)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// fetch the relevant cert
	cert, err := service.storage.GetOneCertById(idParam, true)
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

	// create order object from response
	order := makeNewOrder(&cert, &acmeResponse)

	// save ACME response to order storage
	order.ID = new(int)
	*order.ID, err = service.storage.PostNewOrder(order)
	if err != nil {
		service.logger.Error(err)
		return output.ErrStorageGeneric
	}

	// kickoff order fulfillement (async)
	service.FulfillOrder(order)

	// return response to client
	response := output.JsonResponse{
		Status: http.StatusCreated,
		// Message: "order created", // TODO?
		Message: acmeResponse,
		ID:      *order.ID,
	}

	_, err = service.output.WriteJSON(w, response.Status, response, "response")
	if err != nil {
		service.logger.Error(err)
		return output.ErrWriteJsonFailed
	}

	return nil
}
