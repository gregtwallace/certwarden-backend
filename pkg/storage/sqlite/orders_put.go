package sqlite

import (
	"context"
	"legocerthub-backend/pkg/domain/orders"
)

// UpdateOrderAcme updates the specified order ID with acme.Order response
// data
func (store *Storage) PutOrderAcme(payload orders.UpdateAcmeOrderPayload) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), store.timeout)
	defer cancel()

	// update existing record
	query := `
		UPDATE
			acme_orders
		SET
			status = $1,
			expires = case when $2 is null then expires else $2 end,
			dns_identifiers = $3,
			error = case when $4 is null then error else $4 end,
			authorizations = $5,
			finalize = $6,
			certificate_url = case when $7 is null then certificate_url else $7 end,
			updated_at = $8
		WHERE
			id = $9
		`

	_, err = store.db.ExecContext(ctx, query,
		payload.Status,
		payload.Expires,
		makeJsonStringSlice(payload.DnsIds),
		payload.Error,
		makeJsonStringSlice(payload.Authorizations),
		payload.Finalize,
		payload.CertificateUrl,
		payload.UpdatedAt,
		payload.OrderId,
	)

	if err != nil {
		return err
	}

	// TODO: Handle 0 rows updated.

	return nil
}

// PutOrderInvalid updates the specified order ID to the status of 'invalid'.
func (store *Storage) PutOrderInvalid(orderId int) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), store.timeout)
	defer cancel()

	// update existing record
	query := `
		UPDATE
			acme_orders
		SET
			status = $1
		WHERE
			id = $2
		`

	_, err = store.db.ExecContext(ctx, query,
		"invalid",
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
	ctx, cancel := context.WithTimeout(context.Background(), store.timeout)
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

	_, err = store.db.ExecContext(ctx, query,
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
	ctx, cancel := context.WithTimeout(context.Background(), store.timeout)
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

	_, err = store.db.ExecContext(ctx, query,
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
	ctx, cancel := context.WithTimeout(context.Background(), store.timeout)
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

	_, err = store.db.ExecContext(ctx, query,
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
