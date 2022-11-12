package sqlite

import (
	"context"
	"legocerthub-backend/pkg/domain/orders"
	"time"
)

// GetAllValidCurrentOrders fetches each cert's most recent valid order (essentially this
// is a list of the certificates that are currently being hosted via API key)
func (store *Storage) GetAllValidCurrentOrders() (orders []orders.Order, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), store.Timeout)
	defer cancel()

	query := `
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
		COALESCE(fk.api_key_via_url, false), COALESCE(fk.created_at, -2), COALESCE(fk.updated_at, -2)
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
		certificate_id
	HAVING
		MAX(valid_to)
	ORDER BY
		expires DESC
	`

	rows, err := store.Db.QueryContext(ctx, query,
		time.Now().Unix())
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
			return nil, err
		}

		convertedOrder := oneOrder.toOrder()
		orders = append(orders, convertedOrder)
	}

	return orders, nil
}

// GetOrdersByCert fetches all of the orders for a specified certificate ID
func (store *Storage) GetOrdersByCert(certId int) (orders []orders.Order, err error) {
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
		ao.certificate_id = $1
	ORDER BY
		expires DESC
	`

	rows, err := store.Db.QueryContext(ctx, query, certId)
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
			return nil, err
		}

		convertedOrder := oneOrder.toOrder()

		orders = append(orders, convertedOrder)
	}

	return orders, nil
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

	order = oneOrder.toOrder()

	return order, nil
}
