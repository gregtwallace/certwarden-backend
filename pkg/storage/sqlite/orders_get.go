package sqlite

import (
	"context"
	"fmt"
	"legocerthub-backend/pkg/domain/orders"
	"legocerthub-backend/pkg/pagination_sort"
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
	ctx, cancel := context.WithTimeout(context.Background(), store.Timeout)
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
		c.id, c.name, c.description, c.subject, c.subject_alts, c.challenge_method, 
		c.csr_org, c.csr_ou, c.csr_country, c.csr_state, c.csr_city, c.created_at, c.updated_at,
		c.api_key, c.api_key_via_url,
		
		/* cert's key */
		ck.id, ck.name, ck.description, ck.algorithm, ck.pem, ck.api_key, ck.api_key_via_url,
		ck.created_at, ck.updated_at,

		/* cert's account */
		ca.id, ca.name, ca.description, ca.status, ca.email, ca.accepted_tos, ca.is_staging,
		ca.created_at, ca.updated_at, ca.kid,

		/* cert's account's key */
		ak.id, ak.name, ak.description, ak.algorithm, ak.pem, ak.api_key, ak.api_key_via_url,
		ak.created_at, ak.updated_at,

		/* finalized key */
		COALESCE(fk.id, -2), COALESCE(fk.name, 'null'), COALESCE(fk.description, 'null'), 
		COALESCE(fk.algorithm, 'null'), COALESCE(fk.pem, 'null'), COALESCE(fk.api_key, 'null'),
		COALESCE(fk.api_key_via_url, false), COALESCE(fk.created_at, -2), COALESCE(fk.updated_at, -2),
		
		count(*) OVER() AS full_count
	FROM
		acme_orders ao
		LEFT JOIN certificates c on (ao.certificate_id = c.id)
		LEFT JOIN private_keys ck on (c.private_key_id = ck.id)
		LEFT JOIN acme_accounts ca on (c.acme_account_id = ca.id)
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
	rows, err := store.Db.QueryContext(ctx, query,
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
			&oneOrder.certificate.challengeMethodValue,
			&oneOrder.certificate.organization,
			&oneOrder.certificate.organizationalUnit,
			&oneOrder.certificate.country,
			&oneOrder.certificate.state,
			&oneOrder.certificate.city,
			&oneOrder.certificate.createdAt,
			&oneOrder.certificate.updatedAt,
			&oneOrder.certificate.apiKey,
			&oneOrder.certificate.apiKeyViaUrl,

			&oneOrder.certificate.certificateKeyDb.id,
			&oneOrder.certificate.certificateKeyDb.name,
			&oneOrder.certificate.certificateKeyDb.description,
			&oneOrder.certificate.certificateKeyDb.algorithmValue,
			&oneOrder.certificate.certificateKeyDb.pem,
			&oneOrder.certificate.certificateKeyDb.apiKey,
			&oneOrder.certificate.certificateKeyDb.apiKeyViaUrl,
			&oneOrder.certificate.certificateKeyDb.createdAt,
			&oneOrder.certificate.certificateKeyDb.updatedAt,

			&oneOrder.certificate.certificateAccountDb.id,
			&oneOrder.certificate.certificateAccountDb.name,
			&oneOrder.certificate.certificateAccountDb.description,
			&oneOrder.certificate.certificateAccountDb.status,
			&oneOrder.certificate.certificateAccountDb.email,
			&oneOrder.certificate.certificateAccountDb.acceptedTos,
			&oneOrder.certificate.certificateAccountDb.isStaging,
			&oneOrder.certificate.certificateAccountDb.createdAt,
			&oneOrder.certificate.certificateAccountDb.updatedAt,
			&oneOrder.certificate.certificateAccountDb.kid,

			&oneOrder.certificate.certificateAccountDb.accountKeyDb.id,
			&oneOrder.certificate.certificateAccountDb.accountKeyDb.name,
			&oneOrder.certificate.certificateAccountDb.accountKeyDb.description,
			&oneOrder.certificate.certificateAccountDb.accountKeyDb.algorithmValue,
			&oneOrder.certificate.certificateAccountDb.accountKeyDb.pem,
			&oneOrder.certificate.certificateAccountDb.accountKeyDb.apiKey,
			&oneOrder.certificate.certificateAccountDb.accountKeyDb.apiKeyViaUrl,
			&oneOrder.certificate.certificateAccountDb.accountKeyDb.createdAt,
			&oneOrder.certificate.certificateAccountDb.accountKeyDb.updatedAt,

			&oneOrder.finalizedKey.id,
			&oneOrder.finalizedKey.name,
			&oneOrder.finalizedKey.description,
			&oneOrder.finalizedKey.algorithmValue,
			&oneOrder.finalizedKey.pem,
			&oneOrder.finalizedKey.apiKey,
			&oneOrder.finalizedKey.apiKeyViaUrl,
			&oneOrder.finalizedKey.createdAt,
			&oneOrder.finalizedKey.updatedAt,

			&totalRows,
		)
		if err != nil {
			return nil, 0, err
		}

		convertedOrder := oneOrder.toOrder(store)
		orders = append(orders, convertedOrder)
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
	ctx, cancel := context.WithTimeout(context.Background(), store.Timeout)
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
		c.id, c.name, c.description, c.subject, c.subject_alts, c.challenge_method, 
		c.csr_org, c.csr_ou, c.csr_country, c.csr_state, c.csr_city, c.created_at, c.updated_at,
		c.api_key, c.api_key_via_url,
		
		/* cert's key */
		ck.id, ck.name, ck.description, ck.algorithm, ck.pem, ck.api_key, ck.api_key_via_url,
		ck.created_at, ck.updated_at,

		/* cert's account */
		ca.id, ca.name, ca.description, ca.status, ca.email, ca.accepted_tos, ca.is_staging,
		ca.created_at, ca.updated_at, ca.kid,

		/* cert's account's key */
		ak.id, ak.name, ak.description, ak.algorithm, ak.pem, ak.api_key, ak.api_key_via_url,
		ak.created_at, ak.updated_at,

		/* finalized key */
		COALESCE(fk.id, -2), COALESCE(fk.name, 'null'), COALESCE(fk.description, 'null'), 
		COALESCE(fk.algorithm, 'null'), COALESCE(fk.pem, 'null'), COALESCE(fk.api_key, 'null'),
		COALESCE(fk.api_key_via_url, false), COALESCE(fk.created_at, -2), COALESCE(fk.updated_at, -2),

		count(*) OVER() AS full_count
	FROM
		acme_orders ao
		LEFT JOIN certificates c on (ao.certificate_id = c.id)
		LEFT JOIN private_keys ck on (c.private_key_id = ck.id)
		LEFT JOIN acme_accounts ca on (c.acme_account_id = ca.id)
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

	rows, err := store.Db.QueryContext(ctx, query,
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
			&oneOrder.certificate.challengeMethodValue,
			&oneOrder.certificate.organization,
			&oneOrder.certificate.organizationalUnit,
			&oneOrder.certificate.country,
			&oneOrder.certificate.state,
			&oneOrder.certificate.city,
			&oneOrder.certificate.createdAt,
			&oneOrder.certificate.updatedAt,
			&oneOrder.certificate.apiKey,
			&oneOrder.certificate.apiKeyViaUrl,

			&oneOrder.certificate.certificateKeyDb.id,
			&oneOrder.certificate.certificateKeyDb.name,
			&oneOrder.certificate.certificateKeyDb.description,
			&oneOrder.certificate.certificateKeyDb.algorithmValue,
			&oneOrder.certificate.certificateKeyDb.pem,
			&oneOrder.certificate.certificateKeyDb.apiKey,
			&oneOrder.certificate.certificateKeyDb.apiKeyViaUrl,
			&oneOrder.certificate.certificateKeyDb.createdAt,
			&oneOrder.certificate.certificateKeyDb.updatedAt,

			&oneOrder.certificate.certificateAccountDb.id,
			&oneOrder.certificate.certificateAccountDb.name,
			&oneOrder.certificate.certificateAccountDb.description,
			&oneOrder.certificate.certificateAccountDb.status,
			&oneOrder.certificate.certificateAccountDb.email,
			&oneOrder.certificate.certificateAccountDb.acceptedTos,
			&oneOrder.certificate.certificateAccountDb.isStaging,
			&oneOrder.certificate.certificateAccountDb.createdAt,
			&oneOrder.certificate.certificateAccountDb.updatedAt,
			&oneOrder.certificate.certificateAccountDb.kid,

			&oneOrder.certificate.certificateAccountDb.accountKeyDb.id,
			&oneOrder.certificate.certificateAccountDb.accountKeyDb.name,
			&oneOrder.certificate.certificateAccountDb.accountKeyDb.description,
			&oneOrder.certificate.certificateAccountDb.accountKeyDb.algorithmValue,
			&oneOrder.certificate.certificateAccountDb.accountKeyDb.pem,
			&oneOrder.certificate.certificateAccountDb.accountKeyDb.apiKey,
			&oneOrder.certificate.certificateAccountDb.accountKeyDb.apiKeyViaUrl,
			&oneOrder.certificate.certificateAccountDb.accountKeyDb.createdAt,
			&oneOrder.certificate.certificateAccountDb.accountKeyDb.updatedAt,

			&oneOrder.finalizedKey.id,
			&oneOrder.finalizedKey.name,
			&oneOrder.finalizedKey.description,
			&oneOrder.finalizedKey.algorithmValue,
			&oneOrder.finalizedKey.pem,
			&oneOrder.finalizedKey.apiKey,
			&oneOrder.finalizedKey.apiKeyViaUrl,
			&oneOrder.finalizedKey.createdAt,
			&oneOrder.finalizedKey.updatedAt,

			&totalRows,
		)
		if err != nil {
			return nil, 0, err
		}

		convertedOrder := oneOrder.toOrder(store)

		orders = append(orders, convertedOrder)
	}

	return orders, totalRows, nil
}

