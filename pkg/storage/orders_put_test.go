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

// TODO: need to rewrite some logic to do proper testing
// PutOrderPemData
// Also, refactor methods to use pointers

// Useful for removing CR chars from pem
// UPDATE acme_orders
// SET
//   pem = REPLACE(pem, CHAR(13), '')
// WHERE id = [insert id here];

func TestPutOrderACME(t *testing.T) {
	// modifications of existing orders
	ord203updated := ord203
	ord203updated.Status = "pending"
	ord203updated.Expires = new(time.Unix(1747346029, 0))
	ord203updated.DnsIdentifiers = []string{"dns id a", "another dns id b"}
	ord203updated.Error = &acme.Error{
		Status: 888,
		Type:   "some:type",
		Detail: "more info here",
	}
	ord203updated.Authorizations = []string{"an auth a", "another auth b"}
	ord203updated.Finalize = "new finalize value 2"
	ord203updated.Profile = new("someprof")
	ord203updated.CertificateUrl = new("example.com/cert/here")
	ord203updated.UpdatedAt = time.Unix(45533444, 0)

	ord204updated := ord204
	ord204updated.Status = "valid"
	ord204updated.Expires = nil
	ord204updated.DnsIdentifiers = []string{"dns id 1", "another dns id 2"}
	ord204updated.Error = nil
	ord204updated.Authorizations = []string{"an auth 1", "another auth 2"}
	ord204updated.Finalize = "new finalize value"
	ord204updated.Profile = nil
	ord204updated.CertificateUrl = nil
	ord204updated.UpdatedAt = time.Unix(433444, 0)

	// test cases
	testCases := []struct {
		payload orders.UpdateAcmeOrderPayload

		expectedPutErr error
		expectedGetOrd orders.Order
		expectedGetErr error
	}{
		// invalid id (negative)
		{
			orders.UpdateAcmeOrderPayload{
				OrderID:   -9,
				UpdatedAt: time.Unix(5555, 0),
			},
			storage.ErrWrongUpdateRowCount,
			orders.Order{},
			sql.ErrNoRows,
		},
		// invalid id (positive)
		{
			orders.UpdateAcmeOrderPayload{
				OrderID:   1234,
				UpdatedAt: time.Unix(4444, 0),
			},
			storage.ErrWrongUpdateRowCount,
			orders.Order{},
			sql.ErrNoRows,
		},
		// all vals populated
		{
			orders.UpdateAcmeOrderPayload{
				OrderID: 203,
				Status:  "pending",
				Expires: new(time.Unix(1747346029, 0)),
				DnsIds:  []string{"dns id a", "another dns id b"},
				Error: &acme.Error{
					Status: 888,
					Type:   "some:type",
					Detail: "more info here",
				},
				Authorizations: []string{"an auth a", "another auth b"},
				Finalize:       "new finalize value 2",
				Profile:        new("someprof"),
				CertificateUrl: new("example.com/cert/here"),
				UpdatedAt:      time.Unix(45533444, 0),
			},
			nil,
			ord203updated,
			nil,
		},
		// ACME Error value -> NULL; Expires value -> NULL
		// Profile -> NULL; Certificate URL -> NULL
		{
			orders.UpdateAcmeOrderPayload{
				OrderID:        204,
				Status:         "valid",
				Expires:        nil,
				DnsIds:         []string{"dns id 1", "another dns id 2"},
				Error:          nil,
				Authorizations: []string{"an auth 1", "another auth 2"},
				Finalize:       "new finalize value",
				Profile:        nil,
				CertificateUrl: nil,
				UpdatedAt:      time.Unix(433444, 0),
			},
			nil,
			ord204updated,
			nil,
		},
	}

	// create testing service
	store, err := openStorageWithTestData(t, "putorderacme")
	if err != nil {
		t.Fatal(err)
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d: order id: %d)", i, tc.payload.OrderID), func(t *testing.T) {
			err := store.PutOrderACME(&tc.payload)
			if !helpers_test.ErrorsIs(err, tc.expectedPutErr) {
				t.Errorf("expected put order acme error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedPutErr), helpers_test.ErrorToVal(err))
			}

			ord, err := store.GetOneOrder(tc.payload.OrderID)
			if !helpers_test.ErrorsIs(err, tc.expectedGetErr) {
				t.Errorf("expected get error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedGetErr), helpers_test.ErrorToVal(err))
			}

			compareOrder(t, &ord, &tc.expectedGetOrd)
		})
	}
}

