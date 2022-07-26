package sqlite

import (
	"context"
	"database/sql"
	"legocerthub-backend/pkg/domain/orders"
)

// newOrderToDb translates the ACME new order response into the fields we want to save
// in the database
func newOrderToDb(newOrder orders.Order) orderDb {
	var order orderDb

	// prevent nil pointer
	order.acmeAccount = new(accountDb)
	order.privateKey = new(keyDb)
	order.certificate = new(certificateDb)

	order.acmeAccount.id = intToNullInt32(newOrder.Certificate.AcmeAccount.ID)
	order.privateKey.id = intToNullInt32(newOrder.Certificate.PrivateKey.ID)
	order.certificate.id = intToNullInt32(newOrder.Certificate.ID)

	order.location = stringToNullString(&newOrder.Acme.Location)

	order.status = stringToNullString(&newOrder.Acme.Status)
	order.expires = intToNullInt32(newOrder.Acme.Expires.ToUnixTime())
	order.dnsIdentifiers = sliceToCommaNullString(newOrder.Acme.Identifiers.DnsIdentifiers())
	order.authorizations = sliceToCommaNullString(newOrder.Acme.Authorizations)
	order.finalize = stringToNullString(&newOrder.Acme.Finalize)

	order.createdAt = timeNow()
	order.updatedAt = order.createdAt

	return order
}

// PostNewOrder makes a new order in the db with the cert information and
// ACME response from posting the new order to ACME. If the order already exists
// the order is updated with the newest information.
func (store *Storage) PostNewOrder(newOrder orders.Order) (newId int, err error) {
	// Load response into db obj
	orderDb := newOrderToDb(newOrder)

	ctx, cancel := context.WithTimeout(context.Background(), store.Timeout)
	defer cancel()

	// transaction
	tx, err := store.Db.BeginTx(ctx, nil)
	if err != nil {
		return -2, err
	}
	defer tx.Rollback()

	// check if the order already exists
	query := `
	SELECT id
	FROM
		acme_orders
	WHERE
		acme_location = $1
	`

	row := tx.QueryRowContext(ctx, query, orderDb.location)
	err = row.Scan(&newId)
	if err != nil && err != sql.ErrNoRows {
		return -2, err
	}

	// if doesn't exist (i.e. it is new), insert
	if err == sql.ErrNoRows {
		query := `
		INSERT INTO
			acme_orders
				(
					acme_account_id,
					private_key_id,
					certificate_id,
					acme_location,
					status,
					expires,
					dns_identifiers,
					authorizations,
					finalize,
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
					$11
				)
		RETURNING
			id
		`

		err = tx.QueryRowContext(ctx, query,
			orderDb.acmeAccount.id,
			orderDb.privateKey.id,
			orderDb.certificate.id,
			orderDb.location,
			orderDb.status,
			orderDb.expires,
			orderDb.dnsIdentifiers,
			orderDb.authorizations,
			orderDb.finalize,
			orderDb.createdAt,
			orderDb.updatedAt,
		).Scan(&newId)

	} else {
		// else update existing record
		query := `
		UPDATE
			acme_orders
		SET
			acme_account_id = $1,
			private_key_id = $2,
			certificate_id = $3,
			status = $4,
			expires = $5,
			dns_identifiers = $6,
			authorizations = $7,
			finalize = $8,
			updated_at = $9
		WHERE
			acme_location = $10
		RETURNING
			id
		`

		err = tx.QueryRowContext(ctx, query,
			orderDb.acmeAccount.id,
			orderDb.privateKey.id,
			orderDb.certificate.id,
			orderDb.status,
			orderDb.expires,
			orderDb.dnsIdentifiers,
			orderDb.authorizations,
			orderDb.finalize,
			orderDb.updatedAt,
			orderDb.location,
		).Scan(&newId)

	}

	if err != nil {
		return -2, err
	}

	err = tx.Commit()
	if err != nil {
		return -2, err
	}

	// TODO: Handle 0 rows updated.

	return newId, nil
}
