package sqlite

import (
	"certwarden-backend/pkg/domain/acme_servers"
	"certwarden-backend/pkg/pagination_sort"
	"certwarden-backend/pkg/storage"
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// GetAllAcmeServers returns a slice of all of the ACME Servers in the database
func (store *Storage) GetAllAcmeServers(q pagination_sort.Query) (accounts []acme_servers.Server, totalRowCount int, err error) {
	// validate and set sort
	sortField := q.SortField()

	switch sortField {
	// allow these
	case "id":
		sortField = "id"
	case "name":
		sortField = "name"
	case "description":
		sortField = "description"
	case "is_staging":
		sortField = "is_staging"
	// default if not in allowed list
	default:
		sortField = "name"
	}

	sort := sortField + " " + q.SortDirection()

	// do query
	ctx, cancel := context.WithTimeout(context.Background(), store.timeout)
	defer cancel()

	// WARNING: SQL Injection is possible if the variables are not properly
	// validated prior to this query being assembled!
	query := fmt.Sprintf(`
	SELECT
		aserv.id, aserv.name, aserv.description, aserv.directory_url, aserv.is_staging, aserv.created_at,
		aserv.updated_at,

		count(*) OVER() AS full_count
	FROM
		acme_servers aserv
	ORDER BY
		%s
	LIMIT
		$1
	OFFSET
		$2
	`, sort)

	rows, err := store.db.QueryContext(ctx, query,
		q.Limit(),
		q.Offset(),
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	// for total row count
	var totalRows int

	var allServers []acme_servers.Server
	for rows.Next() {
		var oneServer acmeServerDb
		err = rows.Scan(
			&oneServer.id,
			&oneServer.name,
			&oneServer.description,
			&oneServer.directoryUrl,
			&oneServer.isStaging,
			&oneServer.createdAt,
			&oneServer.updatedAt,

			&totalRows,
		)
		if err != nil {
			return nil, 0, err
		}

		// convert to Server and sppend
		allServers = append(allServers, oneServer.toServer())
	}

	return allServers, totalRows, nil
}

// GetOneServerById returns a Server based on unique id
func (store *Storage) GetOneServerById(acmeServerId int) (acme_servers.Server, error) {
	return store.dbGetOneServer(acmeServerId, "")
}

// GetOneServerByName returns a Server based on unique name
func (store *Storage) GetOneServerByName(name string) (acme_servers.Server, error) {
	return store.dbGetOneServer(-1, name)
}

// dbGetOneServer returns a Server based on unique id or unique name
func (store Storage) dbGetOneServer(acmeServerId int, name string) (acme_servers.Server, error) {
	ctx, cancel := context.WithTimeout(context.Background(), store.timeout)
	defer cancel()

	query := `
	SELECT
		aserv.id, aserv.name, aserv.description, aserv.directory_url, aserv.is_staging, aserv.created_at,
		aserv.updated_at
	FROM
		acme_servers aserv
	WHERE
		aserv.id = $1
		OR
		aserv.name = $2
	`

	row := store.db.QueryRowContext(ctx, query, acmeServerId, name)

	var oneServerDb acmeServerDb
	err := row.Scan(
		&oneServerDb.id,
		&oneServerDb.name,
		&oneServerDb.description,
		&oneServerDb.directoryUrl,
		&oneServerDb.isStaging,
		&oneServerDb.createdAt,
		&oneServerDb.updatedAt,
	)

	if err != nil {
		// if no record exists
		if errors.Is(err, sql.ErrNoRows) {
			err = storage.ErrNoRecord
		}
		return acme_servers.Server{}, err
	}

	// convert to Server
	oneServer := oneServerDb.toServer()

	return oneServer, nil
}
