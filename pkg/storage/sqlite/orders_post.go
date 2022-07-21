package sqlite

import (
	"context"
	"legocerthub-backend/pkg/acme"
	"legocerthub-backend/pkg/domain/certificates"
)

// newOrderToDb translates the ACME new order response into the fields we want to save
// in the database
func newOrderRespToDb(cert certificates.Certificate, response acme.OrderResponse) orderDb {
	var order orderDb

	// prevent nil pointer
	order.acmeAccount = new(accountDb)
	order.privateKey = new(keyDb)
	order.certificate = new(certificateDb)

	order.acmeAccount.id = intToNullInt32(cert.AcmeAccount.ID)
	order.privateKey.id = intToNullInt32(cert.PrivateKey.ID)
	order.certificate.id = intToNullInt32(cert.ID)

	order.location = stringToNullString(&response.Location)

	order.status = stringToNullString(&response.Status)
	order.expires = intToNullInt32(response.Expires.ToUnixTime())
	order.dnsIdentifiers = sliceToCommaNullString(response.DnsIdentifiers())
	order.authorizations = sliceToCommaNullString(response.Authorizations)
	order.finalize = stringToNullString(&response.Finalize)

	order.createdAt = timeNow()
	order.updatedAt = order.createdAt

	return order
}

// PostNewOrder makes a new order in the db with the cert information and
// ACME response from posting the new order to ACME
func (store *Storage) PostNewOrder(cert certificates.Certificate, response acme.OrderResponse) (newId int, err error) {
	// Load response into db obj
	orderDb := newOrderRespToDb(cert, response)

	ctx, cancel := context.WithTimeout(context.Background(), store.Timeout)
	defer cancel()

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

	err = store.Db.QueryRowContext(ctx, query,
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

	if err != nil {
		return -2, err
	}

	// TODO: Handle 0 rows updated.

	return newId, nil
}
