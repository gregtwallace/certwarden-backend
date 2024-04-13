package sqlite

import (
	"certwarden-backend/pkg/domain/orders"
	"certwarden-backend/pkg/pagination_sort"
	"certwarden-backend/pkg/storage"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// GetAllValidCurrentOrders fetches each cert's most recent valid order, if the cert currently has a valid order.
// This is used for a frontend dashboard.
func (store *Storage) GetAllValidCurrentOrders(q pagination_sort.Query) (orders []orders.Order, totalRowCount int, err error) {
	// validate and set sort
	sortField := q.SortField()
	switch sortField {
	case "id":
		sortField = "c.id"
	case "name":
		sortField = "c.name"
	case "subject":
		sortField = "c.subject"
	case "valid_to":
		sortField = "ao.valid_to"
	default:
		sortField = "ao.valid_to"
	}

	sort := sortField + " " + q.SortDirection()

	// query
	ctx, cancel := context.WithTimeout(context.Background(), store.timeout)
	defer cancel()

	// WARNING: SQL Injection is possible if the variables are not properly
	// validated prior to this query being assembled!
	query := fmt.Sprintf(`
	SELECT
		/* order */
		ao.id, ao.acme_location, ao.status, ao.known_revoked, ao.error, ao.expires, ao.dns_identifiers, 
		ao.authorizations, ao.finalize, ao.certificate_url, ao.valid_from, ao.valid_to, ao.created_at,
		ao.updated_at, 

		/* order's cert */
		c.id, c.name, c.description, c.subject, c.subject_alts,
		c.csr_org, c.csr_ou, c.csr_country, c.csr_state, c.csr_city, c.csr_extra_extensions, c.created_at, c.updated_at,
		c.api_key, c.api_key_new, c.api_key_via_url, c.post_processing_command, c.post_processing_environment,
		c.post_processing_client_key,
		
		/* cert's key */
		ck.id, ck.name, ck.description, ck.algorithm, ck.pem, ck.api_key, ck.api_key_new,
		ck.api_key_disabled, ck.api_key_via_url, ck.created_at, ck.updated_at,

		/* cert's account */
		ca.id, ca.name, ca.description, ca.status, ca.email, ca.accepted_tos,
		ca.created_at, ca.updated_at, ca.kid,

		/* cert's account's server */
		aserv.id, aserv.name, aserv.description, aserv.directory_url, aserv.is_staging, aserv.created_at,
		aserv.updated_at,

		/* cert's account's key */
		ak.id, ak.name, ak.description, ak.algorithm, ak.pem, ak.api_key, ak.api_key_new,
		ak.api_key_disabled, ak.api_key_via_url, ak.created_at, ak.updated_at,

		/* finalized key */
		COALESCE(fk.id, -2), COALESCE(fk.name, 'null'), COALESCE(fk.description, 'null'), 
		COALESCE(fk.algorithm, 'null'), COALESCE(fk.pem, 'null'), COALESCE(fk.api_key, 'null'), 
		COALESCE(fk.api_key_new, 'null'), COALESCE(fk.api_key_disabled, false),
		COALESCE(fk.api_key_via_url, false), COALESCE(fk.created_at, -2), COALESCE(fk.updated_at, -2),
		
		count(*) OVER() AS full_count
	FROM
		acme_orders ao
		LEFT JOIN certificates c on (ao.certificate_id = c.id)
		LEFT JOIN private_keys ck on (c.private_key_id = ck.id)
		LEFT JOIN acme_accounts ca on (c.acme_account_id = ca.id)
		LEFT JOIN acme_servers aserv on (ca.acme_server_id = aserv.id)
		LEFT JOIN private_keys ak on (ca.private_key_id = ak.id)
		LEFT JOIN private_keys fk on (ao.finalized_key_id = fk.id)
	WHERE 
		ao.status = "valid"
		AND
		ao.known_revoked = 0
		AND
		ao.valid_to > $1
		AND
		ao.pem NOT NULL
		AND
		ao.certificate_id IS NOT NULL
	GROUP BY
		ao.certificate_id
	HAVING
		MAX(ao.valid_to)
	ORDER BY
		%s
	LIMIT
		$2
	OFFSET
		$3
	`,
		sort)

	// get records
	rows, err := store.db.QueryContext(ctx, query,
		time.Now().Unix(),
		q.Limit(),
		q.Offset(),
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	// for total row count
	var totalRows int

	for rows.Next() {
		var oneOrder orderDb

		err = rows.Scan(
			&oneOrder.id,
			&oneOrder.location,
			&oneOrder.status,
			&oneOrder.knownRevoked,
			&oneOrder.err,
			&oneOrder.expires,
			&oneOrder.dnsIdentifiers,
			&oneOrder.authorizations,
			&oneOrder.finalize,
			&oneOrder.certificateUrl,
			&oneOrder.validFrom,
			&oneOrder.validTo,
			&oneOrder.createdAt,
			&oneOrder.updatedAt,

			&oneOrder.certificate.id,
			&oneOrder.certificate.name,
			&oneOrder.certificate.description,
			&oneOrder.certificate.subject,
			&oneOrder.certificate.subjectAltNames,
			&oneOrder.certificate.organization,
			&oneOrder.certificate.organizationalUnit,
			&oneOrder.certificate.country,
			&oneOrder.certificate.state,
			&oneOrder.certificate.city,
			&oneOrder.certificate.csrExtraExtensions,
			&oneOrder.certificate.createdAt,
			&oneOrder.certificate.updatedAt,
			&oneOrder.certificate.apiKey,
			&oneOrder.certificate.apiKeyNew,
			&oneOrder.certificate.apiKeyViaUrl,
			&oneOrder.certificate.postProcessingCommand,
			&oneOrder.certificate.postProcessingEnvironment,
			&oneOrder.certificate.postProcessingClientKeyB64,

			&oneOrder.certificate.certificateKeyDb.id,
			&oneOrder.certificate.certificateKeyDb.name,
			&oneOrder.certificate.certificateKeyDb.description,
			&oneOrder.certificate.certificateKeyDb.algorithmValue,
			&oneOrder.certificate.certificateKeyDb.pem,
			&oneOrder.certificate.certificateKeyDb.apiKey,
			&oneOrder.certificate.certificateKeyDb.apiKeyNew,
			&oneOrder.certificate.certificateKeyDb.apiKeyDisabled,
			&oneOrder.certificate.certificateKeyDb.apiKeyViaUrl,
			&oneOrder.certificate.certificateKeyDb.createdAt,
			&oneOrder.certificate.certificateKeyDb.updatedAt,

			&oneOrder.certificate.certificateAccountDb.id,
			&oneOrder.certificate.certificateAccountDb.name,
			&oneOrder.certificate.certificateAccountDb.description,
			&oneOrder.certificate.certificateAccountDb.status,
			&oneOrder.certificate.certificateAccountDb.email,
			&oneOrder.certificate.certificateAccountDb.acceptedTos,
			&oneOrder.certificate.certificateAccountDb.createdAt,
			&oneOrder.certificate.certificateAccountDb.updatedAt,
			&oneOrder.certificate.certificateAccountDb.kid,

			&oneOrder.certificate.certificateAccountDb.accountServerDb.id,
			&oneOrder.certificate.certificateAccountDb.accountServerDb.name,
			&oneOrder.certificate.certificateAccountDb.accountServerDb.description,
			&oneOrder.certificate.certificateAccountDb.accountServerDb.directoryUrl,
			&oneOrder.certificate.certificateAccountDb.accountServerDb.isStaging,
			&oneOrder.certificate.certificateAccountDb.accountServerDb.createdAt,
			&oneOrder.certificate.certificateAccountDb.accountServerDb.updatedAt,

			&oneOrder.certificate.certificateAccountDb.accountKeyDb.id,
			&oneOrder.certificate.certificateAccountDb.accountKeyDb.name,
			&oneOrder.certificate.certificateAccountDb.accountKeyDb.description,
			&oneOrder.certificate.certificateAccountDb.accountKeyDb.algorithmValue,
			&oneOrder.certificate.certificateAccountDb.accountKeyDb.pem,
			&oneOrder.certificate.certificateAccountDb.accountKeyDb.apiKey,
			&oneOrder.certificate.certificateAccountDb.accountKeyDb.apiKeyNew,
			&oneOrder.certificate.certificateAccountDb.accountKeyDb.apiKeyDisabled,
			&oneOrder.certificate.certificateAccountDb.accountKeyDb.apiKeyViaUrl,
			&oneOrder.certificate.certificateAccountDb.accountKeyDb.createdAt,
			&oneOrder.certificate.certificateAccountDb.accountKeyDb.updatedAt,

			&oneOrder.finalizedKey.id,
			&oneOrder.finalizedKey.name,
			&oneOrder.finalizedKey.description,
			&oneOrder.finalizedKey.algorithmValue,
			&oneOrder.finalizedKey.pem,
			&oneOrder.finalizedKey.apiKey,
			&oneOrder.finalizedKey.apiKeyNew,
			&oneOrder.finalizedKey.apiKeyDisabled,
			&oneOrder.finalizedKey.apiKeyViaUrl,
			&oneOrder.finalizedKey.createdAt,
			&oneOrder.finalizedKey.updatedAt,

			&totalRows,
		)
		if err != nil {
			return nil, 0, err
		}

		// convert and append
		oneOrderConvert, err := oneOrder.toOrder()
		if err != nil {
			return nil, 0, err
		}

		orders = append(orders, oneOrderConvert)
	}

	return orders, totalRows, nil
}

// GetOrdersByCert fetches all of the orders for a specified certificate ID
func (store *Storage) GetOrdersByCert(certId int, q pagination_sort.Query) (orders []orders.Order, totalRowCount int, err error) {
	// validate and set sort
	sortField := q.SortField()

	switch sortField {
	case "id":
		sortField = "ao.id"
	case "created_at":
		sortField = "ao.created_at"
	case "valid_to":
		sortField = "ao.valid_to"
	case "status":
		sortField = "ao.status"
	case "keyname":
		sortField = "fk.name"
	default:
		sortField = "ao.created_at"
	}

	sort := sortField + " " + q.SortDirection()

	// do query
	ctx, cancel := context.WithTimeout(context.Background(), store.timeout)
	defer cancel()

	// WARNING: SQL Injection is possible if the variables are not properly
	// validated prior to this query being assembled!
	query := fmt.Sprintf(`
	SELECT
		/* order */
		ao.id, ao.acme_location, ao.status, ao.known_revoked, ao.error, ao.expires, ao.dns_identifiers, 
		ao.authorizations, ao.finalize, ao.certificate_url, ao.pem, ao.valid_from, ao.valid_to, ao.created_at,
		ao.updated_at, 

		/* order's cert */
		c.id, c.name, c.description, c.subject, c.subject_alts,
		c.csr_org, c.csr_ou, c.csr_country, c.csr_state, c.csr_city, c.csr_extra_extensions, c.created_at, c.updated_at,
		c.api_key, c.api_key_new, c.api_key_via_url, c.post_processing_command, c.post_processing_environment,
		c.post_processing_client_key,
		
		/* cert's key */
		ck.id, ck.name, ck.description, ck.algorithm, ck.pem, ck.api_key, ck.api_key_new, ck.api_key_disabled,
		ck.api_key_via_url,	ck.created_at, ck.updated_at,

		/* cert's account */
		ca.id, ca.name, ca.description, ca.status, ca.email, ca.accepted_tos,
		ca.created_at, ca.updated_at, ca.kid,

		/* cert's account's server */
		aserv.id, aserv.name, aserv.description, aserv.directory_url, aserv.is_staging, aserv.created_at,
		aserv.updated_at,

		/* cert's account's key */
		ak.id, ak.name, ak.description, ak.algorithm, ak.pem, ak.api_key, ak.api_key_new, ak.api_key_disabled,
		ak.api_key_via_url,	ak.created_at, ak.updated_at,

		/* finalized key */
		COALESCE(fk.id, -2), COALESCE(fk.name, 'null'), COALESCE(fk.description, 'null'), 
		COALESCE(fk.algorithm, 'null'), COALESCE(fk.pem, 'null'), COALESCE(fk.api_key, 'null'),
		COALESCE(fk.api_key_new, 'null'), COALESCE(fk.api_key_disabled, false),
		COALESCE(fk.api_key_via_url, false), COALESCE(fk.created_at, -2), COALESCE(fk.updated_at, -2),

		count(*) OVER() AS full_count
	FROM
		acme_orders ao
		LEFT JOIN certificates c on (ao.certificate_id = c.id)
		LEFT JOIN private_keys ck on (c.private_key_id = ck.id)
		LEFT JOIN acme_accounts ca on (c.acme_account_id = ca.id)
		LEFT JOIN acme_servers aserv on (ca.acme_server_id = aserv.id)
		LEFT JOIN private_keys ak on (ca.private_key_id = ak.id)
		LEFT JOIN private_keys fk on (ao.finalized_key_id = fk.id)
	WHERE
		ao.certificate_id = $1
	ORDER BY
		%s
	LIMIT
		$2
	OFFSET
		$3
	`, sort)

	rows, err := store.db.QueryContext(ctx, query,
		certId,
		q.Limit(),
		q.Offset(),
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	// for total row count
	var totalRows int

	for rows.Next() {
		var oneOrder orderDb

		err = rows.Scan(
			&oneOrder.id,
			&oneOrder.location,
			&oneOrder.status,
			&oneOrder.knownRevoked,
			&oneOrder.err,
			&oneOrder.expires,
			&oneOrder.dnsIdentifiers,
			&oneOrder.authorizations,
			&oneOrder.finalize,
			&oneOrder.certificateUrl,
			&oneOrder.pem,
			&oneOrder.validFrom,
			&oneOrder.validTo,
			&oneOrder.createdAt,
			&oneOrder.updatedAt,

			&oneOrder.certificate.id,
			&oneOrder.certificate.name,
			&oneOrder.certificate.description,
			&oneOrder.certificate.subject,
			&oneOrder.certificate.subjectAltNames,
			&oneOrder.certificate.organization,
			&oneOrder.certificate.organizationalUnit,
			&oneOrder.certificate.country,
			&oneOrder.certificate.state,
			&oneOrder.certificate.city,
			&oneOrder.certificate.csrExtraExtensions,
			&oneOrder.certificate.createdAt,
			&oneOrder.certificate.updatedAt,
			&oneOrder.certificate.apiKey,
			&oneOrder.certificate.apiKeyNew,
			&oneOrder.certificate.apiKeyViaUrl,
			&oneOrder.certificate.postProcessingCommand,
			&oneOrder.certificate.postProcessingEnvironment,
			&oneOrder.certificate.postProcessingClientKeyB64,

			&oneOrder.certificate.certificateKeyDb.id,
			&oneOrder.certificate.certificateKeyDb.name,
			&oneOrder.certificate.certificateKeyDb.description,
			&oneOrder.certificate.certificateKeyDb.algorithmValue,
			&oneOrder.certificate.certificateKeyDb.pem,
			&oneOrder.certificate.certificateKeyDb.apiKey,
			&oneOrder.certificate.certificateKeyDb.apiKeyNew,
			&oneOrder.certificate.certificateKeyDb.apiKeyDisabled,
			&oneOrder.certificate.certificateKeyDb.apiKeyViaUrl,
			&oneOrder.certificate.certificateKeyDb.createdAt,
			&oneOrder.certificate.certificateKeyDb.updatedAt,

			&oneOrder.certificate.certificateAccountDb.id,
			&oneOrder.certificate.certificateAccountDb.name,
			&oneOrder.certificate.certificateAccountDb.description,
			&oneOrder.certificate.certificateAccountDb.status,
			&oneOrder.certificate.certificateAccountDb.email,
			&oneOrder.certificate.certificateAccountDb.acceptedTos,
			&oneOrder.certificate.certificateAccountDb.createdAt,
			&oneOrder.certificate.certificateAccountDb.updatedAt,
			&oneOrder.certificate.certificateAccountDb.kid,

			&oneOrder.certificate.certificateAccountDb.accountServerDb.id,
			&oneOrder.certificate.certificateAccountDb.accountServerDb.name,
			&oneOrder.certificate.certificateAccountDb.accountServerDb.description,
			&oneOrder.certificate.certificateAccountDb.accountServerDb.directoryUrl,
			&oneOrder.certificate.certificateAccountDb.accountServerDb.isStaging,
			&oneOrder.certificate.certificateAccountDb.accountServerDb.createdAt,
			&oneOrder.certificate.certificateAccountDb.accountServerDb.updatedAt,

			&oneOrder.certificate.certificateAccountDb.accountKeyDb.id,
			&oneOrder.certificate.certificateAccountDb.accountKeyDb.name,
			&oneOrder.certificate.certificateAccountDb.accountKeyDb.description,
			&oneOrder.certificate.certificateAccountDb.accountKeyDb.algorithmValue,
			&oneOrder.certificate.certificateAccountDb.accountKeyDb.pem,
			&oneOrder.certificate.certificateAccountDb.accountKeyDb.apiKey,
			&oneOrder.certificate.certificateAccountDb.accountKeyDb.apiKeyNew,
			&oneOrder.certificate.certificateAccountDb.accountKeyDb.apiKeyDisabled,
			&oneOrder.certificate.certificateAccountDb.accountKeyDb.apiKeyViaUrl,
			&oneOrder.certificate.certificateAccountDb.accountKeyDb.createdAt,
			&oneOrder.certificate.certificateAccountDb.accountKeyDb.updatedAt,

			&oneOrder.finalizedKey.id,
			&oneOrder.finalizedKey.name,
			&oneOrder.finalizedKey.description,
			&oneOrder.finalizedKey.algorithmValue,
			&oneOrder.finalizedKey.pem,
			&oneOrder.finalizedKey.apiKey,
			&oneOrder.finalizedKey.apiKeyNew,
			&oneOrder.finalizedKey.apiKeyDisabled,
			&oneOrder.finalizedKey.apiKeyViaUrl,
			&oneOrder.finalizedKey.createdAt,
			&oneOrder.finalizedKey.updatedAt,

			&totalRows,
		)
		if err != nil {
			return nil, 0, err
		}

		// convert and append
		oneOrderConvert, err := oneOrder.toOrder()
		if err != nil {
			return nil, 0, err
		}

		orders = append(orders, oneOrderConvert)
	}

	return orders, totalRows, nil
}

// GetAllIncompleteOrderIds returns an array of all of the incomplete orders in storage.
func (store *Storage) GetAllIncompleteOrderIds() (orderIds []int, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), store.timeout)
	defer cancel()

	query := `
	SELECT
		id
	FROM
		acme_orders
	WHERE
		status = "pending"
		OR
		status = "ready"
		OR
		status = "processing"
	`

	// qeuery db
	rows, err := store.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// read result
	for rows.Next() {
		var orderId int

		err = rows.Scan(&orderId)
		if err != nil {
			return nil, err
		}

		orderIds = append(orderIds, orderId)
	}

	return orderIds, nil
}

// GetExpiringCertIds returns a slice of certificate ids for certificates that are valid for less
// than the specified maxTimeRemaining. If a cert does not have a valid order, it is excluded.
func (store *Storage) GetExpiringCertIds(maxTimeRemaining time.Duration) (certIds []int, err error) {
	// query
	ctx, cancel := context.WithTimeout(context.Background(), store.timeout)
	defer cancel()

	query := `
		SELECT
			ao.certificate_id
		FROM
			acme_orders ao
		WHERE 
			ao.status = "valid"
			AND
			ao.known_revoked = 0
			AND
			ao.valid_to > $1
			AND
			ao.pem NOT NULL
			AND
			ao.certificate_id IS NOT NULL
		GROUP BY
			ao.certificate_id
		HAVING
			MAX(ao.valid_to)
			AND
			ao.valid_to < $2
		`

	// calculate the max expiration (unix) for the query
	maxExpirationUnix := time.Now().Add(maxTimeRemaining).Unix()

	// get records
	rows, err := store.db.QueryContext(ctx, query,
		time.Now().Unix(),
		maxExpirationUnix,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var certId int

		err = rows.Scan(&certId)
		if err != nil {
			return nil, err
		}

		certIds = append(certIds, certId)
	}

	return certIds, nil
}

// GetNewestIncompleteCertOrderId returns the most recent incomplete order for a specified certId,
// assuming there is one.
func (store *Storage) GetNewestIncompleteCertOrderId(certId int) (orderId int, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), store.timeout)
	defer cancel()

	query := `
	SELECT
		id
	FROM
		acme_orders
	WHERE
		certificate_id = $1
		AND
		(
			status = "pending"
			OR
			status = "ready"
			OR
			status = "processing"
		)
		AND
		expires > $2
	GROUP BY
		certificate_id
	HAVING
		MAX(expires)
	`

	row := store.db.QueryRowContext(ctx, query,
		certId,
		timeNow(),
	)

	err = row.Scan(
		&orderId,
	)
	if err != nil {
		return -2, err
	}

	return orderId, nil
}

