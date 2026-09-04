package storage

import (
	"certwarden-backend/pkg/domain/orders"
	"context"
	"database/sql"
	"errors"
)

// PostNewOrder makes a new order in the db. An error is returned if the order
// location already exists (or any other error)
// TODO: Refactor to make this behave like the other 'post' methods. (e.g., return
// the new order if success, otherwise just return error). Switch 'put' method to
// take the order URL (which is unique) and then all this func will need to do is
// return already existing error.
// TODO: Should not rely on int value when non-nil error is returned.
func (store *Storage) PostNewOrder(payload *orders.NewOrderAcmePayload) (newId int, err error) {
	ctx, cancel := context.WithTimeout(store.shutdownContext, store.timeout)
	defer cancel()

	// transaction
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return -2, err
	}
	defer tx.Rollback()

	// check if the order already exists
	query := `
	SELECT
		id
	FROM
		acme_orders
	WHERE
		acme_location = $1
	`

	row := tx.QueryRowContext(ctx, query, payload.Location)
	err = row.Scan(&newId)
	// if err == nil, record was found. return the existingId and a corresponding error
	if err == nil {
		return newId, orders.ErrOrderExists
	} else if !errors.Is(err, sql.ErrNoRows) {
		// any other error, except no rows (because that indicates this order is truly new to db)
		return -2, err
	}

	// deal with error obj
	acmeErr, err := structToNullableJsonString(payload.Error)
	if err != nil {
		return -2, err
	}

	// slices
	dnsIds, err := sliceToJsonString(payload.DnsIds, false)
	if err != nil {
		return -2, err
	}

	authz, err := sliceToJsonString(payload.Authorizations, false)
	if err != nil {
		return -2, err
	}

	query = `
	INSERT INTO
		acme_orders
			(
				certificate_id,
				acme_account_id,
				status,
				known_revoked,
				expires,
				dns_identifiers,
				error,
				authorizations,
				finalize,
				profile,
				acme_location,
				created_at,
				updated_at
			)
	VALUES
			(
				$1,
				$2,
				$3,
				$4,
				$5,
				$6,
				$7,
				$8,
				$9,
				$10,
				$11,
				$12,
				$13
			)
	RETURNING
		id
	`

	err = tx.QueryRowContext(ctx, query,
		payload.CertId,
		payload.AccountId,
		payload.Status,
		payload.KnownRevoked,
		timePointerToNullInt64(payload.Expires),
		dnsIds,
		acmeErr,
		authz,
		payload.Finalize,
		payload.Profile,
		payload.Location,
		payload.CreatedAt,
		payload.UpdatedAt,
	).Scan(&newId)

	if err != nil {
		return -2, err
	}

	err = tx.Commit()
	if err != nil {
		return -2, err
	}

	return newId, nil
}
