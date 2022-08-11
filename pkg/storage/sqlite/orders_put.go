package sqlite

import (
	"context"
	"legocerthub-backend/pkg/acme"
	"legocerthub-backend/pkg/domain/orders"
)

// acmeOrderToDb translates the acme order object to the db obj
func acmeOrderToDb(order acme.Order) orderDb {
	// create db obj
	var orderDb orderDb

	// dnsIds
	dnsIds := order.Identifiers.DnsIdentifiers()

	orderDb.status = stringToNullString(&order.Status)
	orderDb.err = acmeErrorToNullString(order.Error)
	orderDb.expires = intToNullInt32(order.Expires.ToUnixTime())
	orderDb.dnsIdentifiers = sliceToCommaNullString(&dnsIds)
	orderDb.authorizations = sliceToCommaNullString(&order.Authorizations)
	orderDb.finalize = stringToNullString(&order.Finalize)
	orderDb.certificateUrl = stringToNullString(&order.Certificate)

	orderDb.updatedAt = timeNow()

	return orderDb
}

// UpdateOrderAcme updates the specified order ID with acme.Order response
// data
func (store *Storage) UpdateOrderAcme(orderId int, order acme.Order) (err error) {
	// Load order into db obj
	orderDb := acmeOrderToDb(order)

	ctx, cancel := context.WithTimeout(context.Background(), store.Timeout)
	defer cancel()

	// update existing record
	query := `
		UPDATE
			acme_orders
		SET
			status = case when $1 is null then status else $1 end,
			error = case when $2 is null then error else $2 end,
			expires = case when $3 is null then expires else $3 end,
			dns_identifiers = case when $4 is null then dns_identifiers else $4 end,
			authorizations = case when $5 is null then authorizations else $5 end,
			finalize = case when $6 is null then finalize else $6 end,
			certificate_url = case when $7 is null then certificate_url else $7 end,
			updated_at = $8
		WHERE
			id = $9
		`

	_, err = store.Db.ExecContext(ctx, query,
		orderDb.status,
		orderDb.err,
		orderDb.expires,
		orderDb.dnsIdentifiers,
		orderDb.authorizations,
		orderDb.finalize,
		orderDb.certificateUrl,
		orderDb.updatedAt,
		orderId,
	)

	if err != nil {
		return err
	}

	// TODO: Handle 0 rows updated.

	return nil
}

// UpdateFinalizedKey updates the specified order ID with key id
func (store *Storage) UpdateFinalizedKey(orderId int, keyId int) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), store.Timeout)
	defer cancel()

	// no checks or validation (shouldn't be needed)

	// update existing record
	query := `
		UPDATE
			acme_orders
		SET
			finalized_key_id = $1,
			updated_at = $2
		WHERE
			id = $3
		`

	_, err = store.Db.ExecContext(ctx, query,
		keyId,
		timeNow(),
		orderId,
	)

	if err != nil {
		return err
	}

	// TODO: Handle 0 rows updated.

	return nil
}

// UpdateOrderCert updates the specified order ID with the specified certificate data
func (store *Storage) UpdateOrderCert(orderId int, payload orders.CertPayload) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), store.Timeout)
	defer cancel()

	// no checks or validation (shouldn't be needed)

	// update existing record
	query := `
		UPDATE
			acme_orders
		SET
			pem = $1,
			valid_from = $2,
			valid_to = $3,
			updated_at = $4
		WHERE
			id = $5
		`

	_, err = store.Db.ExecContext(ctx, query,
		payload.Pem,
		payload.ValidFrom,
		payload.ValidTo,
		timeNow(),
		orderId,
	)

	if err != nil {
		return err
	}

	// TODO: Handle 0 rows updated.

	return nil
}

// RevokeOrder updates the revoked flag in db to true (1)
func (store *Storage) RevokeOrder(orderId int) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), store.Timeout)
	defer cancel()

	// no checks or validation (shouldn't be needed)

	// update existing record
	query := `
		UPDATE
			acme_orders
		SET
			known_revoked = $1,
			updated_at = $2
		WHERE
			id = $3
		`

	_, err = store.Db.ExecContext(ctx, query,
		1, // true
		timeNow(),
		orderId,
	)

	if err != nil {
		return err
	}

	// TODO: Handle 0 rows updated.

	return nil
}
