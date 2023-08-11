package acme_servers

import (
	"legocerthub-backend/pkg/output"
	"legocerthub-backend/pkg/pagination_sort"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// allAcmeServersResponse provides the json response struct
// to answer a query for a portion of the ACME servers
type allAcmeServersResponse struct {
	Servers   []ServerSummaryResponse `json:"acme_servers"`
	TotalKeys int                     `json:"total_records"`
}

// GetAllServers returns all of the ACME servers
func (service *Service) GetAllServers(w http.ResponseWriter, r *http.Request) (err error) {
	// parse pagination and sorting
	query := pagination_sort.ParseRequestToQuery(r)

	// get keys from storage
	servers, totalRows, err := service.storage.GetAllAcmeServers(query)
	if err != nil {
		service.logger.Error(err)
		return output.ErrStorageGeneric
	}

	// assemble response
	response := allAcmeServersResponse{
		TotalKeys: totalRows,
	}

	// populate keysSummaries for output
	for i := range servers {
		summary, err := servers[i].summaryResponse(service)
		if err != nil {
			return err
		}

		response.Servers = append(response.Servers, summary)
	}

	// return response to client
	err = service.output.WriteJSON(w, http.StatusOK, response, "all_acme_servers")
	if err != nil {
		return err
	}

	return nil
}

// GetOneServer returns a single acme server
func (service *Service) GetOneServer(w http.ResponseWriter, r *http.Request) (err error) {
	// params
	idParam := httprouter.ParamsFromContext(r.Context()).ByName("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// get the server from storage (and validate id)
	server, err := service.getServer(id)
	if err != nil {
		return err
	}

	// make detailed response
	detailedResp, err := server.detailedResponse(service)
	if err != nil {
		return err
	}

	// return response to client
	err = service.output.WriteJSON(w, http.StatusOK, detailedResp, "acme_server")
	if err != nil {
		return err
	}

	return nil
}
