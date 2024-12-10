package download

import (
	"certwarden-backend/pkg/domain/orders"
	"certwarden-backend/pkg/output"
	"certwarden-backend/pkg/storage"
	"errors"
	"strings"
	"time"
)

// getCertNewestValidOrder returns the most recent valid order for the specified certificate if the
// apiKey matches the requested cert. It also checks the apiKeyViaUrl property if the client is making
// a request with the apiKey in the Url. includeKeyPEM controls if the key API key is also checked
// and sensitive Private Key PEM data is included in the Order.
func (service *Service) getCertNewestValidOrder(certName string, apiKeyOrKeys string, apiKeyViaUrl bool, includeKeyPEM bool) (orders.Order, *output.JsonError) {
	// if apiKeyOrKeys is blank, definitely unauthorized
	if apiKeyOrKeys == "" {
		service.logger.Debug(errBlankApiKey)
		return orders.Order{}, output.JsonErrUnauthorized
	}

	// get the cert's newest valid order from storage
	order, err := service.storage.GetCertNewestValidOrderByName(certName)
	if err != nil {
		// special error case for no record found
		if errors.Is(err, storage.ErrNoRecord) {
			service.logger.Debug(err)
			// not yet authorized
			return orders.Order{}, output.JsonErrNotFound(nil)
		} else {
			service.logger.Error(err)
			// not yet authorized
			return orders.Order{}, output.JsonErrStorageGeneric(nil)
		}
	}

	// separate the apiKeys
	apiKeys := strings.Split(apiKeyOrKeys, ".")

	// always check cert api key
	certApiKey := apiKeys[0]

	// if apiKey came from URL, and cert does not support this, error
	if apiKeyViaUrl && !order.Certificate.ApiKeyViaUrl {
		service.logger.Debug(errApiKeyFromUrlDisallowed)
		return orders.Order{}, output.JsonErrUnauthorized
	}

	// verify cert apikey matches cert's cert apikey (new or old)
	if (certApiKey != order.Certificate.ApiKey) && (certApiKey != order.Certificate.ApiKeyNew) {
		service.logger.Debug(errWrongApiKey)
		return orders.Order{}, output.JsonErrUnauthorized
	}

	// pem cant be blank
	if order.Pem == nil || *order.Pem == "" {
		service.logger.Debug(errNoPem)
		return orders.Order{}, output.JsonErrNotFound(errNoPem)
	}

	// next steps depend on if we're also checking the key API key
	// if NOT also accessing the key,
	if !includeKeyPEM {
		// if not checking key API key, verify apiKeyOrKeys was only 1 key
		if len(apiKeys) != 1 {
			return orders.Order{}, output.JsonErrUnauthorized
		}

		// if only checking cert key, nuke key private data as a safety precaution
		order.FinalizedKey.Pem = ""

		// before return, update cert last access, dont fail our though if this step fails, just log error
		err = service.storage.PutCertLastAccess(order.Certificate.ID, time.Now().Unix())
		if err != nil {
			service.logger.Errorf("download: failed to update cert (id: %d) last access time (%s)", order.Certificate.ID, err)
		}

		// return order without private key pem
		return order, nil
	}

	//
	// also check the key's API key
	//

	// error if not exactly 2 apiKeys
	if len(apiKeys) != 2 {
		return orders.Order{}, output.JsonErrUnauthorized
	}

	// check key API key
	keyApiKey := apiKeys[1]

	// confirm the private key is valid
	if order.FinalizedKey == nil {
		service.logger.Debug(errFinalizedKeyMissing)
		return orders.Order{}, output.JsonErrNotFound(errFinalizedKeyMissing)
	}

	// if key api key is disabled via API, error
	if order.FinalizedKey.ApiKeyDisabled {
		service.logger.Debug(errApiDisabled)
		return orders.Order{}, output.JsonErrUnauthorized
	}

	// if apiKey came from URL, and key does not support this, error
	if apiKeyViaUrl && !order.FinalizedKey.ApiKeyViaUrl {
		service.logger.Debug(errApiKeyFromUrlDisallowed)
		return orders.Order{}, output.JsonErrUnauthorized
	}

	// validate the apiKey for the private key is correct (new or old)
	if (keyApiKey != order.FinalizedKey.ApiKey) && (keyApiKey != order.FinalizedKey.ApiKeyNew) {
		service.logger.Debug(errWrongApiKey)
		return orders.Order{}, output.JsonErrUnauthorized
	}

	// before return, update cert AND KEY last access, dont fail our though if this step fails, just log error
	nowT := time.Now()
	err = service.storage.PutCertLastAccess(order.Certificate.ID, nowT.Unix())
	if err != nil {
		service.logger.Errorf("download: failed to update cert (id: %d) last access time (%s)", order.Certificate.ID, err)
	}
	err = service.storage.PutKeyLastAccess(order.FinalizedKey.ID, nowT.Unix())
	if err != nil {
		service.logger.Errorf("download: failed to update key (id: %d) last access time (%s)", order.FinalizedKey.ID, err)
	}

	// return order
	return order, nil
}
