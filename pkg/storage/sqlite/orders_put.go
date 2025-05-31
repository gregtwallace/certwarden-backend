package sqlite

import (
	"certwarden-backend/pkg/domain/orders"
	"context"
	"encoding/json"
	"fmt"
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
			expires = $2,
			dns_identifiers = $3,
			error = $4,
			authorizations = $5,
			finalize = $6,
			profile = $7,
			certificate_url = $8,
			updated_at = $9
		WHERE
			id = $10
		`

	_, err = store.db.ExecContext(ctx, query,
		payload.Status,
		payload.Expires,
		makeJsonStringSlice(payload.DnsIds),
		payload.Error,
		makeJsonStringSlice(payload.Authorizations),
		payload.Finalize,
		payload.Profile,
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

// PutRenewalInfo updates the specified order ID with its renewal information object
func (store *Storage) PutRenewalInfo(payload orders.UpdateRenewalInfoPayload) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), store.timeout)
	defer cancel()

	// update existing record
	query := `
		UPDATE
			acme_orders
		SET
			renewal_info = $1,
			updated_at = $2
		WHERE
			id = $3
		`

	// marshal struct
	ari, err := json.Marshal(payload.RenewalInfo)
	if err != nil {
		return fmt.Errorf("storage: failed to marshal renewal info (%s)", err)
	}

	_, err = store.db.ExecContext(ctx, query,
		string(ari),
		payload.UpdatedAt,
		payload.OrderID,
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

// UpdateOrderCert updates the specified order ID with the specified certificate data and ari
func (store *Storage) UpdateOrderCert(orderId int, payload *orders.CertPayload) (err error) {
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
			chain_root_cn = $4,
			renewal_info = $5,
			updated_at = $6
		WHERE
			id = $7
		`

	// marshal struct
	ari, err := json.Marshal(payload.RenewalInfo)
	if err != nil {
		return fmt.Errorf("storage: failed to marshal renewal info (%s)", err)
	}

	_, err = store.db.ExecContext(ctx, query,
		payload.AcmeCert.PEM(),
		payload.AcmeCert.NotBefore().Unix(),
		payload.AcmeCert.NotAfter().Unix(),
		payload.AcmeCert.ChainRootCN(),
		string(ari),
		payload.UpdatedAt.Unix(),
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
