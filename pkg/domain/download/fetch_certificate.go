package download

import (
	"certwarden-backend/pkg/domain/orders"
	"certwarden-backend/pkg/output"
	"certwarden-backend/pkg/storage"
	"errors"
)

// getCertNewestValidOrder returns the most recent valid order for the specified certificate if the
// apiKey matches the requested cert. It also checks the apiKeyViaUrl property if the client is making
// a request with the apiKey in the Url.
func (service *Service) getCertNewestValidOrder(certName string, apiKey string, apiKeyViaUrl bool) (orders.Order, *output.Error) {
	// if apiKey is blank, definitely unauthorized
	if apiKey == "" {
		service.logger.Debug(errBlankApiKey)
		return orders.Order{}, output.ErrUnauthorized
	}

	// get the cert's newest valid order from storage
	order, err := service.storage.GetCertNewestValidOrderByName(certName)
	if err != nil {
		// special error case for no record found
		if errors.Is(err, storage.ErrNoRecord) {
			service.logger.Debug(err)
			return orders.Order{}, output.ErrNotFound
		} else {
			service.logger.Error(err)
			return orders.Order{}, output.ErrStorageGeneric
		}
	}

	// if apiKey came from URL, and cert does not support this, error
	if apiKeyViaUrl && !order.Certificate.ApiKeyViaUrl {
		service.logger.Debug(errApiKeyFromUrlDisallowed)
		return orders.Order{}, output.ErrUnauthorized
	}

	// verify apikey matches cert apikey (new or old)
	if (apiKey != order.Certificate.ApiKey) && (apiKey != order.Certificate.ApiKeyNew) {
		service.logger.Debug(errWrongApiKey)
		return orders.Order{}, output.ErrUnauthorized
	}

	// pem cant be blank
	if order.Pem == nil || *order.Pem == "" {
		service.logger.Debug(errNoPem)
		return orders.Order{}, output.ErrNotFound
	}

	// return pem content and cert name
	return order, nil
}
