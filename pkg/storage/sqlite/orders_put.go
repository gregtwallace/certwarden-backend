package sqlite

import (
	"context"
	"legocerthub-backend/pkg/acme"
)

// acmeOrderToDb translates the acme order object to the db obj
func acmeOrderToDb(order acme.Order) orderDb {
	// create db obj
	var orderDb orderDb

	orderDb.status = stringToNullString(&order.Status)
	orderDb.err = acmeErrorToNullString(order.Error)
	orderDb.expires = intToNullInt32(order.Expires.ToUnixTime())
	orderDb.dnsIdentifiers = sliceToCommaNullString(order.Identifiers.DnsIdentifiers())
	orderDb.authorizations = sliceToCommaNullString(order.Authorizations)
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
