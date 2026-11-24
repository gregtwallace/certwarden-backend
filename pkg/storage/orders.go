package storage

import (
	"certwarden-backend/pkg/acme"
	"certwarden-backend/pkg/domain/orders"
	"certwarden-backend/pkg/domain/private_keys"
	"database/sql"
	"encoding/json"
	"time"
)

// orderDb is a single acme order, as database table fields
// corresponds to orders.Order
type orderDb struct {
	id             int
	certificate    certificateDb
	location       string
	status         string
	knownRevoked   bool
	acmeErr        sql.NullString // json: acme.Error
	expires        sql.NullInt64
	dnsIdentifiers []byte // json: []string
	authorizations []byte // json: []string
	finalize       string
	finalizedKey   keyDb
	certificateUrl sql.NullString
	pem            sql.NullString
	chainRootCN    sql.NullString
	validFrom      sql.NullInt64
	validTo        sql.NullInt64
	createdAt      int64
	updatedAt      int64
	profile        sql.NullString
	renewalInfo    sql.NullString
}

func (order *orderDb) toOrder() (*orders.Order, error) {
	// handle if key is not null (id value would not be okay from coalesce if null)
	var key *private_keys.Key
	if order.finalizedKey.id >= 0 {
		key = order.finalizedKey.toKey()
	}

	// handle acme Error
	acmeErr, err := jsonStringToNullableStruct[acme.Error](nullStringToString(order.acmeErr))
	if err != nil {
		return nil, err
	}

	// convert cert
	cert, err := order.certificate.toCertificate()
	if err != nil {
		return nil, err
	}

	// renewal info
	ri := orders.UnmarshalRenewalInfo([]byte(order.renewalInfo.String))
	if !order.renewalInfo.Valid {
		ri = nil
	}

	// slices
	dnsIds := []string{}
	err = json.Unmarshal(order.dnsIdentifiers, &dnsIds)
	if err != nil {
		return nil, err
	}

	authz := []string{}
	err = json.Unmarshal(order.authorizations, &authz)
	if err != nil {
		return nil, err
	}

	return &orders.Order{
		ID:             order.id,
		Certificate:    *cert,
		Location:       order.location,
		Status:         order.status,
		KnownRevoked:   order.knownRevoked,
		Error:          acmeErr,
		Expires:        nullInt64UnixToTime(order.expires),
		DnsIdentifiers: dnsIds,
		Authorizations: authz,
		Finalize:       order.finalize,
		FinalizedKey:   key,
		CertificateUrl: nullStringToString(order.certificateUrl),
		Pem:            nullStringToString(order.pem),
		ValidFrom:      nullInt64UnixToTime(order.validFrom),
		ValidTo:        nullInt64UnixToTime(order.validTo),
		ChainRootCN:    nullStringToString(order.chainRootCN),
		CreatedAt:      time.Unix(order.createdAt, 0),
		UpdatedAt:      time.Unix(order.updatedAt, 0),
		Profile:        nullStringToString(order.profile),
		RenewalInfo:    ri,
	}, nil
}
