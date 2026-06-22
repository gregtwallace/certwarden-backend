package test_helpers_test

import (
	"certwarden-backend/pkg/acme"
	"certwarden-backend/pkg/test_helpers"
	"database/sql"
	"errors"
	"fmt"
	"testing"
)

func TestErrorsIs(t *testing.T) {
	testCases := []struct {
		err       error
		target    error
		isTheSame bool
	}{
		{
			err:       nil,
			target:    nil,
			isTheSame: true,
		},
		{
			err:       nil,
			target:    test_helpers.MakeTestErrorStringComp("an error"),
			isTheSame: false,
		},
		{
			err:       test_helpers.MakeTestErrorStringComp("an error"),
			target:    nil,
			isTheSame: false,
		},
		{
			err:       sql.ErrNoRows,
			target:    acme.ErrChallengeMalformed,
			isTheSame: false,
		},
		{
			err:       errors.New("an error 1"),
			target:    errors.New("another error 2"),
			isTheSame: false,
		},
		{
			err:       sql.ErrNoRows,
			target:    test_helpers.MakeTestErrorStringComp("an error"),
			isTheSame: false,
		},
		{
			err:       acme.ErrChallengeMalformed,
			target:    test_helpers.MakeTestErrorStringComp("an error"),
			isTheSame: false,
		},
		{
			err:       errors.New("some error"),
			target:    test_helpers.MakeTestErrorStringComp("uh oh, some error"),
			isTheSame: false,
		},
		{
			err:       errors.New("some error"),
			target:    test_helpers.MakeTestErrorStringComp("some error, uh oh"),
			isTheSame: false,
		},
		{
			err:       test_helpers.MakeTestErrorStringComp("uh oh, some error"),
			target:    errors.New("some error"),
			isTheSame: false,
		},
		{
			err:       test_helpers.MakeTestErrorStringComp("some error, uh oh"),
			target:    errors.New("some error"),
			isTheSame: false,
		},
		{
			err:       errors.New("uh oh, some error"),
			target:    test_helpers.MakeTestErrorStringComp("some error"),
			isTheSame: true,
		},
		{
			err:       errors.New("some error, uh oh"),
			target:    test_helpers.MakeTestErrorStringComp("some error"),
			isTheSame: true,
		},
		{
			err:       errors.New("uh oh, some error"),
			target:    test_helpers.MakeTestErrorStringComp("SOME errOR"),
			isTheSame: true,
		},
		{
			err:       errors.New("some error, uh oh"),
			target:    test_helpers.MakeTestErrorStringComp("SOME errOR"),
			isTheSame: true,
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("#%d:", i), func(t *testing.T) {
			res := test_helpers.ErrorsIs(tc.err, tc.target)
			if res != tc.isTheSame {
				t.Errorf("err '%s' with target '%s' expected '%t' but got '%t'", test_helpers.ErrorToVal(tc.err), test_helpers.ErrorToVal(tc.target), tc.isTheSame, res)
			}
		})
	}
}
