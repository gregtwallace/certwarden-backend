package helpers_test

import (
	"errors"
	"strings"
)

// testErrorStringComp is a special error type to check error text for a matching
// value (as opposed to a strict type match); this is useful when the exact type
// is not importable
type testErrorStringComp struct {
	Inner error
}

func (e testErrorStringComp) Error() string {
	return e.Inner.Error()
}

func (e testErrorStringComp) Unwrap() error {
	return e.Inner
}

// NewTestErrorStringComp wraps the provided error text in a special error type that
// will be parsed and compared when the custom ErrorsIs is called
func NewTestErrorStringComp(errText string) testErrorStringComp {
	return testErrorStringComp{Inner: errors.New(errText)}
}

// ErrorsIs
func ErrorsIs(err error, target error) bool {
	// check if target is our special error type and if not, just do a normal errors.Is()
	tError, isTestErrStringCmp := errors.AsType[testErrorStringComp](target)
	if !isTestErrStringCmp {
		return errors.Is(err, target)
	}

	// if one is nil but not the other, they are not the same, return false early
	// to avoid calls to nil value
	if (err == nil && target != nil) ||
		(err != nil && target == nil) {
		return false
	}

	// comparison is case-insensitive
	return strings.Contains(
		strings.ToLower(err.Error()),
		strings.ToLower(tError.Unwrap().Error()),
	)
}
