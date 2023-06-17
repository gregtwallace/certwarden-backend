package acme_servers

import (
	"fmt"
	"legocerthub-backend/pkg/acme"
	"legocerthub-backend/pkg/pagination_sort"
)

// Server is the struct for an ACME Server
type Server struct {
	ID           int
	Name         string
	Description  string
	DirectoryURL string
	IsStaging    bool
	CreatedAt    int
	UpdatedAt    int
}

// AcmeService returns the service for a specific ACME Server specified
// by its ID. If the ID is not valid, an error is returned.
func (service *Service) AcmeService(id int) (*acme.Service, error) {
	// check id is valid
	acmeService, exist := service.acmeServers[id]
	// if not valid, or if the service pointer is nil
	if !exist || acmeService == nil {
		return nil, fmt.Errorf("specified acme service id (%d) is not valid", id)
	}

	// return valid acme service
	return acmeService, nil
}

// serverSummaryResponse contains abbreviated details about an ACME server
type ServerSummaryResponse struct {
	// static
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	DirectoryURL string `json:"directory_url"`
	IsStaging    bool   `json:"is_staging"`
	// from remote server
	ExternalAccountRequired bool   `json:"external_account_required"`
	TermsOfService          string `json:"terms_of_service"`
}

func (serv Server) summaryResponse(service *Service) (ServerSummaryResponse, error) {
	acmeService, err := service.AcmeService(serv.ID)
	if err != nil {
		return ServerSummaryResponse{}, err
	}

	return ServerSummaryResponse{
		ID:                      serv.ID,
		Name:                    serv.Name,
		Description:             serv.Description,
		DirectoryURL:            serv.DirectoryURL,
		IsStaging:               serv.IsStaging,
		ExternalAccountRequired: acmeService.RequiresEAB(),
		TermsOfService:          acmeService.TosUrl(),
	}, nil
}

// serverDetailedResponse contains full details about an ACME server
type serverDetailedResponse struct {
	ServerSummaryResponse
	CreatedAt int `json:"created_at"`
	UpdatedAt int `json:"updated_at"`
}

func (serv Server) detailedResponse(service *Service) (serverDetailedResponse, error) {
	summaryResp, err := serv.summaryResponse(service)
	if err != nil {
		return serverDetailedResponse{}, err
	}

	return serverDetailedResponse{
		ServerSummaryResponse: summaryResp,
		CreatedAt:             serv.CreatedAt,
		UpdatedAt:             serv.UpdatedAt,
	}, nil
}

// ListAllServersSummaries returns a slice of summaries about all configured Servers
func (service *Service) ListAllServersSummaries() ([]ServerSummaryResponse, error) {
	// fetch from storage
	servers, _, err := service.storage.GetAllAcmeServers(pagination_sort.QueryAll)
	if err != nil {
		return nil, err
	}

	// convert to summary responses
	serverSummaries := []ServerSummaryResponse{}
	for i := range servers {
		// populate summary
		serverSummary, err := servers[i].summaryResponse(service)
		if err != nil {
			return nil, err
		}

		// append summary
		serverSummaries = append(serverSummaries, serverSummary)
	}

	return serverSummaries, nil
}
