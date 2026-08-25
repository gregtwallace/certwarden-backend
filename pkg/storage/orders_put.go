package storage

import (
	"certwarden-backend/pkg/domain/orders"
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// UpdateOrderAcme updates the specified order ID with acme.Order response data
func (store *Storage) PutOrderACME(payload *orders.UpdateAcmeOrderPayload) error {
	ctx, cancel := context.WithTimeout(store.shutdownContext, store.timeout)
	defer cancel()

	// handle null expires
	var expiresVal *int64
	if payload.Expires != nil {
		expiresVal = new(payload.Expires.Unix())
	}

	// deal with error obj
	var acmeErr *string
	if payload.Error != nil {
		ae, err := json.Marshal(payload.Error)
		if err != nil {
			return err
		}
		acmeErr = new(string(ae))
	}

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

	res, err := store.db.ExecContext(ctx, query,
		payload.Status,
		expiresVal,
		makeJsonStringSlice(payload.DnsIds, true),
		acmeErr,
		makeJsonStringSlice(payload.Authorizations, true),
		payload.Finalize,
		payload.Profile,
		payload.CertificateUrl,
		payload.UpdatedAt.Unix(),
		payload.OrderID,
	)
	if err != nil {
		return err
	}

	// verify update actually happened
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected != 1 {
		return errorWrongUpdateRowCount(1, rowsAffected)
	}

	return nil
}

// PutOrderRenewalInfo updates the specified order ID with its renewal information object
func (store *Storage) PutOrderRenewalInfo(payload orders.UpdateRenewalInfoPayload) error {
	ctx, cancel := context.WithTimeout(store.shutdownContext, store.timeout)
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
		return fmt.Errorf("storage: failed to marshal renewal info (%w)", err)
	}

	res, err := store.db.ExecContext(ctx, query,
		string(ari),
		payload.UpdatedAt.Unix(),
		payload.OrderID,
	)
	if err != nil {
		return err
	}

	// verify update actually happened
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected != 1 {
		return errorWrongUpdateRowCount(1, rowsAffected)
	}

	return nil
}

// PutOrderStatusInvalid updates the specified order ID to the status of 'invalid'.
func (store *Storage) PutOrderStatusInvalid(orderId int, updatedAt time.Time) error {
	ctx, cancel := context.WithTimeout(store.shutdownContext, store.timeout)
	defer cancel()

	// update existing record
	query := `
		UPDATE
			acme_orders
		SET
			status = 'invalid',
			updated_at = $1
		WHERE
			id = $2
		`

	res, err := store.db.ExecContext(ctx, query,
		updatedAt.Unix(),
		orderId,
	)
	if err != nil {
		return err
	}

	// verify update actually happened
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected != 1 {
		return errorWrongUpdateRowCount(1, rowsAffected)
	}

	return nil
}

// UpdateFinalizedKey updates the specified order ID with key id
func (store *Storage) PutOrderFinalizedKey(orderId, keyId int, updatedAt time.Time) error {
	ctx, cancel := context.WithTimeout(store.shutdownContext, store.timeout)
	defer cancel()

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

	res, err := store.db.ExecContext(ctx, query,
		keyId,
		updatedAt.Unix(),
		orderId,
	)
	if err != nil {
		return err
	}

	// verify update actually happened
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected != 1 {
		return errorWrongUpdateRowCount(1, rowsAffected)
	}

	return nil
}

// PutOrderPemData updates the specified order ID with the specified certificate data and ari
// Todo: Refactor this to remove ARI
func (store *Storage) PutOrderPemData(orderId int, payload orders.CertPayload) error {
	ctx, cancel := context.WithTimeout(store.shutdownContext, store.timeout)
	defer cancel()

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

	// marshal ARI struct
	ari, err := json.Marshal(payload.RenewalInfo)
	if err != nil {
		return fmt.Errorf("storage: failed to marshal renewal info (%w)", err)
	}

	res, err := store.db.ExecContext(ctx, query,
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

	// verify update actually happened
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected != 1 {
		return errorWrongUpdateRowCount(1, rowsAffected)
	}

	return nil
}

// PutOrderRevoke updates the revoked flag in db to true (1)
func (store *Storage) PutOrderRevoke(orderId int, updatedAt time.Time) error {
	ctx, cancel := context.WithTimeout(store.shutdownContext, store.timeout)
	defer cancel()

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

	res, err := store.db.ExecContext(ctx, query,
		1, // true
		updatedAt.Unix(),
		orderId,
	)
	if err != nil {
		return err
	}

	// verify update actually happened
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected != 1 {
		return errorWrongUpdateRowCount(1, rowsAffected)
	}

	return nil
}
