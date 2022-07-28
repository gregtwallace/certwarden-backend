package sqlite

import (
	"context"
	"legocerthub-backend/pkg/domain/orders"
	"legocerthub-backend/pkg/domain/private_keys"
)

// accountDbToAcc turns the database representation of a certificate into a Certificate
func (orderDb *orderDb) orderDbToOrder() (order orders.Order, err error) {
	// convert embedded private key db
	var finalKey = new(private_keys.Key)
	if orderDb.finalizedKey != nil && orderDb.finalizedKey.id.Valid {
		*finalKey, err = orderDb.finalizedKey.keyDbToKey()
		if err != nil {
			return orders.Order{}, err
		}
	} else {
		finalKey = nil
	}

	return orders.Order{
		ID: nullInt32ToInt(orderDb.id),
		// omit account and cert, not needed
		Location:       nullStringToString(orderDb.location),
		Status:         nullStringToString(orderDb.status),
		Error:          nullStringToAcmeError(orderDb.err),
		Expires:        nullInt32ToInt(orderDb.expires),
		DnsIdentifiers: commaNullStringToSlice(orderDb.dnsIdentifiers),
		Authorizations: commaNullStringToSlice(orderDb.authorizations),
		Finalize:       nullStringToString(orderDb.finalize),
		FinalizedKey:   finalKey,
		CertificateUrl: nullStringToString(orderDb.certificateUrl),
		CreatedAt:      nullInt32ToInt(orderDb.createdAt),
		UpdatedAt:      nullInt32ToInt(orderDb.updatedAt),
	}, nil
}

// GetCertOrders fetches all of the orders for a specified certificate ID
func (store *Storage) GetCertOrders(certId int) (orders []orders.Order, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), store.Timeout)
	defer cancel()

	query := `
	SELECT
		ao.id, ao.acme_location, ao.status, ao.error, ao.expires, ao.dns_identifiers, ao.authorizations, ao.finalize, 
		ao.certificate_url, ao.created_at, ao.updated_at,
		pk.id, pk.name
	FROM
		acme_orders ao
		LEFT JOIN private_keys pk on (ao.finalized_key_id = pk.id)
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
		// initialize keyDb pointer (or nil deref)
		oneOrder.finalizedKey = new(keyDb)
		err = rows.Scan(
			&oneOrder.id,
			&oneOrder.location,
			&oneOrder.status,
			&oneOrder.err,
			&oneOrder.expires,
			&oneOrder.dnsIdentifiers,
			&oneOrder.authorizations,
			&oneOrder.finalize,
			&oneOrder.certificateUrl,
			&oneOrder.createdAt,
			&oneOrder.updatedAt,
			&oneOrder.finalizedKey.id,
			&oneOrder.finalizedKey.name,
		)
		if err != nil {
			return nil, err
		}

		convertedOrder, err := oneOrder.orderDbToOrder()
		if err != nil {
			return nil, err
		}

		orders = append(orders, convertedOrder)
	}

	return orders, nil
}
