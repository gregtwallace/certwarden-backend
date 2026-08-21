package storage_test

import (
	"certwarden-backend/pkg/domain/orders"
	"certwarden-backend/pkg/helpers_test"
	"certwarden-backend/pkg/pagination_sort"
	"certwarden-backend/pkg/storage"
	"database/sql"
	"fmt"
	"slices"
	"testing"
	"time"
)

func TestGetAllValidCurrentOrders(t *testing.T) {
	testCases := []struct {
		q                 pagination_sort.Query
		expectedTotalCt   int
		expectedResultLen int

		testIndx       int
		expectedAtIndx orders.Order
	}{
		{pagination_sort.Query{}, 2, 2, 0, ord198},
		{pagination_sort.Query{}, 2, 2, 1, ord203},
		{queryBuilderForTest(1, 0, "subject", false), 2, 1, 0, ord203},
		{queryBuilderForTest(1, 1, "subject", false), 2, 1, 0, ord198},
		{queryBuilderForTest(30, 0, "last_access", false), 2, 2, 0, ord203},
		{queryBuilderForTest(4, 0, "id", true), 2, 2, 1, ord203},
	}

	// create testing service
	store, err := openStorageWithTestData(t, "getallcerts")
	if err != nil {
		t.Fatal(err)
	}

	// override timenow
	revertToDefaultTimeNow := storage.SetTimeNow(time.Unix(1779991589, 0))
	t.Cleanup(revertToDefaultTimeNow)

	// run tests
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("#%d (order id: %d)", i, tc.expectedAtIndx.ID), func(t *testing.T) {
			ords, totalCt, err := store.GetAllValidCurrentOrders(tc.q)
			if err != nil {
				t.Errorf("get all valid current orders failed")
				return
			}

			if totalCt != tc.expectedTotalCt {
				t.Errorf("incorrect total count, expected '%d' but got '%d'", tc.expectedTotalCt, totalCt)
			}
			if len(ords) != tc.expectedResultLen {
				t.Errorf("incorrect result length, expected '%d' but got '%d'", tc.expectedResultLen, len(ords))
			}
			if tc.testIndx <= len(ords)-1 {
				compareOrder(t, &ords[tc.testIndx], &tc.expectedAtIndx)
			} else {
				t.Errorf("couldnt test result at index '%d' because length of result array was only '%d'", tc.testIndx, len(ords))
			}
		})
	}
}

func TestGetOrdersByCert(t *testing.T) {
	testCases := []struct {
		certId            int
		q                 pagination_sort.Query
		expectedTotalCt   int
		expectedResultLen int

		testIndx       int
		expectedAtIndx orders.Order
	}{
		{35, pagination_sort.Query{}, 2, 2, 0, ord204},
		{28, pagination_sort.Query{}, 21, 21, 19, ord175},
		{18, queryBuilderForTest(5, 0, "id", true), 31, 5, 0, ord203},
		{18, queryBuilderForTest(5, 4, "valid_to", true), 31, 5, 0, ord203},
		{33, queryBuilderForTest(300, 0, "status", false), 10, 10, 0, ord186},
	}

	// create testing service
	store, err := openStorageWithTestData(t, "getordersbycert")
	if err != nil {
		t.Fatal(err)
	}

	// run tests
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("#%d (order id: %d)", i, tc.expectedAtIndx.ID), func(t *testing.T) {
			ords, totalCt, err := store.GetOrdersByCert(tc.certId, tc.q)
			if err != nil {
				t.Errorf("get orders by cert failed: %s", err)
				return
			}

			if totalCt != tc.expectedTotalCt {
				t.Errorf("incorrect total count, expected '%d' but got '%d'", tc.expectedTotalCt, totalCt)
			}
			if len(ords) != tc.expectedResultLen {
				t.Errorf("incorrect result length, expected '%d' but got '%d'", tc.expectedResultLen, len(ords))
			}
			if tc.testIndx <= len(ords)-1 {
				compareOrder(t, &ords[tc.testIndx], &tc.expectedAtIndx)
			} else {
				t.Errorf("couldnt test result at index '%d' because length of result array was only '%d'", tc.testIndx, len(ords))
			}
		})
	}
}

