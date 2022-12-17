package certificates

import (
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
	Certificates      []certificateSummaryResponse `json:"certificates"`
	TotalCertificates int                          `json:"total_records"`
}

// GetAllCertificates fetches all certs from storage and outputs them as JSON
func (service *Service) GetAllCerts(w http.ResponseWriter, r *http.Request) (err error) {
	// parse pagination and sorting
	query := pagination_sort.ParseRequestToQuery(r)

	// get certs from storage
	certs, totalRows, err := service.storage.GetAllCerts(query)
	if err != nil {
		service.logger.Error(err)
		return output.ErrStorageGeneric
	}

	// make response (for json output)
	response := allKeysResponse{
		TotalCertificates: totalRows,
	}

	// populate cert summaries for output
	for i := range certs {
		response.Certificates = append(response.Certificates, certs[i].summaryResponse())
	}

	// return response to client
	_, err = service.output.WriteJSON(w, http.StatusOK, response, "all_certificates")
	if err != nil {
		service.logger.Error(err)
		return output.ErrWriteJsonFailed
	}

	return nil
}

// GetOneCert is an http handler that returns one Certificate based on its unique id in the
// form of JSON written to w
func (service *Service) GetOneCert(w http.ResponseWriter, r *http.Request) (err error) {
	// get id from param
	idParam := httprouter.ParamsFromContext(r.Context()).ByName("certid")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// if id is new, provide some info
	if validation.IsIdNew(id) {
		return service.GetNewCertOptions(w, r)
	}

	// get from storage
	cert, err := service.GetCertificate(id)
	if err != nil {
		return err
	}

	// return response to client
	_, err = service.output.WriteJSON(w, http.StatusOK, cert.detailedResponse(service.https || service.devMode), "certificate")
	if err != nil {
		service.logger.Error(err)
		return output.ErrWriteJsonFailed
	}

	return nil
}

// GetNewCertOptions is an http handler that returns information the client GUI needs to properly
// present options when the user is creating a certificate
func (service *Service) GetNewCertOptions(w http.ResponseWriter, r *http.Request) (err error) {
	// certificate options / info to assist client with new certificate posting
	newCertOptions := newCertOptions{}

	// available accounts
	accounts, err := service.accounts.GetUsableAccounts()
	if err != nil {
		service.logger.Error(err)
		return output.ErrStorageGeneric
	}

	for i := range accounts {
		newCertOptions.UsableAccounts = append(newCertOptions.UsableAccounts, accounts[i].SummaryResponse())
	}

	// available private keys
	keys, err := service.keys.AvailableKeys()
	if err != nil {
		service.logger.Error(err)
		return output.ErrStorageGeneric
	}

	for i := range keys {
		newCertOptions.AvailableKeys = append(newCertOptions.AvailableKeys, keys[i].SummaryResponse())
	}

	// available challenge methods
	newCertOptions.AvailableChallengeMethods = service.challenges.ListOfMethods()

	// return response to client
	_, err = service.output.WriteJSON(w, http.StatusOK, newCertOptions, "certificate_options")
	if err != nil {
		service.logger.Error(err)
		return output.ErrWriteJsonFailed
	}

	return nil
}
