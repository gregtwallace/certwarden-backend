package acme_servers

import "legocerthub-backend/pkg/pagination_sort"

// AcmeServerValid returns true if the specified acme server id is valid
func (service *Service) AcmeServerValid(acmeServerId int) bool {
	// get available keys list
	acmeServers, _, err := service.storage.GetAllAcmeServers(pagination_sort.QueryAll)
	if err != nil {
		return false
	}

	// verify specified id is in the available list
	for i := range acmeServers {
		if acmeServers[i].ID == acmeServerId {
			return true
		}
	}

	return false
}
