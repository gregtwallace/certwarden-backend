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
		expectedAtIndx *orders.Order
	}{
		{pagination_sort.Query{}, 2, 2, 0, &ord198},
		{pagination_sort.Query{}, 2, 2, 1, &ord203},
		{queryBuilderForTest(1, 0, "subject", false), 2, 1, 0, &ord203},
		{queryBuilderForTest(1, 1, "subject", false), 2, 1, 0, &ord198},
		{queryBuilderForTest(30, 0, "last_access", false), 2, 2, 0, &ord203},
		{queryBuilderForTest(4, 0, "id", true), 2, 2, 1, &ord203},
	}

	// create testing service
	store := openStorageWithTestData(t, "getallcerts")

	// override timenow
	revertToDefaultTimeNow := storage.SetTimeNow(t, time.Unix(1779991589, 0))
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
				compareOrder(t, ords[tc.testIndx], tc.expectedAtIndx)
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
		expectedAtIndx *orders.Order
	}{
		{-1, pagination_sort.Query{}, 0, 0, 0, nil},
		{1, pagination_sort.Query{}, 0, 0, 0, nil},
		{35, pagination_sort.Query{}, 2, 2, 0, &ord204},
		{28, pagination_sort.Query{}, 21, 21, 19, &ord175},
		{18, queryBuilderForTest(5, 0, "id", true), 31, 5, 0, &ord203},
		{18, queryBuilderForTest(5, 4, "valid_to", true), 31, 5, 0, &ord203},
		{33, queryBuilderForTest(300, 0, "status", false), 10, 10, 0, &ord186},
		{26, pagination_sort.Query{}, 7, 7, 1, &ord150},
	}

	// create testing service
	store := openStorageWithTestData(t, "getordersbycert")

	// run tests
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("#%d", i), func(t *testing.T) {
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
				compareOrder(t, ords[tc.testIndx], tc.expectedAtIndx)
			} else if len(ords) != 0 {
				t.Errorf("couldnt test result at index '%d' because length of result array was only '%d'", tc.testIndx, len(ords))
			}
		})
	}
}

func TestGetAllIncompleteOrderIds(t *testing.T) {
	expectedOrderIDs := []int{99, 98, 97, 96}

	// create testing service
	store := openStorageWithTestData(t, "getallincompleteorderids")

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
	store := openStorageWithTestData(t, "getnewestincompletecertorderid")

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

		expectedOrds []*orders.Order
		expectedErr  error
	}{
		{[]int{}, nil, sql.ErrNoRows},                             // nothing requested
		{[]int{-1}, nil, sql.ErrNoRows},                           // just one, but its bad
		{[]int{-1, 666}, nil, sql.ErrNoRows},                      // two bad
		{[]int{666, 203}, nil, storage.ErrWrongRowCount},          // one bad, one good
		{[]int{203}, []*orders.Order{&ord203}, nil},               // one good
		{[]int{198, 203}, []*orders.Order{&ord198, &ord203}, nil}, // two good
		{[]int{150}, []*orders.Order{&ord150}, nil},               // finalized but key has null last_access
	}

	// create testing service
	store := openStorageWithTestData(t, "getorders")

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
				compareOrder(t, ords[i], tc.expectedOrds[i])
			}

		})
	}
}

func TestGetOneOrder(t *testing.T) {
	testCases := []struct {
		id int

		expectedOrd *orders.Order
		expectedErr error
	}{
		{-1, nil, sql.ErrNoRows},
		{666, nil, sql.ErrNoRows},
		{203, &ord203, nil},
		{198, &ord198, nil},
		{99, &ord99, nil}, // many nulls
		{206, &ord206, nil},
		{150, &ord150, nil}, // finalized but key has null last_access
	}

	// create testing service
	store := openStorageWithTestData(t, "getoneorder")

	// run tests
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("#%d (id: %d)", i, tc.id), func(t *testing.T) {
			ord, err := store.GetOneOrder(tc.id)
			if !helpers_test.ErrorsIs(err, tc.expectedErr) {
				t.Errorf("expected error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedErr), helpers_test.ErrorToVal(err))
			}

			compareOrder(t, ord, tc.expectedOrd)
		})
	}
}

func TestGetCertNewestValidOrderById(t *testing.T) {
	testCases := []struct {
		id int

		expectedOrd *orders.Order
		expectedErr error
	}{
		{-1, nil, sql.ErrNoRows},
		{666, nil, sql.ErrNoRows},
		{18, &ord203, nil},       // 18: newest is valid, case is wrong (also has a createdAt tie that must be broken by order.id)
		{35, nil, sql.ErrNoRows}, // 35: no valid order
		{28, nil, sql.ErrNoRows}, // 28: newest valid is expired
		{31, nil, sql.ErrNoRows}, // 31: all valid orders but expired
		{33, &ord198, nil},       // 33: newest is valid but revoked, drop back to next newest valid
		{26, nil, sql.ErrNoRows}, // 26: newest valid is expired
	}

	// create testing service
	store := openStorageWithTestData(t, "getcertnewestvalidorderbyid")

	// override timenow
	revertToDefaultTimeNow := storage.SetTimeNow(t, time.Unix(1779991589, 0))
	t.Cleanup(revertToDefaultTimeNow)

	// run tests
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("#%d (id: %d)", i, tc.id), func(t *testing.T) {
			ord, err := store.GetCertNewestValidOrderById(tc.id)
			if !helpers_test.ErrorsIs(err, tc.expectedErr) {
				t.Errorf("expected error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedErr), helpers_test.ErrorToVal(err))
			}

			compareOrder(t, ord, tc.expectedOrd)
		})
	}
}

func TestGetCertNewestValidOrderByName(t *testing.T) {
	testCases := []struct {
		name string

		expectedOrd *orders.Order
		expectedErr error
	}{
		{"fake-bad-name", nil, sql.ErrNoRows},
		{"", nil, sql.ErrNoRows},
		{"serverDEFault", &ord203, nil},                                    // 18: newest is valid, case is wrong
		{"STAGING_persist--test005.test.example2.com", nil, sql.ErrNoRows}, // 35: no valid order
		{"a0.alias.test.example.com", nil, sql.ErrNoRows},                  // 28: newest valid is expired
		{"SomeSmallTest", nil, sql.ErrNoRows},                              // 31: all valid orders but expired
		{"STAGING_persist--test007.test.example2.com", &ord198, nil},       // 33: newest is valid but revoked, drop back to next newest valid
		{"test008.test.example.com", nil, sql.ErrNoRows},                   // 26: newest valid is expired
	}

	// create testing service
	store := openStorageWithTestData(t, "getcertnewestvalidorderbyname")

	// override timenow
	revertToDefaultTimeNow := storage.SetTimeNow(t, time.Unix(1779991589, 0))
	t.Cleanup(revertToDefaultTimeNow)

	// run tests
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("#%d (name: %s)", i, tc.name), func(t *testing.T) {
			ord, err := store.GetCertNewestValidOrderByName(tc.name)
			if !helpers_test.ErrorsIs(err, tc.expectedErr) {
				t.Errorf("expected error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedErr), helpers_test.ErrorToVal(err))
			}

			compareOrder(t, ord, tc.expectedOrd)
		})
	}
}
