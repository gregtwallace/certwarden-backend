package test_helpers_test

import (
	"certwarden-backend/pkg/acme"
	"certwarden-backend/pkg/test_helpers"
	"database/sql"
	"fmt"
	"testing"
)

func TestErrorsIs(t *testing.T) {
	testCases := []struct {
		err            error
		target         error
		expectedResult bool
	}{
		{
			err:            nil,
			target:         nil,
			expectedResult: true,
		},
		{
			err:            nil,
			target:         test_helpers.ErrAnyType,
			expectedResult: false,
		},
		{
			err:            test_helpers.ErrAnyType,
			target:         nil,
			expectedResult: false,
		},
		{
			err:            sql.ErrNoRows,
			target:         acme.ErrChallengeMalformed,
			expectedResult: false,
		},
		{
			err:            sql.ErrNoRows,
			target:         test_helpers.ErrAnyType,
			expectedResult: true,
		},
		{
			err:            acme.ErrChallengeMalformed,
			target:         test_helpers.ErrAnyType,
			expectedResult: true,
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("#%d:", i), func(t *testing.T) {
			res := test_helpers.ErrorsIs(tc.err, tc.target)
			if res != tc.expectedResult {
				t.Errorf("err '%s' with target '%s' expected '%t' but got '%t'", test_helpers.ErrorToVal(tc.err), test_helpers.ErrorToVal(tc.target), tc.expectedResult, res)
			}
		})
	}
}
