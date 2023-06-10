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

// serverInformationResponse is the struct that holds data about a Server that
// is returned to a client
type ServerInformationResponse struct {
	ID                      int    `json:"id"`
	Name                    string `json:"name"`
	Description             string `json:"description"`
	DirectoryURL            string `json:"directory_url"`
	IsStaging               bool   `json:"is_staging"`
	TermsOfService          string `json:"terms_of_service"`
	ExternalAccountRequired bool   `json:"external_account_required"`
}

// ListServersInfo returns a slice of information about all configured Servers
func (service *Service) ListServersInfo() ([]ServerInformationResponse, error) {
	// fetch from storage
	servers, _, err := service.storage.GetAllAcmeServers(pagination_sort.QueryAll)
	if err != nil {
		return nil, err
	}

	// convert to info response
	serversInfo := []ServerInformationResponse{}
	for i := range servers {
		// need to access server service for some info
		service, err := service.AcmeService(servers[i].ID)
		if err != nil {
			return nil, err
		}

		// populate info
		serverInfo := ServerInformationResponse{
			ID:                      servers[i].ID,
			Name:                    servers[i].Name,
			Description:             servers[i].Description,
			DirectoryURL:            servers[i].DirectoryURL,
			IsStaging:               servers[i].IsStaging,
			TermsOfService:          service.TosUrl(),
			ExternalAccountRequired: service.RequiresEAB(),
		}

		// append info
		serversInfo = append(serversInfo, serverInfo)
	}

	return serversInfo, nil
}