func TestGetAllIncompleteOrderIds(t *testing.T) {
	expectedOrderIDs := []int{99, 98, 97, 96}

	// create testing service
	store, err := openStorageWithTestData(t, "getallincompleteorderids")
	if err != nil {
		t.Fatal(err)
	}

	// run test
	ordIDs, err := store.GetAllIncompleteOrderIds()
	if !helpers_test.ErrorsIs(err, nil) {
		t.Errorf("expected error '%s' but got '%s'", helpers_test.ErrorToVal(nil), helpers_test.ErrorToVal(err))
	}

	if !slices.Equal(ordIDs, expectedOrderIDs) {
		t.Errorf("expected order ids '%d' but got '%d'", expectedOrderIDs, ordIDs)
	}
}

func TestGetNewestIncompleteCertOrderId(t *testing.T) {
	testCases := []struct {
		certID int

		expectedOrderID int
		expectedErr     error
	}{
		{-1, -2, sql.ErrNoRows},
		{666, -2, sql.ErrNoRows},
		{18, 99, nil},           // 18: newest is valid, but there is a pending order out there
		{35, -2, sql.ErrNoRows}, // 35: no valid order
		{28, -2, sql.ErrNoRows}, // 28: newest valid is expired
		{31, -2, sql.ErrNoRows}, // 31: all valid orders but expired
		{33, -2, sql.ErrNoRows}, // 33: newest is valid but revoked, drop back to next newest valid
		{26, -2, sql.ErrNoRows}, // 26: newest valid is expired

	}

	// create testing service
	store, err := openStorageWithTestData(t, "getnewestincompletecertorderid")
	if err != nil {
		t.Fatal(err)
	}

	// run tests
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("#%d (id: %d)", i, tc.certID), func(t *testing.T) {
			ordID, err := store.GetNewestIncompleteCertOrderId(tc.certID)
			if !helpers_test.ErrorsIs(err, tc.expectedErr) {
				t.Errorf("expected error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedErr), helpers_test.ErrorToVal(err))
			}

			if ordID != tc.expectedOrderID {
				t.Errorf("expected order id '%d' but got '%d'", tc.expectedOrderID, ordID)
			}
		})
	}
}

func TestGetOrders(t *testing.T) {
	testCases := []struct {
		ids []int

		expectedOrds []orders.Order
		expectedErr  error
	}{
		{[]int{-1}, []orders.Order{}, sql.ErrNoRows},           // just one bad
		{[]int{-1, 666}, []orders.Order{}, sql.ErrNoRows},      // just two bad
		{[]int{666, 203}, []orders.Order{ord203}, nil},         // one bad, one good
		{[]int{198, 203}, []orders.Order{ord198, ord203}, nil}, // two good
	}

	// create testing service
	store, err := openStorageWithTestData(t, "getorders")
	if err != nil {
		t.Fatal(err)
	}

	// run tests
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("#%d (id: %d)", i, tc.ids), func(t *testing.T) {
			ords, err := store.GetOrders(tc.ids)
			if !helpers_test.ErrorsIs(err, tc.expectedErr) {
				t.Errorf("expected error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedErr), helpers_test.ErrorToVal(err))
			}

			if len(ords) != len(tc.expectedOrds) {
				t.Errorf("expected len of '%d' but actual len is '%d' (aborting element comparison)", len(tc.expectedOrds), len(ords))
				return
			}

			for i := range ords {
				compareOrder(t, &ords[i], &tc.expectedOrds[i])
			}

		})
	}
}