func TestPutOrderRenewalInfo(t *testing.T) {
	// modifications of existing orders
	ord203newARI := ord203
	ord203newARI.RenewalInfo = &orders.RenewalInfo{
		SuggestedWindow: struct {
			Start time.Time "json:\"start\""
			End   time.Time "json:\"end\""
		}{
			Start: time.Unix(1581937245, 0),
			End:   time.Unix(1582014934, 0),
		},
		ExplanationURL: new("https://www.example.com/"),
		RetryAfter:     new(time.Unix(1579410250, 0)),
	}
	ord203newARI.UpdatedAt = time.Unix(132435555, 0)

	ord204newARI := ord204
	ord204newARI.RenewalInfo = &orders.RenewalInfo{
		SuggestedWindow: struct {
			Start time.Time "json:\"start\""
			End   time.Time "json:\"end\""
		}{
			Start: time.Unix(1581137265, 0),
			End:   time.Unix(1582114936, 0),
		},
		ExplanationURL: nil,
		RetryAfter:     new(time.Unix(1579410150, 0)),
	}
	ord204newARI.UpdatedAt = time.Unix(152435555, 0)

	// test cases
	testCases := []struct {
		payload orders.UpdateRenewalInfoPayload

		expectedPutErr error
		expectedGetOrd orders.Order
		expectedGetErr error
	}{
		// invalid id (negative)
		{
			orders.UpdateRenewalInfoPayload{
				OrderID: -9,
				RenewalInfo: &orders.RenewalInfo{
					SuggestedWindow: struct {
						Start time.Time "json:\"start\""
						End   time.Time "json:\"end\""
					}{
						Start: time.Unix(1781937245, 0),
						End:   time.Unix(1782014934, 0),
					},
					ExplanationURL: nil,
					RetryAfter:     new(time.Unix(1779410250, 0)), //
				},
				UpdatedAt: time.Unix(2334234234, 0),
			},

			storage.ErrWrongUpdateRowCount,
			orders.Order{},
			sql.ErrNoRows,
		},
		// invalid id (positive)
		{
			orders.UpdateRenewalInfoPayload{
				OrderID: 1234,
				RenewalInfo: &orders.RenewalInfo{
					SuggestedWindow: struct {
						Start time.Time "json:\"start\""
						End   time.Time "json:\"end\""
					}{
						Start: time.Unix(1781937245, 0),
						End:   time.Unix(1782014934, 0),
					},
					ExplanationURL: nil,
					RetryAfter:     new(time.Unix(1779410250, 0)), //
				},
				UpdatedAt: time.Unix(2334234234, 0),
			},

			storage.ErrWrongUpdateRowCount,
			orders.Order{},
			sql.ErrNoRows,
		},
		// has existing ARI
		{
			orders.UpdateRenewalInfoPayload{
				OrderID: 203,
				RenewalInfo: &orders.RenewalInfo{
					SuggestedWindow: struct {
						Start time.Time "json:\"start\""
						End   time.Time "json:\"end\""
					}{
						Start: time.Unix(1581937245, 0),
						End:   time.Unix(1582014934, 0),
					},
					ExplanationURL: new("https://www.example.com/"),
					RetryAfter:     new(time.Unix(1579410250, 0)),
				},
				UpdatedAt: time.Unix(132435555, 0),
			},

			nil,
			ord203newARI,
			nil,
		},
		// has NULL ari
		{
			orders.UpdateRenewalInfoPayload{
				OrderID: 204,
				RenewalInfo: &orders.RenewalInfo{
					SuggestedWindow: struct {
						Start time.Time "json:\"start\""
						End   time.Time "json:\"end\""
					}{
						Start: time.Unix(1581137265, 0),
						End:   time.Unix(1582114936, 0),
					},
					ExplanationURL: nil,
					RetryAfter:     new(time.Unix(1579410150, 0)),
				},
				UpdatedAt: time.Unix(152435555, 0),
			},

			nil,
			ord204newARI,
			nil,
		},
	}

	// create testing service
	store, err := openStorageWithTestData(t, "putorderrenewalinfo")
	if err != nil {
		t.Fatal(err)
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d: order id: %d)", i, tc.payload.OrderID), func(t *testing.T) {
			err := store.PutOrderRenewalInfo(tc.payload)
			if !helpers_test.ErrorsIs(err, tc.expectedPutErr) {
				t.Errorf("expected put order renewal info error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedPutErr), helpers_test.ErrorToVal(err))
			}

			ord, err := store.GetOneOrder(tc.payload.OrderID)
			if !helpers_test.ErrorsIs(err, tc.expectedGetErr) {
				t.Errorf("expected get error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedGetErr), helpers_test.ErrorToVal(err))
			}

			compareOrder(t, &ord, &tc.expectedGetOrd)
		})
	}
}

func TestPutOrderStatusInvalid(t *testing.T) {
	// modifications of existing orders
	ord203nowInvalid := ord203
	ord203nowInvalid.Status = "invalid"
	ord203nowInvalid.UpdatedAt = time.Unix(125435555, 0)

	ord204newUpdatedAt := ord204
	ord204newUpdatedAt.UpdatedAt = time.Unix(111135555, 0)

	// test cases
	testCases := []struct {
		ordID     int
		updatedAt time.Time

		expectedPutErr error
		expectedGetOrd orders.Order
		expectedGetErr error
	}{
		// invalid id (negative)
		{
			-3,
			time.Unix(265465487, 0),

			storage.ErrWrongUpdateRowCount,
			orders.Order{},
			sql.ErrNoRows,
		},
		// invalid id (positive)
		{
			452,
			time.Unix(5435645, 0),

			storage.ErrWrongUpdateRowCount,
			orders.Order{},
			sql.ErrNoRows,
		},
		// set to 'invalid'
		{
			203,
			time.Unix(125435555, 0),

			nil,
			ord203nowInvalid,
			nil,
		},
		// already at 'invalid' state
		{
			204,
			time.Unix(111135555, 0),

			nil,
			ord204newUpdatedAt,
			nil,
		},
	}

	// create testing service
	store, err := openStorageWithTestData(t, "putorderstatusinvalid")
	if err != nil {
		t.Fatal(err)
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d: order id: %d)", i, tc.ordID), func(t *testing.T) {
			err := store.PutOrderStatusInvalid(tc.ordID, tc.updatedAt)
			if !helpers_test.ErrorsIs(err, tc.expectedPutErr) {
				t.Errorf("expected put order invalid error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedPutErr), helpers_test.ErrorToVal(err))
			}

			ord, err := store.GetOneOrder(tc.ordID)
			if !helpers_test.ErrorsIs(err, tc.expectedGetErr) {
				t.Errorf("expected get error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedGetErr), helpers_test.ErrorToVal(err))
			}

			compareOrder(t, &ord, &tc.expectedGetOrd)
		})
	}
}

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
				t.Errorf("expected put order finalized key error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedPutErr), helpers_test.ErrorToVal(err))
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
				t.Errorf("expected put order pem data error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedPutErr), helpers_test.ErrorToVal(err))
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