// GetOrders fetches the Order for each ID in the orderIDs slice and returns the
// slice of Order
func (store *Storage) GetOrders(orderIDs []int) (orders []orders.Order, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), store.timeout)
	defer cancel()

	// turn IDs into comma sep string
	var stringIDs []string
	for _, i := range orderIDs {
		stringIDs = append(stringIDs, strconv.Itoa(i))
	}
	IDsString := strings.Join(stringIDs, ", ")

	// WARNING: SQL Injection is possible if the variables are not properly
	// validated prior to this query being assembled!
	query := fmt.Sprintf(`
	SELECT
		/* order */
		ao.id, ao.acme_location, ao.status, ao.known_revoked, ao.error, ao.expires, ao.dns_identifiers, 
		ao.authorizations, ao.finalize, ao.certificate_url, ao.pem, ao.valid_from, ao.valid_to, ao.created_at,
		ao.updated_at, 

		/* order's cert */
		c.id, c.name, c.description, c.subject, c.subject_alts,
		c.csr_org, c.csr_ou, c.csr_country, c.csr_state, c.csr_city, c.csr_extra_extensions, c.created_at, c.updated_at,
		c.api_key, c.api_key_new, c.api_key_via_url, c.post_processing_command, c.post_processing_environment,
		c.post_processing_client_key,
		
		/* cert's key */
		ck.id, ck.name, ck.description, ck.algorithm, ck.pem, ck.api_key, ak.api_key_new, ck.api_key_disabled,
		ck.api_key_via_url,	ck.created_at, ck.updated_at,

		/* cert's account */
		ca.id, ca.name, ca.description, ca.status, ca.email, ca.accepted_tos,
		ca.created_at, ca.updated_at, ca.kid,

		/* cert's account's server */
		aserv.id, aserv.name, aserv.description, aserv.directory_url, aserv.is_staging, aserv.created_at,
		aserv.updated_at,

		/* cert's account's key */
		ak.id, ak.name, ak.description, ak.algorithm, ak.pem, ak.api_key, ak.api_key_new, ak.api_key_disabled,
		ak.api_key_via_url,	ak.created_at, ak.updated_at,

		/* finalized key */
		COALESCE(fk.id, -2), COALESCE(fk.name, 'null'), COALESCE(fk.description, 'null'), 
		COALESCE(fk.algorithm, 'null'), COALESCE(fk.pem, 'null'), COALESCE(fk.api_key, 'null'),
		COALESCE(fk.api_key_new, 'null'), COALESCE(fk.api_key_disabled, false), COALESCE(fk.api_key_via_url, false),
		COALESCE(fk.created_at, -2), COALESCE(fk.updated_at, -2)
	FROM
		acme_orders ao
		LEFT JOIN certificates c on (ao.certificate_id = c.id)
		LEFT JOIN private_keys ck on (c.private_key_id = ck.id)
		LEFT JOIN acme_accounts ca on (c.acme_account_id = ca.id)
		LEFT JOIN acme_servers aserv on (ca.acme_server_id = aserv.id)
		LEFT JOIN private_keys ak on (ca.private_key_id = ak.id)
		LEFT JOIN private_keys fk on (ao.finalized_key_id = fk.id)
	WHERE
		ao.id IN (%s)
	`, IDsString)

	// query records
	rows, err := store.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var oneOrder orderDb

		err = rows.Scan(
			&oneOrder.id,
			&oneOrder.location,
			&oneOrder.status,
			&oneOrder.knownRevoked,
			&oneOrder.err,
			&oneOrder.expires,
			&oneOrder.dnsIdentifiers,
			&oneOrder.authorizations,
			&oneOrder.finalize,
			&oneOrder.certificateUrl,
			&oneOrder.pem,
			&oneOrder.validFrom,
			&oneOrder.validTo,
			&oneOrder.createdAt,
			&oneOrder.updatedAt,

			&oneOrder.certificate.id,
			&oneOrder.certificate.name,
			&oneOrder.certificate.description,
			&oneOrder.certificate.subject,
			&oneOrder.certificate.subjectAltNames,
			&oneOrder.certificate.organization,
			&oneOrder.certificate.organizationalUnit,
			&oneOrder.certificate.country,
			&oneOrder.certificate.state,
			&oneOrder.certificate.city,
			&oneOrder.certificate.csrExtraExtensions,
			&oneOrder.certificate.createdAt,
			&oneOrder.certificate.updatedAt,
			&oneOrder.certificate.apiKey,
			&oneOrder.certificate.apiKeyNew,
			&oneOrder.certificate.apiKeyViaUrl,
			&oneOrder.certificate.postProcessingCommand,
			&oneOrder.certificate.postProcessingEnvironment,
			&oneOrder.certificate.postProcessingClientKeyB64,

			&oneOrder.certificate.certificateKeyDb.id,
			&oneOrder.certificate.certificateKeyDb.name,
			&oneOrder.certificate.certificateKeyDb.description,
			&oneOrder.certificate.certificateKeyDb.algorithmValue,
			&oneOrder.certificate.certificateKeyDb.pem,
			&oneOrder.certificate.certificateKeyDb.apiKey,
			&oneOrder.certificate.certificateKeyDb.apiKeyNew,
			&oneOrder.certificate.certificateKeyDb.apiKeyDisabled,
			&oneOrder.certificate.certificateKeyDb.apiKeyViaUrl,
			&oneOrder.certificate.certificateKeyDb.createdAt,
			&oneOrder.certificate.certificateKeyDb.updatedAt,

			&oneOrder.certificate.certificateAccountDb.id,
			&oneOrder.certificate.certificateAccountDb.name,
			&oneOrder.certificate.certificateAccountDb.description,
			&oneOrder.certificate.certificateAccountDb.status,
			&oneOrder.certificate.certificateAccountDb.email,
			&oneOrder.certificate.certificateAccountDb.acceptedTos,
			&oneOrder.certificate.certificateAccountDb.createdAt,
			&oneOrder.certificate.certificateAccountDb.updatedAt,
			&oneOrder.certificate.certificateAccountDb.kid,

			&oneOrder.certificate.certificateAccountDb.accountServerDb.id,
			&oneOrder.certificate.certificateAccountDb.accountServerDb.name,
			&oneOrder.certificate.certificateAccountDb.accountServerDb.description,
			&oneOrder.certificate.certificateAccountDb.accountServerDb.directoryUrl,
			&oneOrder.certificate.certificateAccountDb.accountServerDb.isStaging,
			&oneOrder.certificate.certificateAccountDb.accountServerDb.createdAt,
			&oneOrder.certificate.certificateAccountDb.accountServerDb.updatedAt,

			&oneOrder.certificate.certificateAccountDb.accountKeyDb.id,
			&oneOrder.certificate.certificateAccountDb.accountKeyDb.name,
			&oneOrder.certificate.certificateAccountDb.accountKeyDb.description,
			&oneOrder.certificate.certificateAccountDb.accountKeyDb.algorithmValue,
			&oneOrder.certificate.certificateAccountDb.accountKeyDb.pem,
			&oneOrder.certificate.certificateAccountDb.accountKeyDb.apiKey,
			&oneOrder.certificate.certificateAccountDb.accountKeyDb.apiKeyNew,
			&oneOrder.certificate.certificateAccountDb.accountKeyDb.apiKeyDisabled,
			&oneOrder.certificate.certificateAccountDb.accountKeyDb.apiKeyViaUrl,
			&oneOrder.certificate.certificateAccountDb.accountKeyDb.createdAt,
			&oneOrder.certificate.certificateAccountDb.accountKeyDb.updatedAt,

			&oneOrder.finalizedKey.id,
			&oneOrder.finalizedKey.name,
			&oneOrder.finalizedKey.description,
			&oneOrder.finalizedKey.algorithmValue,
			&oneOrder.finalizedKey.pem,
			&oneOrder.finalizedKey.apiKey,
			&oneOrder.finalizedKey.apiKeyNew,
			&oneOrder.finalizedKey.apiKeyDisabled,
			&oneOrder.finalizedKey.apiKeyViaUrl,
			&oneOrder.finalizedKey.createdAt,
			&oneOrder.finalizedKey.updatedAt,
		)
		if err != nil {
			// if no record exists
			if errors.Is(err, sql.ErrNoRows) {
				err = storage.ErrNoRecord
			}
			return nil, err
		}

		// convert and append
		oneOrderConvert, err := oneOrder.toOrder()
		if err != nil {
			return nil, err
		}
		orders = append(orders, oneOrderConvert)
	}

	return orders, nil
}