func TestGetOneOrder(t *testing.T) {
	testCases := []struct {
		id int

		expectedOrd orders.Order
		expectedErr error
	}{
		{-1, orders.Order{}, sql.ErrNoRows},
		{666, orders.Order{}, sql.ErrNoRows},
		{203, ord203, nil},
		{198, ord198, nil},
		{99, ord99, nil}, // many nulls
	}

	// create testing service
	store, err := openStorageWithTestData(t, "getoneorder")
	if err != nil {
		t.Fatal(err)
	}

	// run tests
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("#%d (id: %d)", i, tc.id), func(t *testing.T) {
			ord, err := store.GetOneOrder(tc.id)
			if !helpers_test.ErrorsIs(err, tc.expectedErr) {
				t.Errorf("expected error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedErr), helpers_test.ErrorToVal(err))
			}

			compareOrder(t, &ord, &tc.expectedOrd)
		})
	}
}

func TestGetCertNewestValidOrderById(t *testing.T) {
	testCases := []struct {
		id int

		expectedOrd orders.Order
		expectedErr error
	}{
		{-1, orders.Order{}, sql.ErrNoRows},
		{666, orders.Order{}, sql.ErrNoRows},
		{18, ord203, nil},                   // 18: newest is valid, case is wrong (also has a createdAt tie that must be broken by order.id)
		{35, orders.Order{}, sql.ErrNoRows}, // 35: no valid order
		{28, orders.Order{}, sql.ErrNoRows}, // 28: newest valid is expired
		{31, orders.Order{}, sql.ErrNoRows}, // 31: all valid orders but expired
		{33, ord198, nil},                   // 33: newest is valid but revoked, drop back to next newest valid
		{26, orders.Order{}, sql.ErrNoRows}, // 26: newest valid is expired
	}

	// create testing service
	store, err := openStorageWithTestData(t, "getcertnewestvalidorderbyid")
	if err != nil {
		t.Fatal(err)
	}

	// override timenow
	revertToDefaultTimeNow := storage.SetTimeNow(time.Unix(1779991589, 0))
	t.Cleanup(revertToDefaultTimeNow)

	// run tests
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("#%d (id: %d)", i, tc.id), func(t *testing.T) {
			ord, err := store.GetCertNewestValidOrderById(tc.id)
			if !helpers_test.ErrorsIs(err, tc.expectedErr) {
				t.Errorf("expected error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedErr), helpers_test.ErrorToVal(err))
			}

			compareOrder(t, &ord, &tc.expectedOrd)
		})
	}
}

func TestGetCertNewestValidOrderByName(t *testing.T) {
	testCases := []struct {
		name string

		expectedOrd orders.Order
		expectedErr error
	}{
		{"fake-bad-name", orders.Order{}, sql.ErrNoRows},
		{"", orders.Order{}, sql.ErrNoRows},
		{"serverDEFault", ord203, nil},                                                // 18: newest is valid, case is wrong
		{"STAGING_persist--test005.test.example2.com", orders.Order{}, sql.ErrNoRows}, // 35: no valid order
		{"a0.alias.test.example.com", orders.Order{}, sql.ErrNoRows},                  // 28: newest valid is expired
		{"SomeSmallTest", orders.Order{}, sql.ErrNoRows},                              // 31: all valid orders but expired
		{"STAGING_persist--test007.test.example2.com", ord198, nil},                   // 33: newest is valid but revoked, drop back to next newest valid
		{"test008.test.example.com", orders.Order{}, sql.ErrNoRows},                   // 26: newest valid is expired
	}

	// create testing service
	store, err := openStorageWithTestData(t, "getcertnewestvalidorderbyname")
	if err != nil {
		t.Fatal(err)
	}

	// override timenow
	revertToDefaultTimeNow := storage.SetTimeNow(time.Unix(1779991589, 0))
	t.Cleanup(revertToDefaultTimeNow)

	// run tests
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("#%d (name: %s)", i, tc.name), func(t *testing.T) {
			ord, err := store.GetCertNewestValidOrderByName(tc.name)
			if !helpers_test.ErrorsIs(err, tc.expectedErr) {
				t.Errorf("expected error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedErr), helpers_test.ErrorToVal(err))
			}

			compareOrder(t, &ord, &tc.expectedOrd)
		})
	}
}
