package acme_accounts

import (
	"legocerthub-backend/pkg/domain/acme_servers"
	"legocerthub-backend/pkg/domain/private_keys"
	"legocerthub-backend/pkg/output"
	"legocerthub-backend/pkg/pagination_sort"
	"legocerthub-backend/pkg/validation"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// accountsResponse provides the json response struct
// to answer a query for a portion of the accounts
type accountsResponse struct {
	output.JsonResponse
	TotalAccounts int                      `json:"total_records"`
	Accounts      []AccountSummaryResponse `json:"acme_accounts"`
}

// GetAllAccounts is an http handler that returns all acme accounts in the form of JSON written to w
func (service *Service) GetAllAccounts(w http.ResponseWriter, r *http.Request) *output.Error {
	// parse pagination and sorting
	query := pagination_sort.ParseRequestToQuery(r)

	// get all from storage
	accounts, totalRows, err := service.storage.GetAllAccounts(query)
	if err != nil {
		service.logger.Error(err)
		return output.ErrStorageGeneric
	}

	// populate account summaries for output
	outputAccounts := []AccountSummaryResponse{}
	for i := range accounts {
		outputAccounts = append(outputAccounts, accounts[i].SummaryResponse())
	}

	// write response
	response := &accountsResponse{}
	response.StatusCode = http.StatusOK
	response.Message = "ok"
	response.TotalAccounts = totalRows
	response.Accounts = outputAccounts

	err = service.output.WriteJSON(w, response)
	if err != nil {
		service.logger.Errorf("failed to write json (%s)", err)
		return output.ErrWriteJsonError
	}

	return nil
}

type accountResponse struct {
	output.JsonResponse
	Account accountDetailedResponse `json:"acme_account"`
}

// GetOneAccount is an http handler that returns one acme account based on its unique id in the
// form of JSON written to w
func (service *Service) GetOneAccount(w http.ResponseWriter, r *http.Request) *output.Error {
	// get id from param
	idParam := httprouter.ParamsFromContext(r.Context()).ByName("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// if id is new, provide some info
	if validation.IsIdNew(id) {
		return service.GetNewAccountOptions(w, r)
	}

	// get from storage
	account, outErr := service.getAccount(id)
	if outErr != nil {
		return outErr
	}

	detailedResp, err := account.detailedResponse(service)
	if err != nil {
		service.logger.Errorf("failed to generate account summary response (%s)", err)
		return output.ErrInternal
	}

	// write response
	response := &accountResponse{}
	response.StatusCode = http.StatusOK
	response.Message = "ok"
	response.Account = detailedResp

	// return response to client
	err = service.output.WriteJSON(w, response)
	if err != nil {
		service.logger.Errorf("failed to write json (%s)", err)
		return output.ErrWriteJsonError
	}

	return nil
}

// new account info
// used to return info about valid options when making a new account
type newAccountOptions struct {
	output.JsonResponse
	AcmeAccountOptions struct {
		AcmeServers   []acme_servers.ServerSummaryResponse `json:"acme_servers"`
		AvailableKeys []private_keys.KeySummaryResponse    `json:"private_keys"`
	} `json:"acme_account_options"`
}

// GetNewAccountOptions is an http handler that returns information the client GUI needs to properly
// present options when the user is creating an account
func (service *Service) GetNewAccountOptions(w http.ResponseWriter, r *http.Request) *output.Error {
	// acme servers
	acmeServers, err := service.acmeServerService.ListAllServersSummaries()
	if err != nil {
		service.logger.Error(err)
		return output.ErrInternal
	}

	// available private keys
	rawKeys, err := service.keys.AvailableKeys()
	if err != nil {
		service.logger.Error(err)
		return output.ErrStorageGeneric
	}

	keys := []private_keys.KeySummaryResponse{}
	for i := range rawKeys {
		keys = append(keys, rawKeys[i].SummaryResponse())
	}

	// write response
	response := &newAccountOptions{}
	response.StatusCode = http.StatusOK
	response.Message = "ok"
	response.AcmeAccountOptions.AcmeServers = acmeServers
	response.AcmeAccountOptions.AvailableKeys = keys

	err = service.output.WriteJSON(w, response)
	if err != nil {
		service.logger.Errorf("failed to write json (%s)", err)
		return output.ErrWriteJsonError
	}

	return nil
}
