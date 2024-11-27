package orders

import (
	"certwarden-backend/pkg/output"
	"certwarden-backend/pkg/storage"
	"certwarden-backend/pkg/validation"
	"errors"
	"fmt"
	"time"
)

var (
	errCertIdBad  = errors.New("orders: certificate id is invalid")
	errOrderIdBad = errors.New("orders: order id is invalid")
	errIdMismatch = errors.New("orders: order id does not match cert")

	errNoPemContent = errors.New("orders: order doesnt have pem content")

	errOrderRetryFinal      = errors.New("orders: can't retry an order that is in a final state (valid or invalid)")
	errOrderRevokeBadReason = errors.New("orders: bad revocation reason code")
)

// getOrder returns the Order specified by the ids, so long as the Order belongs
// to the Certificate.  An error is returned if the order doesn't exist or if the
// order does not belong to the cert.
func (service *Service) getOrder(certId int, orderId int) (Order, *output.JsonError) {
	// basic check
	if !validation.IsIdExistingValidRange(certId) {
		service.logger.Debug(errCertIdBad)
		return Order{}, output.JsonErrValidationFailed(errCertIdBad)
	}
	if !validation.IsIdExistingValidRange(orderId) {
		service.logger.Debug(errOrderIdBad)
		return Order{}, output.JsonErrValidationFailed(errOrderIdBad)
	}

	// get order from storage
	order, err := service.storage.GetOneOrder(orderId)
	if err != nil {
		// special error case for no record found
		if errors.Is(err, storage.ErrNoRecord) {
			service.logger.Debug(err)
			return Order{}, output.JsonErrNotFound(fmt.Errorf("order id %d not found", orderId))
		} else {
			service.logger.Error(err)
			return Order{}, output.JsonErrStorageGeneric(err)
		}
	}

	// check the cert id on the order matches the cert
	if certId != order.Certificate.ID {
		service.logger.Debug(errIdMismatch)
		return Order{}, output.JsonErrValidationFailed(errIdMismatch)
	}

	return order, nil
}

// isOrderRetryable returns an error if the order is not valid, the order doesn't
// belong to the specified cert, or the order is not in a state that can be retried.
func (service *Service) isOrderRetryable(certId int, orderId int) *output.JsonError {
	order, err := service.getOrder(certId, orderId)
	if err != nil {
		return err
	}

	// check if order is in a final state (can't retry)
	if order.Status == "valid" || order.Status == "invalid" {
		service.logger.Debug(errOrderRetryFinal)
		return output.JsonErrValidationFailed(errOrderRetryFinal)
	}

	return nil
}

// isOrderRevocable verifies order belongs to cert and confirms the order
// is in a state that can be revoked ('valid' and 'valid_to' < current time)
func (service *Service) getOrderForRevocation(certId, orderId int) (Order, *output.JsonError) {
	order, err := service.getOrder(certId, orderId)
	if err != nil {
		return Order{}, err
	}

	// check order is in a state that can be revoked
	// nil check
	if order.ValidTo == nil {
		return Order{}, output.JsonErrValidationFailed(errors.New("valid_to is nil"))
	}

	// confirm order is valid, not already revoked, and not expired (time)
	if !(order.Status == "valid" && !order.KnownRevoked && time.Now().Before(*order.ValidTo)) {
		return Order{}, output.JsonErrValidationFailed(errors.New("order is either: not valid, already revoked, or expired"))
	}

	return order, nil
}

// validRevocationReason returns an error if the specified reasonCode
// is not valid (see: rfc5280 section-5.3.1)
func (service *Service) validRevocationReason(reasonCode int) error {
	// valid codes are 0 through 10 inclusive, except 7
	if reasonCode < 0 || reasonCode == 7 || reasonCode > 10 {
		service.logger.Debug(errOrderRevokeBadReason)
		return output.JsonErrValidationFailed(errOrderRevokeBadReason)
	}

	return nil
}
