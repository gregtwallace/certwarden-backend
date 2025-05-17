package certificates

import (
	"certwarden-backend/pkg/domain/acme_accounts"
	"certwarden-backend/pkg/domain/private_keys"
	"certwarden-backend/pkg/domain/private_keys/key_crypto"
	"certwarden-backend/pkg/output"
	"certwarden-backend/pkg/pagination_sort"
	"certwarden-backend/pkg/validation"
	"fmt"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// allCertsResponse provides the json response struct
// to answer a query for a portion of the certs
type allCertsResponse struct {
	output.JsonResponse
	TotalCertificates int                          `json:"total_records"`
	Certificates      []certificateSummaryResponse `json:"certificates"`
}

// GetAllCertificates fetches all certs from storage and outputs them as JSON
func (service *Service) GetAllCerts(w http.ResponseWriter, r *http.Request) *output.JsonError {
	// parse pagination and sorting
	query := pagination_sort.ParseRequestToQuery(r)

	// get certs from storage
	certs, totalRows, err := service.storage.GetAllCerts(query)
	if err != nil {
		service.logger.Error(err)
		return output.JsonErrStorageGeneric(err)
	}

	// populate cert summaries for output
	outputCerts := []certificateSummaryResponse{}
	for i := range certs {
		outputCerts = append(outputCerts, certs[i].summaryResponse())
	}

	// write response
	response := &allCertsResponse{}
	response.StatusCode = http.StatusOK
	response.Message = "ok"
	response.TotalCertificates = totalRows
	response.Certificates = outputCerts

	err = service.output.WriteJSON(w, response)
	if err != nil {
		service.logger.Errorf("failed to write json (%s)", err)
		return output.JsonErrWriteJsonError(err)
	}

	return nil
}

type certificateResponse struct {
	output.JsonResponse
	Certificate certificateDetailedResponse `json:"certificate"`
}

// GetOneCert is an http handler that returns one Certificate based on its unique id in the
// form of JSON written to w
func (service *Service) GetOneCert(w http.ResponseWriter, r *http.Request) *output.JsonError {
	// get id from param
	idParam := httprouter.ParamsFromContext(r.Context()).ByName("certid")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		service.logger.Debug(err)
		return output.JsonErrValidationFailed(err)
	}

	// if id is new, provide some info
	if validation.IsIdNew(id) {
		return service.GetNewCertOptions(w, r)
	}

	// get from storage
	cert, outErr := service.GetCertificate(id)
	if outErr != nil {
		return outErr
	}

	// write response
	response := &certificateResponse{}
	response.StatusCode = http.StatusOK
	response.Message = "ok"
	response.Certificate = cert.detailedResponse()

	err = service.output.WriteJSON(w, response)
	if err != nil {
		service.logger.Errorf("failed to write json (%s)", err)
		return output.JsonErrWriteJsonError(err)
	}

	return nil
}

type newCertOptions struct {
	output.JsonResponse
	CertificateOptions struct {
		AvailableKeys  []private_keys.KeySummaryResponse `json:"private_keys"`
		KeyAlgorithms  []key_crypto.Algorithm            `json:"key_algorithms"`
		UsableAccounts []usableAccount                   `json:"acme_accounts"`
	} `json:"certificate_options"`
}

// usableAccount uses a custom AcmeServer struct to also output valid profile names
type usableAccount struct {
	acme_accounts.AccountSummaryResponse
	AcmeServer struct {
		ID           int               `json:"id"`
		Name         string            `json:"name"`
		DirectoryURL string            `json:"directory_url"`
		IsStaging    bool              `json:"is_staging"`
		Profiles     map[string]string `json:"profiles,omitempty"`
	} `json:"acme_server"`
}

// GetNewCertOptions is an http handler that returns information the client GUI needs to properly
// present options when the user is creating a certificate
func (service *Service) GetNewCertOptions(w http.ResponseWriter, r *http.Request) *output.JsonError {
	// available private keys
	keys, err := service.keys.AvailableKeys()
	if err != nil {
		service.logger.Error(err)
		return output.JsonErrStorageGeneric(err)
	}

	outputKeys := []private_keys.KeySummaryResponse{}
	for i := range keys {
		outputKeys = append(outputKeys, keys[i].SummaryResponse())
	}

	// available accounts
	accounts, err := service.accounts.GetUsableAccounts()
	if err != nil {
		service.logger.Error(err)
		return output.JsonErrStorageGeneric(err)
	}

	outputAccounts := []usableAccount{}
	for i := range accounts {
		acct := usableAccount{}
		acct.AccountSummaryResponse = accounts[i].SummaryResponse()

		// get profiles
		acmeService, err := service.acmeServerService.AcmeService(accounts[i].AcmeServer.ID)
		if err != nil {
			err = fmt.Errorf("failed to retrieve acme service to list profiles (%s)", err)
			service.logger.Error(err)
			return output.JsonErrInternal(err)
		}

		// redo AcmeServer
		acct.AcmeServer.ID = accounts[i].AcmeServer.ID
		acct.AcmeServer.Name = accounts[i].AcmeServer.Name
		acct.AcmeServer.DirectoryURL = accounts[i].AcmeServer.DirectoryURL
		acct.AcmeServer.IsStaging = accounts[i].AcmeServer.IsStaging
		acct.AcmeServer.Profiles = acmeService.Profiles()

		outputAccounts = append(outputAccounts, acct)
	}

	// write response
	response := &newCertOptions{}
	response.StatusCode = http.StatusOK
	response.Message = "ok"
	response.CertificateOptions.AvailableKeys = outputKeys
	response.CertificateOptions.KeyAlgorithms = key_crypto.ListOfAlgorithms()
	response.CertificateOptions.UsableAccounts = outputAccounts

	err = service.output.WriteJSON(w, response)
	if err != nil {
		service.logger.Errorf("failed to write json (%s)", err)
		return output.JsonErrWriteJsonError(err)
	}

	return nil
}