// GetAllIncompleteOrderIds returns an array of all of the incomplete orders in storage.
func (store *Storage) GetAllIncompleteOrderIds() (orderIds []int, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), store.Timeout)
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
	rows, err := store.Db.QueryContext(ctx, query)
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

	return certIds, nil
}

// GetNewestIncompleteCertOrderId returns the most recent incomplete order for a specified certId,
// assuming there is one.
func (store *Storage) GetNewestIncompleteCertOrderId(certId int) (orderId int, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), store.Timeout)
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

	row := store.Db.QueryRowContext(ctx, query,
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

// GetOneOrder fetches a specific Order by ID
func (store *Storage) GetOneOrder(orderId int) (order orders.Order, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), store.Timeout)
	defer cancel()

	query := `
	SELECT
		/* order */
		ao.id, ao.acme_location, ao.status, ao.known_revoked, ao.error, ao.expires, ao.dns_identifiers, 
		ao.authorizations, ao.finalize, ao.certificate_url, ao.pem, ao.valid_from, ao.valid_to, ao.created_at,
		ao.updated_at, 

		/* order's cert */
		c.id, c.name, c.description, c.subject, c.subject_alts, c.challenge_method, 
		c.csr_org, c.csr_ou, c.csr_country, c.csr_state, c.csr_city, c.created_at, c.updated_at,
		c.api_key, c.api_key_via_url,
		
		/* cert's key */
		ck.id, ck.name, ck.description, ck.algorithm, ck.pem, ck.api_key, ck.api_key_via_url,
		ck.created_at, ck.updated_at,

		/* cert's account */
		ca.id, ca.name, ca.description, ca.status, ca.email, ca.accepted_tos, ca.is_staging,
		ca.created_at, ca.updated_at, ca.kid,

		/* cert's account's key */
		ak.id, ak.name, ak.description, ak.algorithm, ak.pem, ak.api_key, ak.api_key_via_url,
		ak.created_at, ak.updated_at,

		/* finalized key */
		COALESCE(fk.id, -2), COALESCE(fk.name, 'null'), COALESCE(fk.description, 'null'), 
		COALESCE(fk.algorithm, 'null'), COALESCE(fk.pem, 'null'), COALESCE(fk.api_key, 'null'),
		COALESCE(fk.api_key_via_url, false), COALESCE(fk.created_at, -2), COALESCE(fk.updated_at, -2)
	FROM
		acme_orders ao
		LEFT JOIN certificates c on (ao.certificate_id = c.id)
		LEFT JOIN private_keys ck on (c.private_key_id = ck.id)
		LEFT JOIN acme_accounts ca on (c.acme_account_id = ca.id)
		LEFT JOIN private_keys ak on (ca.private_key_id = ak.id)
		LEFT JOIN private_keys fk on (ao.finalized_key_id = fk.id)
	WHERE
		ao.id = $1
	ORDER BY
		expires DESC
	`

	row := store.Db.QueryRowContext(ctx, query, orderId)

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
		&oneOrder.certificate.challengeMethodValue,
		&oneOrder.certificate.organization,
		&oneOrder.certificate.organizationalUnit,
		&oneOrder.certificate.country,
		&oneOrder.certificate.state,
		&oneOrder.certificate.city,
		&oneOrder.certificate.createdAt,
		&oneOrder.certificate.updatedAt,
		&oneOrder.certificate.apiKey,
		&oneOrder.certificate.apiKeyViaUrl,

		&oneOrder.certificate.certificateKeyDb.id,
		&oneOrder.certificate.certificateKeyDb.name,
		&oneOrder.certificate.certificateKeyDb.description,
		&oneOrder.certificate.certificateKeyDb.algorithmValue,
		&oneOrder.certificate.certificateKeyDb.pem,
		&oneOrder.certificate.certificateKeyDb.apiKey,
		&oneOrder.certificate.certificateKeyDb.apiKeyViaUrl,
		&oneOrder.certificate.certificateKeyDb.createdAt,
		&oneOrder.certificate.certificateKeyDb.updatedAt,

		&oneOrder.certificate.certificateAccountDb.id,
		&oneOrder.certificate.certificateAccountDb.name,
		&oneOrder.certificate.certificateAccountDb.description,
		&oneOrder.certificate.certificateAccountDb.status,
		&oneOrder.certificate.certificateAccountDb.email,
		&oneOrder.certificate.certificateAccountDb.acceptedTos,
		&oneOrder.certificate.certificateAccountDb.isStaging,
		&oneOrder.certificate.certificateAccountDb.createdAt,
		&oneOrder.certificate.certificateAccountDb.updatedAt,
		&oneOrder.certificate.certificateAccountDb.kid,

		&oneOrder.certificate.certificateAccountDb.accountKeyDb.id,
		&oneOrder.certificate.certificateAccountDb.accountKeyDb.name,
		&oneOrder.certificate.certificateAccountDb.accountKeyDb.description,
		&oneOrder.certificate.certificateAccountDb.accountKeyDb.algorithmValue,
		&oneOrder.certificate.certificateAccountDb.accountKeyDb.pem,
		&oneOrder.certificate.certificateAccountDb.accountKeyDb.apiKey,
		&oneOrder.certificate.certificateAccountDb.accountKeyDb.apiKeyViaUrl,
		&oneOrder.certificate.certificateAccountDb.accountKeyDb.createdAt,
		&oneOrder.certificate.certificateAccountDb.accountKeyDb.updatedAt,

		&oneOrder.finalizedKey.id,
		&oneOrder.finalizedKey.name,
		&oneOrder.finalizedKey.description,
		&oneOrder.finalizedKey.algorithmValue,
		&oneOrder.finalizedKey.pem,
		&oneOrder.finalizedKey.apiKey,
		&oneOrder.finalizedKey.apiKeyViaUrl,
		&oneOrder.finalizedKey.createdAt,
		&oneOrder.finalizedKey.updatedAt,
	)
	if err != nil {
		return orders.Order{}, err
	}

	order = oneOrder.toOrder(store)

	return order, nil
}
