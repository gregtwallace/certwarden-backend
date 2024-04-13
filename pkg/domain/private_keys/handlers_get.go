package private_keys

import (
	"certwarden-backend/pkg/domain/private_keys/key_crypto"
	"certwarden-backend/pkg/output"
	"certwarden-backend/pkg/pagination_sort"
	"certwarden-backend/pkg/storage"
	"certwarden-backend/pkg/validation"
	"errors"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// allKeysResponse provides the json response struct
// to answer a query for a portion of the keys
type allKeysResponse struct {
	output.JsonResponse
	TotalKeys int                  `json:"total_records"`
	Keys      []KeySummaryResponse `json:"private_keys"`
}

// GetAllKeys returns all of the private keys in storage as JSON
func (service *Service) GetAllKeys(w http.ResponseWriter, r *http.Request) *output.Error {
	// parse pagination and sorting
	query := pagination_sort.ParseRequestToQuery(r)

	// get keys from storage
	keys, totalRows, err := service.storage.GetAllKeys(query)
	if err != nil {
		service.logger.Error(err)
		return output.ErrStorageGeneric
	}

	// populate keysSummaries for output
	outputKeys := []KeySummaryResponse{}
	for i := range keys {
		outputKeys = append(outputKeys, keys[i].SummaryResponse())
	}

	// write response
	response := &allKeysResponse{}
	response.StatusCode = http.StatusOK
	response.Message = "ok"
	response.TotalKeys = totalRows
	response.Keys = outputKeys

	err = service.output.WriteJSON(w, response)
	if err != nil {
		service.logger.Errorf("failed to write json (%s)", err)
		return output.ErrWriteJsonError
	}

	return nil
}

type privateKeyResponse struct {
	output.JsonResponse
	PrivateKey keyDetailedResponse `json:"private_key"`
}

// GetOneKey returns a single private key as JSON
func (service *Service) GetOneKey(w http.ResponseWriter, r *http.Request) *output.Error {
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
	key, outErr := service.getKey(id)
	if outErr != nil {
		return outErr
	}

	// write response
	response := &privateKeyResponse{}
	response.StatusCode = http.StatusOK
	response.Message = "ok"
	response.PrivateKey = key.detailedResponse()

	// return response to client
	err = service.output.WriteJSON(w, response)
	if err != nil {
		service.logger.Errorf("failed to write json (%s)", err)
		return output.ErrWriteJsonError
	}

	return nil
}

// DownloadOneKey returns the pem for a single key to the client
func (service *Service) DownloadOneKey(w http.ResponseWriter, r *http.Request) *output.Error {
	// params
	idParam := httprouter.ParamsFromContext(r.Context()).ByName("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// basic check
	if !validation.IsIdExistingValidRange(id) {
		service.logger.Debug(ErrIdBad)
		return output.ErrValidationFailed
	}

	// get the key from storage (and validate id)
	key, err := service.storage.GetOneKeyById(id)
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

	// return pem file to client
	service.output.WritePem(w, r, key)

	return nil
}

// new private key options
// used to return info about valid options when making a new key
type newKeyOptions struct {
	output.JsonResponse
	PrivateKeyOptions struct {
		KeyAlgorithms []key_crypto.Algorithm `json:"key_algorithms"`
	} `json:"private_key_options"`
}

// GetNewKeyOptions returns configuration options for a new private key as JSON
func (service *Service) GetNewKeyOptions(w http.ResponseWriter, r *http.Request) *output.Error {
	// write response
	response := &newKeyOptions{}
	response.StatusCode = http.StatusOK
	response.Message = "ok"
	response.PrivateKeyOptions.KeyAlgorithms = key_crypto.ListOfAlgorithms()

	// return response to client
	err := service.output.WriteJSON(w, response)
	if err != nil {
		service.logger.Errorf("failed to write json (%s)", err)
		return output.ErrWriteJsonError
	}

	return nil
}
