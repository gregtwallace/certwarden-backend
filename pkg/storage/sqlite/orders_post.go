package sqlite

import (
	"context"
	"database/sql"
	"legocerthub-backend/pkg/acme"
	"legocerthub-backend/pkg/domain/certificates"
)

// newOrderToDb translates the order response into the fields we want to save
// in the database
func newOrderToDb(cert certificates.Certificate, order acme.Order) orderDb {
	// create db obj
	var orderDb orderDb

	// prevent nil pointer
	orderDb.acmeAccount = new(accountDb)
	orderDb.certificate = new(certificateDb)

	orderDb.acmeAccount.id = intToNullInt32(cert.AcmeAccount.ID)
	orderDb.certificate.id = intToNullInt32(cert.ID)

	orderDb.location = stringToNullString(&order.Location)
	orderDb.status = stringToNullString(&order.Status)
	orderDb.knownRevoked = false
	orderDb.err = acmeErrorToNullString(order.Error)
	orderDb.expires = intToNullInt32(order.Expires.ToUnixTime())
	orderDb.dnsIdentifiers = sliceToCommaNullString(order.Identifiers.DnsIdentifiers())
	orderDb.authorizations = sliceToCommaNullString(order.Authorizations)
	orderDb.finalize = stringToNullString(&order.Finalize)
	orderDb.certificateUrl = stringToNullString(&order.Certificate)

	orderDb.createdAt = timeNow()
	orderDb.updatedAt = orderDb.createdAt

	return orderDb
}

// PostNewOrder makes a new order in the db with the cert information and
// ACME response from posting the new order to ACME. If the order already exists
// the order is updated with the newest information.
func (store *Storage) PostNewOrder(cert certificates.Certificate, order acme.Order) (newId int, err error) {
	// Load response into db obj
	orderDb := newOrderToDb(cert, order)

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
					certificate_id,
					acme_location,
					status,
					known_revoked,
					expires,
					dns_identifiers,
					authorizations,
					finalize,
					certificate_url,
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
			orderDb.acmeAccount.id,
			orderDb.certificate.id,
			orderDb.location,
			orderDb.status,
			orderDb.knownRevoked,
			orderDb.expires,
			orderDb.dnsIdentifiers,
			orderDb.authorizations,
			orderDb.finalize,
			orderDb.certificateUrl,
			orderDb.createdAt,
			orderDb.updatedAt,
		).Scan(&newId)

	} else {
		// update the existing order with the new details
		err = store.UpdateOrderAcme(newId, order)
		if err != nil {
			return -2, err
		}
		return newId, nil
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
