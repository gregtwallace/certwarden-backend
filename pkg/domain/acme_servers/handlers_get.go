package acme_servers

import (
	"certwarden-backend/pkg/output"
	"certwarden-backend/pkg/pagination_sort"
	"fmt"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// acmeServersResponse provides the json response struct
// to answer a query for a portion of the ACME servers
type acmeServersResponse struct {
	output.JsonResponse
	TotalServers int                     `json:"total_records"`
	Servers      []ServerSummaryResponse `json:"acme_servers"`
}

// GetAllServers returns all of the ACME servers
func (service *Service) GetAllServers(w http.ResponseWriter, r *http.Request) *output.JsonError {
	// parse pagination and sorting
	query := pagination_sort.ParseRequestToQuery(r)

	// get from storage
	servers, totalRows, err := service.storage.GetAllAcmeServers(query)
	if err != nil {
		service.logger.Error(err)
		return output.JsonErrStorageGeneric(err)
	}

	// populate summaries for output
	scmeServers := []ServerSummaryResponse{}
	for i := range servers {
		summary, err := servers[i].summaryResponse(service)
		if err != nil {
			err = fmt.Errorf("failed to generate server summary response (%s)", err)
			service.logger.Error(err)
			return output.JsonErrInternal(err)
		}

		scmeServers = append(scmeServers, summary)
	}

	// write response
	response := &acmeServersResponse{}
	response.StatusCode = http.StatusOK
	response.Message = "ok"
	response.TotalServers = totalRows
	response.Servers = scmeServers

	err = service.output.WriteJSON(w, response)
	if err != nil {
		service.logger.Errorf("failed to write json (%s)", err)
		return output.JsonErrWriteJsonError(err)
	}

	return nil
}

type acmeServerResponse struct {
	output.JsonResponse
	Server serverDetailedResponse `json:"acme_server"`
}

// GetOneServer returns a single acme server
func (service *Service) GetOneServer(w http.ResponseWriter, r *http.Request) *output.JsonError {
	// params
	idParam := httprouter.ParamsFromContext(r.Context()).ByName("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		service.logger.Debug(err)
		return output.JsonErrValidationFailed(err)
	}

	// get the server from storage (and validate id)
	server, outErr := service.getServer(id)
	if outErr != nil {
		return outErr
	}

	// make detailed response
	detailedResp, err := server.detailedResponse(service)
	if err != nil {
		err = fmt.Errorf("failed to generate server summary response (%s)", err)
		service.logger.Error(err)
		return output.JsonErrInternal(err)
	}

	// write response
	response := &acmeServerResponse{}
	response.StatusCode = http.StatusOK
	response.Message = "ok"
	response.Server = detailedResp

	// return response to client
	err = service.output.WriteJSON(w, response)
	if err != nil {
		service.logger.Errorf("failed to write json (%s)", err)
		return output.JsonErrWriteJsonError(err)
	}

	return nil
}