// GetOneOrder fetches a specific Order by ID
func (store *Storage) GetOneOrder(orderID int) (order orders.Order, err error) {
	result, err := store.GetOrders([]int{orderID})
	if err != nil {
		return orders.Order{}, err
	}

	if len(result) < 1 {
		return orders.Order{}, storage.ErrNoRecord
	}

	return result[0], nil
}

// GetCertNewestValidOrderById returns the most recent valid order for the specified
// cert id
func (store *Storage) GetCertNewestValidOrderById(id int) (order orders.Order, err error) {
	return store.getCertNewestValidOrder(id, "")
}

// GetCertNewestValidOrderByName returns the most recent valid order for the specified
// cert name
func (store *Storage) GetCertNewestValidOrderByName(name string) (order orders.Order, err error) {
	return store.getCertNewestValidOrder(-1, name)
}

// getCertNewestValidOrder fetches the newest valid order for the specified cert
func (store *Storage) getCertNewestValidOrder(certId int, certName string) (order orders.Order, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), store.timeout)
	defer cancel()

	query := `
	SELECT
		/* order */
		ao.id, ao.acme_location, ao.status, ao.known_revoked, ao.error, ao.expires, ao.dns_identifiers, 
		ao.authorizations, ao.finalize, ao.certificate_url, ao.pem, ao.valid_from, ao.valid_to, ao.created_at,
		ao.updated_at, 

		/* order's cert */
		c.id, c.name, c.description, c.subject, c.subject_alts,
		c.csr_org, c.csr_ou, c.csr_country, c.csr_state, c.csr_city, c.csr_extra_extensions, c.created_at, c.updated_at,
		c.api_key, c.api_key_new, c.api_key_via_url, c.post_processing_command, c.post_processing_environment,
		c.post_processing_client_key,
		
		/* cert's key */
		ck.id, ck.name, ck.description, ck.algorithm, ck.pem, ck.api_key, ak.api_key_new, ck.api_key_disabled,
		ck.api_key_via_url,	ck.created_at, ck.updated_at,

		/* cert's account */
		ca.id, ca.name, ca.description, ca.status, ca.email, ca.accepted_tos,
		ca.created_at, ca.updated_at, ca.kid,

		/* cert's account's server */
		aserv.id, aserv.name, aserv.description, aserv.directory_url, aserv.is_staging, aserv.created_at,
		aserv.updated_at,

		/* cert's account's key */
		ak.id, ak.name, ak.description, ak.algorithm, ak.pem, ak.api_key, ak.api_key_new, ak.api_key_disabled,
		ak.api_key_via_url,	ak.created_at, ak.updated_at,

		/* finalized key */
		COALESCE(fk.id, -2), COALESCE(fk.name, 'null'), COALESCE(fk.description, 'null'), 
		COALESCE(fk.algorithm, 'null'), COALESCE(fk.pem, 'null'), COALESCE(fk.api_key, 'null'),
		COALESCE(fk.api_key_new, 'null'), COALESCE(fk.api_key_disabled, false), COALESCE(fk.api_key_via_url, false),
		COALESCE(fk.created_at, -2), COALESCE(fk.updated_at, -2)
	FROM
		acme_orders ao
		LEFT JOIN certificates c on (ao.certificate_id = c.id)
		LEFT JOIN private_keys ck on (c.private_key_id = ck.id)
		LEFT JOIN acme_accounts ca on (c.acme_account_id = ca.id)
		LEFT JOIN acme_servers aserv on (ca.acme_server_id = aserv.id)
		LEFT JOIN private_keys ak on (ca.private_key_id = ak.id)
		LEFT JOIN private_keys fk on (ao.finalized_key_id = fk.id)

	WHERE
		ao.status = "valid"
		AND
		ao.known_revoked = 0
		AND
		ao.valid_to > $1
		AND
		ao.pem NOT NULL
		AND
		(
			ao.certificate_id = $2
			OR
			c.name = $3
		)
	GROUP BY
		certificate_id
	HAVING
		MAX(valid_to)
	`

	row := store.db.QueryRowContext(ctx, query,
		time.Now().Unix(),
		certId,
		certName,
	)

	var oneOrder orderDb

	err = row.Scan(
		&oneOrder.id,
		&oneOrder.location,
		&oneOrder.status,
		&oneOrder.knownRevoked,
		&oneOrder.err,
		&oneOrder.expires,
		&oneOrder.dnsIdentifiers,
		&oneOrder.authorizations,
		&oneOrder.finalize,
		&oneOrder.certificateUrl,
		&oneOrder.pem,
		&oneOrder.validFrom,
		&oneOrder.validTo,
		&oneOrder.createdAt,
		&oneOrder.updatedAt,

		&oneOrder.certificate.id,
		&oneOrder.certificate.name,
		&oneOrder.certificate.description,
		&oneOrder.certificate.subject,
		&oneOrder.certificate.subjectAltNames,
		&oneOrder.certificate.organization,
		&oneOrder.certificate.organizationalUnit,
		&oneOrder.certificate.country,
		&oneOrder.certificate.state,
		&oneOrder.certificate.city,
		&oneOrder.certificate.csrExtraExtensions,
		&oneOrder.certificate.createdAt,
		&oneOrder.certificate.updatedAt,
		&oneOrder.certificate.apiKey,
		&oneOrder.certificate.apiKeyNew,
		&oneOrder.certificate.apiKeyViaUrl,
		&oneOrder.certificate.postProcessingCommand,
		&oneOrder.certificate.postProcessingEnvironment,
		&oneOrder.certificate.postProcessingClientKeyB64,

		&oneOrder.certificate.certificateKeyDb.id,
		&oneOrder.certificate.certificateKeyDb.name,
		&oneOrder.certificate.certificateKeyDb.description,
		&oneOrder.certificate.certificateKeyDb.algorithmValue,
		&oneOrder.certificate.certificateKeyDb.pem,
		&oneOrder.certificate.certificateKeyDb.apiKey,
		&oneOrder.certificate.certificateKeyDb.apiKeyNew,
		&oneOrder.certificate.certificateKeyDb.apiKeyDisabled,
		&oneOrder.certificate.certificateKeyDb.apiKeyViaUrl,
		&oneOrder.certificate.certificateKeyDb.createdAt,
		&oneOrder.certificate.certificateKeyDb.updatedAt,

		&oneOrder.certificate.certificateAccountDb.id,
		&oneOrder.certificate.certificateAccountDb.name,
		&oneOrder.certificate.certificateAccountDb.description,
		&oneOrder.certificate.certificateAccountDb.status,
		&oneOrder.certificate.certificateAccountDb.email,
		&oneOrder.certificate.certificateAccountDb.acceptedTos,
		&oneOrder.certificate.certificateAccountDb.createdAt,
		&oneOrder.certificate.certificateAccountDb.updatedAt,
		&oneOrder.certificate.certificateAccountDb.kid,

		&oneOrder.certificate.certificateAccountDb.accountServerDb.id,
		&oneOrder.certificate.certificateAccountDb.accountServerDb.name,
		&oneOrder.certificate.certificateAccountDb.accountServerDb.description,
		&oneOrder.certificate.certificateAccountDb.accountServerDb.directoryUrl,
		&oneOrder.certificate.certificateAccountDb.accountServerDb.isStaging,
		&oneOrder.certificate.certificateAccountDb.accountServerDb.createdAt,
		&oneOrder.certificate.certificateAccountDb.accountServerDb.updatedAt,

		&oneOrder.certificate.certificateAccountDb.accountKeyDb.id,
		&oneOrder.certificate.certificateAccountDb.accountKeyDb.name,
		&oneOrder.certificate.certificateAccountDb.accountKeyDb.description,
		&oneOrder.certificate.certificateAccountDb.accountKeyDb.algorithmValue,
		&oneOrder.certificate.certificateAccountDb.accountKeyDb.pem,
		&oneOrder.certificate.certificateAccountDb.accountKeyDb.apiKey,
		&oneOrder.certificate.certificateAccountDb.accountKeyDb.apiKeyNew,
		&oneOrder.certificate.certificateAccountDb.accountKeyDb.apiKeyDisabled,
		&oneOrder.certificate.certificateAccountDb.accountKeyDb.apiKeyViaUrl,
		&oneOrder.certificate.certificateAccountDb.accountKeyDb.createdAt,
		&oneOrder.certificate.certificateAccountDb.accountKeyDb.updatedAt,

		&oneOrder.finalizedKey.id,
		&oneOrder.finalizedKey.name,
		&oneOrder.finalizedKey.description,
		&oneOrder.finalizedKey.algorithmValue,
		&oneOrder.finalizedKey.pem,
		&oneOrder.finalizedKey.apiKey,
		&oneOrder.finalizedKey.apiKeyNew,
		&oneOrder.finalizedKey.apiKeyDisabled,
		&oneOrder.finalizedKey.apiKeyViaUrl,
		&oneOrder.finalizedKey.createdAt,
		&oneOrder.finalizedKey.updatedAt,
	)
	if err != nil {
		// if no record exists
		if errors.Is(err, sql.ErrNoRows) {
			err = storage.ErrNoRecord
		}
		return orders.Order{}, err
	}

	order, err = oneOrder.toOrder()
	if err != nil {
		return orders.Order{}, err
	}

	return order, nil
}
