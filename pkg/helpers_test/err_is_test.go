package helpers_test_test

import (
	"certwarden-backend/pkg/acme"
	"certwarden-backend/pkg/helpers_test"
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
			target:    helpers_test.MakeTestErrorStringComp("an error"),
			isTheSame: false,
		},
		{
			err:       helpers_test.MakeTestErrorStringComp("an error"),
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
			target:    helpers_test.MakeTestErrorStringComp("an error"),
			isTheSame: false,
		},
		{
			err:       acme.ErrChallengeMalformed,
			target:    helpers_test.MakeTestErrorStringComp("an error"),
			isTheSame: false,
		},
		{
			err:       errors.New("some error"),
			target:    helpers_test.MakeTestErrorStringComp("uh oh, some error"),
			isTheSame: false,
		},
		{
			err:       errors.New("some error"),
			target:    helpers_test.MakeTestErrorStringComp("some error, uh oh"),
			isTheSame: false,
		},
		{
			err:       helpers_test.MakeTestErrorStringComp("uh oh, some error"),
			target:    errors.New("some error"),
			isTheSame: false,
		},
		{
			err:       helpers_test.MakeTestErrorStringComp("some error, uh oh"),
			target:    errors.New("some error"),
			isTheSame: false,
		},
		{
			err:       errors.New("uh oh, some error"),
			target:    helpers_test.MakeTestErrorStringComp("some error"),
			isTheSame: true,
		},
		{
			err:       errors.New("some error, uh oh"),
			target:    helpers_test.MakeTestErrorStringComp("some error"),
			isTheSame: true,
		},
		{
			err:       errors.New("uh oh, some error"),
			target:    helpers_test.MakeTestErrorStringComp("SOME errOR"),
			isTheSame: true,
		},
		{
			err:       errors.New("some error, uh oh"),
			target:    helpers_test.MakeTestErrorStringComp("SOME errOR"),
			isTheSame: true,
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("#%d:", i), func(t *testing.T) {
			res := helpers_test.ErrorsIs(tc.err, tc.target)
			if res != tc.isTheSame {
				t.Errorf("err '%s' with target '%s' expected '%t' but got '%t'", helpers_test.ErrorToVal(tc.err), helpers_test.ErrorToVal(tc.target), tc.isTheSame, res)
			}
		})
	}
}
