package private_keys

import (
	"legocerthub-backend/pkg/domain/private_keys/key_crypto"
	"legocerthub-backend/pkg/output"
	"legocerthub-backend/pkg/pagination_sort"
	"legocerthub-backend/pkg/validation"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// allKeysResponse provides the json response struct
// to answer a query for a portion of the keys
type allKeysResponse struct {
	Keys      []KeySummaryResponse `json:"private_keys"`
	TotalKeys int                  `json:"total_records"`
}

// GetAllKeys returns all of the private keys in storage as JSON
func (service *Service) GetAllKeys(w http.ResponseWriter, r *http.Request) (err error) {
	// parse pagination and sorting
	query := pagination_sort.ParseRequestToQuery(r)

	// get keys from storage
	keys, totalRows, err := service.storage.GetAllKeys(query)
	if err != nil {
		service.logger.Error(err)
		return output.ErrStorageGeneric
	}

	// assemble response
	response := allKeysResponse{
		TotalKeys: totalRows,
	}

	// populate keysSummaries for output
	for i := range keys {
		response.Keys = append(response.Keys, keys[i].SummaryResponse())
	}

	// return response to client
	_, err = service.output.WriteJSON(w, http.StatusOK, response, "all_private_keys")
	if err != nil {
		service.logger.Error(err)
		return output.ErrWriteJsonFailed
	}

	return nil
}

// GetOneKey returns a single private key as JSON
func (service *Service) GetOneKey(w http.ResponseWriter, r *http.Request) (err error) {
	// params
	idParam := httprouter.ParamsFromContext(r.Context()).ByName("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// if id is new, provide some info
	if validation.IsIdNew(id) {
		return service.GetNewKeyOptions(w, r)
	}

	// get the key from storage (and validate id)
	key, err := service.getKey(id)
	if err != nil {
		return err
	}

	// return response to client
	_, err = service.output.WriteJSON(w, http.StatusOK, key.detailedResponse(service.https || service.devMode), "private_key")
	if err != nil {
		service.logger.Error(err)
		return output.ErrWriteJsonFailed
	}

	return nil
}

// GetNewKeyOptions returns configuration options for a new private key as JSON
func (service *Service) GetNewKeyOptions(w http.ResponseWriter, r *http.Request) error {
	newKeyOptions := newKeyOptions{}
	newKeyOptions.KeyAlgorithms = key_crypto.ListOfAlgorithms()

	// return response to client
	_, err := service.output.WriteJSON(w, http.StatusOK, newKeyOptions, "private_key_options")
	if err != nil {
		service.logger.Error(err)
		return output.ErrWriteJsonFailed
	}

	return nil
}
