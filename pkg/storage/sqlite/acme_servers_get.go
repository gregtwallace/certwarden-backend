package sqlite

import (
	"context"
	"fmt"
	"legocerthub-backend/pkg/domain/acme_servers"
	"legocerthub-backend/pkg/pagination_sort"
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
