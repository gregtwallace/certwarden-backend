package storage_test

import (
	"certwarden-backend/pkg/acme"
	"certwarden-backend/pkg/domain/orders"
	"certwarden-backend/pkg/helpers_test"
	"certwarden-backend/pkg/storage"
	"database/sql"
	"fmt"
	"testing"
	"time"
)

// TODO:
// PutOrderAcme
// PutRenewalInfo
// PutOrderInvalid
// PutOrderPemData // TODO: need to rewrite some logic to do proper testing

// Useful for removing CR chars from pem
// UPDATE acme_orders
// SET
//   pem = REPLACE(pem, CHAR(13), '')
// WHERE id = [insert id here];

func TestPutOrderFinalizedKey(t *testing.T) {
	// modifications of existing orders
	ord203wKey58 := ord203
	ord203wKey58.FinalizedKey = &key58
	ord203wKey58.UpdatedAt = time.Unix(25435555, 0)

	ord204wKey62 := ord204
	ord204wKey62.FinalizedKey = &key62
	ord204wKey62.UpdatedAt = time.Unix(11135555, 0)

	// test cases
	testCases := []struct {
		ordID      int
		finalKeyID int
		updatedAt  time.Time

		expectedPutErr error
		expectedGetOrd orders.Order
		expectedGetErr error
	}{
		// invalid id (negative)
		{
			-3,
			58,
			time.Unix(265465487, 0),

			storage.ErrWrongUpdateRowCount,
			orders.Order{},
			sql.ErrNoRows,
		},
		// invalid id (positive)
		{
			452,
			58,
			time.Unix(5435645, 0),

			storage.ErrWrongUpdateRowCount,
			orders.Order{},
			sql.ErrNoRows,
		},
		// valid order ID but invalid keyID
		{
			206,
			2,
			time.Unix(5435555, 0),

			helpers_test.NewTestErrorStringComp("FOREIGN KEY constraint failed"),
			ord206,
			nil,
		},
		// valid order ID & valid keyID (overwriting existing finalKeyID)
		{
			203,
			58,
			time.Unix(25435555, 0),

			nil,
			ord203wKey58,
			nil,
		},
		// valid order ID & valid keyID (overwriting NULL finalKeyID)
		{
			204,
			62,
			time.Unix(11135555, 0),

			nil,
			ord204wKey62,
			nil,
		},
	}

	// create testing service
	store, err := openStorageWithTestData(t, "putorderfinalizedkey")
	if err != nil {
		t.Fatal(err)
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d: order id: %d)", i, tc.ordID), func(t *testing.T) {
			err := store.PutOrderFinalizedKey(tc.ordID, tc.finalKeyID, tc.updatedAt)
			if !helpers_test.ErrorsIs(err, tc.expectedPutErr) {
				t.Errorf("expected put order revoke error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedPutErr), helpers_test.ErrorToVal(err))
			}

			ord, err := store.GetOneOrder(tc.ordID)
			if !helpers_test.ErrorsIs(err, tc.expectedGetErr) {
				t.Errorf("expected get error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedGetErr), helpers_test.ErrorToVal(err))
			}

			compareOrder(t, &ord, &tc.expectedGetOrd)
		})
	}
}

func TestPutOrderPemData(t *testing.T) {
	testCases := []struct {
		ordID   int
		payload orders.CertPayload

		expectedPutErr error
		expectedGetOrd orders.Order
		expectedGetErr error
	}{
		// invalid id -2
		{
			-2,
			orders.CertPayload{
				AcmeCert:    new(acme.Certificate{}),
				RenewalInfo: new(orders.RenewalInfo{}),
				UpdatedAt:   time.Unix(2, 0),
			},
			storage.ErrWrongUpdateRowCount,
			orders.Order{},
			sql.ErrNoRows,
		},
		// invalid id 444
		{
			444,
			orders.CertPayload{
				AcmeCert:    new(acme.Certificate{}),
				RenewalInfo: new(orders.RenewalInfo{}),
				UpdatedAt:   time.Unix(2, 0),
			},
			storage.ErrWrongUpdateRowCount,
			orders.Order{},
			sql.ErrNoRows,
		},
		// TODO: Addl tests
	}

	// create testing service
	store, err := openStorageWithTestData(t, "putorderpemdata")
	if err != nil {
		t.Fatal(err)
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d: order id: %d)", i, tc.ordID), func(t *testing.T) {
			err := store.PutOrderPemData(tc.ordID, tc.payload)
			if !helpers_test.ErrorsIs(err, tc.expectedPutErr) {
				t.Errorf("expected put order revoke error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedPutErr), helpers_test.ErrorToVal(err))
			}

			ord, err := store.GetOneOrder(tc.ordID)
			if !helpers_test.ErrorsIs(err, tc.expectedGetErr) {
				t.Errorf("expected get error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedGetErr), helpers_test.ErrorToVal(err))
			}

			compareOrder(t, &ord, &tc.expectedGetOrd)
		})
	}
}

func TestPutOrderRevoke(t *testing.T) {
	// modifications of existing orders
	ord204nowRevoked := ord204
	ord204nowRevoked.KnownRevoked = true
	ord204nowRevoked.UpdatedAt = time.Unix(45345345, 0)

	ord206newUpdatedAt := ord206
	ord206newUpdatedAt.UpdatedAt = time.Unix(55663333, 0)

	// test cases
	testCases := []struct {
		ordId     int
		updatedAt time.Time

		expectedPutErr error
		expectedGetOrd orders.Order
		expectedGetErr error
	}{
		// invalid id -1
		{
			-1,
			time.Unix(77755552, 0),
			storage.ErrWrongUpdateRowCount,
			orders.Order{},
			sql.ErrNoRows,
		},
		// invalid id 555
		{
			555,
			time.Unix(76655552, 0),
			storage.ErrWrongUpdateRowCount,
			orders.Order{},
			sql.ErrNoRows,
		},
		// valid starting as known revoked false
		{
			204,
			time.Unix(45345345, 0),
			nil,
			ord204nowRevoked,
			nil,
		},
		// valid starting as known revoked true
		{
			206,
			time.Unix(55663333, 0),
			nil,
			ord206newUpdatedAt,
			nil,
		},
	}

	// create testing service
	store, err := openStorageWithTestData(t, "putrevokeorder")
	if err != nil {
		t.Fatal(err)
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d: order id: %d)", i, tc.ordId), func(t *testing.T) {
			err := store.PutOrderRevoke(tc.ordId, tc.updatedAt)
			if !helpers_test.ErrorsIs(err, tc.expectedPutErr) {
				t.Errorf("expected put order revoke error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedPutErr), helpers_test.ErrorToVal(err))
			}

			ord, err := store.GetOneOrder(tc.ordId)
			if !helpers_test.ErrorsIs(err, tc.expectedGetErr) {
				t.Errorf("expected get error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedGetErr), helpers_test.ErrorToVal(err))
			}

			compareOrder(t, &ord, &tc.expectedGetOrd)
		})
	}
}
