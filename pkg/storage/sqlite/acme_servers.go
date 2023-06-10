package sqlite

import "legocerthub-backend/pkg/domain/acme_servers"

// acmeServerDb is a single acme server, as database table fields
// corresponds to acme_servers.Server
type acmeServerDb struct {
	id           int
	name         string
	description  string
	directoryUrl string
	isStaging    bool
	createdAt    int
	updatedAt    int
}

// toServer maps the database acme server info to the acme_servers
// Server object
func (serv acmeServerDb) toServer() acme_servers.Server {
	return acme_servers.Server{
		ID:           serv.id,
		Name:         serv.name,
		Description:  serv.description,
		DirectoryURL: serv.directoryUrl,
		IsStaging:    serv.isStaging,
		CreatedAt:    serv.createdAt,
		UpdatedAt:    serv.updatedAt,
	}
}
