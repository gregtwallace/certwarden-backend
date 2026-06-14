package storage

import (
	"certwarden-backend/pkg/domain/acme_servers"
	"time"
)

// acmeServerDb is a single acme server, as database table fields
// corresponds to acme_servers.Server
type acmeServerDb struct {
	id           int
	name         string
	description  string
	directoryUrl string
	isStaging    bool
	createdAt    int64
	updatedAt    int64
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
		CreatedAt:    time.Unix(serv.createdAt, 0),
		UpdatedAt:    time.Unix(serv.updatedAt, 0),
	}
}
