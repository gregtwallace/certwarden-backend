package test_helpers_test

import (
	"certwarden-backend/pkg/test_helpers"
	"errors"
	"testing"
)

func someFunctionA() error {
	return nil
}

func TestGetFunctionName(t *testing.T) {
	fName := test_helpers.GetFunctionName(someFunctionA)
	if fName != "someFunctionA" {
		t.Errorf("getfunctionname expected 'someFunctionA', but got '%s'", fName)
	}

	fName = test_helpers.GetFunctionName(test_helpers.ErrorToVal)
	if fName != "ErrorToVal" {
		t.Errorf("getfunctionname expected 'ErrorToVal', but got '%s'", fName)
	}

	// nil input
	var f *func() error
	f = nil
	fName = test_helpers.GetFunctionName(f)
	if fName != "<nil>" {
		t.Errorf("getfunctionname expected '<nil>', but got '%s'", fName)
	}
}

func TestStringPointerToVal(t *testing.T) {
	s := "test-1"
	result := test_helpers.StringPointerToVal(&s)
	if result != "test-1" {
		t.Errorf("stringpointertoval expected 'test-1', but got '%s'", result)
	}

	s = "some other test, again"
	result = test_helpers.StringPointerToVal(&s)
	if result != "some other test, again" {
		t.Errorf("stringpointertoval expected 'some other test, again', but got '%s'", result)
	}

	// nil
	var ptr *string
	result = test_helpers.StringPointerToVal(ptr)
	if result != "<nil>" {
		t.Errorf("stringpointertoval expected '<nil>', but got '%s'", result)
	}
}

func TestErrorToVal(t *testing.T) {
	e := errors.New("test-2")
	result := test_helpers.ErrorToVal(e)
	if result != "test-2" {
		t.Errorf("errortoval expected 'test-2', but got '%s'", result)
	}

	e = errors.New("some other test 2, again")
	result = test_helpers.ErrorToVal(e)
	if result != "some other test 2, again" {
		t.Errorf("errortoval expected 'some other test 2, again', but got '%s'", result)
	}

	// nil
	e = nil
	result = test_helpers.ErrorToVal(e)
	if result != "<nil>" {
		t.Errorf("errortoval expected '<nil>', but got '%s'", result)
	}
}
