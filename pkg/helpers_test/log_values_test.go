package helpers_test_test

import (
	"certwarden-backend/pkg/helpers_test"
	"errors"
	"testing"
	"time"
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

func TestIntPointerToVal(t *testing.T) {
	i := 15
	result := helpers_test.IntPointerToVal(&i)
	if result != "15" {
		t.Errorf("intpointertoval expected '15', but got '%s'", result)
	}

	i = -15
	result = helpers_test.IntPointerToVal(&i)
	if result != "-15" {
		t.Errorf("intpointertoval expected '-15', but got '%s'", result)
	}

	i = 89342342034920234
	result = helpers_test.IntPointerToVal(&i)
	if result != "89342342034920234" {
		t.Errorf("intpointertoval expected '89342342034920234', but got '%s'", result)
	}

	// nil
	var in *int
	result = helpers_test.IntPointerToVal(in)
	if result != "<nil>" {
		t.Errorf("intpointertoval expected '<nil>', but got '%s'", result)
	}
}

// custom error type for testing concrete nil pointer (reflection nil check)
type nilError struct{ text string }

func (ne *nilError) Error() string { return ne.text }

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

	// nil - concrete error type pointer
	var ne *nilError
	result = helpers_test.ErrorToVal(ne)
	if result != "<nil>" {
		t.Errorf("errortoval expected '<nil>', but got '%s'", result)
	}
}

func TestTimePointerToVal(t *testing.T) {
	ti := time.Unix(1787150875, 0)
	result := helpers_test.TimeToVal(&ti)
	if result != "2026-08-19T14:47:55Z" {
		t.Errorf("timepointertoval expected '2026-08-19T14:47:55Z', but got '%s'", result)
	}

	ti = time.Unix(-55, 0)
	result = helpers_test.TimeToVal(&ti)
	if result != "1969-12-31T23:59:05Z" {
		t.Errorf("timepointertoval expected '1969-12-31T23:59:05Z', but got '%s'", result)
	}

	ti = time.Unix(1785150875, 0)
	result = helpers_test.TimeToVal(&ti)
	if result != "2026-07-27T11:14:35Z" {
		t.Errorf("timepointertoval expected '2026-07-27T11:14:35Z', but got '%s'", result)
	}

	// nil
	var tn *time.Time
	result = helpers_test.TimeToVal(tn)
	if result != "<nil>" {
		t.Errorf("timepointertoval expected '<nil>', but got '%s'", result)
	}
}
