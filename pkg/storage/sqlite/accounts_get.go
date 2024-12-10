package sqlite

import (
	"certwarden-backend/pkg/domain/acme_accounts"
	"certwarden-backend/pkg/pagination_sort"
	"certwarden-backend/pkg/storage"
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// GetAllAccounts returns a slice of all of the Accounts in the database
func (store *Storage) GetAllAccounts(q pagination_sort.Query) (accounts []acme_accounts.Account, totalRowCount int, err error) {
	// validate and set sort
	sortField := q.SortField()

	switch sortField {
	// allow these
	case "id":
		sortField = "aa.id"
	case "name":
		sortField = "aa.name"
	case "description":
		sortField = "aa.description"
	case "servername":
		sortField = "aserv.name"
	case "keyname":
		sortField = "pk.name"
	case "status":
		sortField = "aa.status"
	case "email":
		sortField = "aa.email"
	case "is_staging":
		sortField = "aserv.is_staging"
	// default if not in allowed list
	default:
		sortField = "aa.name"
	}

	sort := sortField + " " + q.SortDirection()

	// do query
	ctx, cancel := context.WithTimeout(context.Background(), store.timeout)
	defer cancel()

	// WARNING: SQL Injection is possible if the variables are not properly
	// validated prior to this query being assembled!
	query := fmt.Sprintf(`
	SELECT
		aa.id, aa.name, aa.description, aa.status, aa.email, aa.accepted_tos,
		aa.created_at, aa.updated_at, aa.kid,

		aserv.id, aserv.name, aserv.description, aserv.directory_url, aserv.is_staging, aserv.created_at,
		aserv.updated_at,

		pk.id, pk.name, pk.description, pk.algorithm, pk.pem, pk.api_key, pk.api_key_new,
		pk.api_key_disabled, pk.api_key_via_url, pk.last_access, pk.created_at, pk.updated_at,

		count(*) OVER() AS full_count
	FROM
		acme_accounts aa
		LEFT JOIN acme_servers aserv on (aa.acme_server_id = aserv.id)
		LEFT JOIN private_keys pk on (aa.private_key_id = pk.id)
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

	var allAccounts []acme_accounts.Account
	for rows.Next() {
		var oneAccount accountDb
		err = rows.Scan(
			&oneAccount.id,
			&oneAccount.name,
			&oneAccount.description,
			&oneAccount.status,
			&oneAccount.email,
			&oneAccount.acceptedTos,
			&oneAccount.createdAt,
			&oneAccount.updatedAt,
			&oneAccount.kid,

			&oneAccount.accountServerDb.id,
			&oneAccount.accountServerDb.name,
			&oneAccount.accountServerDb.description,
			&oneAccount.accountServerDb.directoryUrl,
			&oneAccount.accountServerDb.isStaging,
			&oneAccount.accountServerDb.createdAt,
			&oneAccount.accountServerDb.updatedAt,

			&oneAccount.accountKeyDb.id,
			&oneAccount.accountKeyDb.name,
			&oneAccount.accountKeyDb.description,
			&oneAccount.accountKeyDb.algorithmValue,
			&oneAccount.accountKeyDb.pem,
			&oneAccount.accountKeyDb.apiKey,
			&oneAccount.accountKeyDb.apiKeyNew,
			&oneAccount.accountKeyDb.apiKeyDisabled,
			&oneAccount.accountKeyDb.apiKeyViaUrl,
			&oneAccount.accountKeyDb.lastAccess,
			&oneAccount.accountKeyDb.createdAt,
			&oneAccount.accountKeyDb.updatedAt,

			&totalRows,
		)
		if err != nil {
			return nil, 0, err
		}

		// convert to Account
		convertedAccount := oneAccount.toAccount()

		allAccounts = append(allAccounts, convertedAccount)
	}

	return allAccounts, totalRows, nil
}

// GetOneAccountById returns an Account based on its unique id
func (store *Storage) GetOneAccountById(id int) (acme_accounts.Account, error) {
	return store.getOneAccount(id, "")
}

// GetOneAccountByName returns an Account based on its unique name
func (store *Storage) GetOneAccountByName(name string) (acme_accounts.Account, error) {
	return store.getOneAccount(-1, name)
}

// getOneAccount returns an Account based on either its unique id or its unique name
func (store *Storage) getOneAccount(id int, name string) (acme_accounts.Account, error) {
	ctx, cancel := context.WithTimeout(context.Background(), store.timeout)
	defer cancel()

	query := `
	SELECT
		aa.id, aa.name, aa.description, aa.status, aa.email, aa.accepted_tos,
		aa.created_at, aa.updated_at, aa.kid,

		aserv.id, aserv.name, aserv.description, aserv.directory_url, aserv.is_staging, aserv.created_at,
		aserv.updated_at,

		pk.id, pk.name, pk.description, pk.algorithm, pk.pem, pk.api_key, pk.api_key_new, 
		pk.api_key_disabled, pk.api_key_via_url, pk.last_access, pk.created_at, pk.updated_at
	FROM
		acme_accounts aa
		LEFT JOIN acme_servers aserv on (aa.acme_server_id = aserv.id)
		LEFT JOIN private_keys pk on (aa.private_key_id = pk.id)
	WHERE aa.id = $1 OR aa.name = $2
	ORDER BY aa.id`

	row := store.db.QueryRowContext(ctx, query, id, name)

	var oneAccount accountDb

	err := row.Scan(
		&oneAccount.id,
		&oneAccount.name,
		&oneAccount.description,
		&oneAccount.status,
		&oneAccount.email,
		&oneAccount.acceptedTos,
		&oneAccount.createdAt,
		&oneAccount.updatedAt,
		&oneAccount.kid,

		&oneAccount.accountServerDb.id,
		&oneAccount.accountServerDb.name,
		&oneAccount.accountServerDb.description,
		&oneAccount.accountServerDb.directoryUrl,
		&oneAccount.accountServerDb.isStaging,
		&oneAccount.accountServerDb.createdAt,
		&oneAccount.accountServerDb.updatedAt,

		&oneAccount.accountKeyDb.id,
		&oneAccount.accountKeyDb.name,
		&oneAccount.accountKeyDb.description,
		&oneAccount.accountKeyDb.algorithmValue,
		&oneAccount.accountKeyDb.pem,
		&oneAccount.accountKeyDb.apiKey,
		&oneAccount.accountKeyDb.apiKeyNew,
		&oneAccount.accountKeyDb.apiKeyDisabled,
		&oneAccount.accountKeyDb.apiKeyViaUrl,
		&oneAccount.accountKeyDb.lastAccess,
		&oneAccount.accountKeyDb.createdAt,
		&oneAccount.accountKeyDb.updatedAt,
	)

	if err != nil {
		// if no record exists
		if errors.Is(err, sql.ErrNoRows) {
			err = storage.ErrNoRecord
		}
		return acme_accounts.Account{}, err
	}

	return oneAccount.toAccount(), nil
}
