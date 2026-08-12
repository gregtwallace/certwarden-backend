package helpers_test_test

import (
	"certwarden-backend/pkg/helpers_test"
	"errors"
	"testing"
)

func someFunctionA() error {
	return nil
}

func TestGetFunctionName(t *testing.T) {
	fName := helpers_test.GetFunctionName(someFunctionA)
	if fName != "someFunctionA" {
		t.Errorf("getfunctionname expected 'someFunctionA', but got '%s'", fName)
	}

	fName = helpers_test.GetFunctionName(helpers_test.ErrorToVal)
	if fName != "ErrorToVal" {
		t.Errorf("getfunctionname expected 'ErrorToVal', but got '%s'", fName)
	}

	// nil input
	var f *func() error = nil
	fName = helpers_test.GetFunctionName(f)
	if fName != "<nil>" {
		t.Errorf("getfunctionname expected '<nil>', but got '%s'", fName)
	}
}

func TestStringPointerToVal(t *testing.T) {
	s := "test-1"
	result := helpers_test.StringPointerToVal(&s)
	if result != "test-1" {
		t.Errorf("stringpointertoval expected 'test-1', but got '%s'", result)
	}

	s = "some other test, again"
	result = helpers_test.StringPointerToVal(&s)
	if result != "some other test, again" {
		t.Errorf("stringpointertoval expected 'some other test, again', but got '%s'", result)
	}

	// nil
	var ptr *string
	result = helpers_test.StringPointerToVal(ptr)
	if result != "<nil>" {
		t.Errorf("stringpointertoval expected '<nil>', but got '%s'", result)
	}
}

func TestErrorToVal(t *testing.T) {
	e := errors.New("test-2")
	result := helpers_test.ErrorToVal(e)
	if result != "test-2" {
		t.Errorf("errortoval expected 'test-2', but got '%s'", result)
	}

	e = errors.New("some other test 2, again")
	result = helpers_test.ErrorToVal(e)
	if result != "some other test 2, again" {
		t.Errorf("errortoval expected 'some other test 2, again', but got '%s'", result)
	}

	// nil
	e = nil
	result = helpers_test.ErrorToVal(e)
	if result != "<nil>" {
		t.Errorf("errortoval expected '<nil>', but got '%s'", result)
	}
}
