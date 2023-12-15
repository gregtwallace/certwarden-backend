package sqlite

import (
	"context"
	"database/sql"
	"legocerthub-backend/pkg/domain/orders"
)

// PostNewOrder makes a new order in the db. An error is returned if the order
// location already exists (or any other error)
func (store *Storage) PostNewOrder(payload orders.NewOrderAcmePayload) (newId int, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), store.timeout)
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
	} else if err != sql.ErrNoRows {
		// any other error, except no rows (because that indicates this order is truly new to db)
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
				$12
			)
	RETURNING
		id
	`

	err = tx.QueryRowContext(ctx, query,
		payload.CertId,
		payload.AccountId,
		payload.Status,
		payload.KnownRevoked,
		payload.Expires,
		makeJsonStringSlice(payload.DnsIds),
		payload.Error,
		makeJsonStringSlice(payload.Authorizations),
		payload.Finalize,
		payload.Location,
		payload.CreatedAt,
		payload.UpdatedAt,
	).Scan(&newId)

	err = tx.Commit()
	if err != nil {
		return -2, err
	}

	// TODO: Handle 0 rows updated.

	return newId, nil
}
