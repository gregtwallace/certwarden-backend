package test_helpers

import "errors"

var ErrAnyType = errors.New("error of any type")

// ErrorsIs is a custom implementation of errors.Is() to provide an error type
// that will match any error type; otherwise, it is just a wrapper for errors.Is();
// Only needs to be used when avoiding an import of the exact error type
func ErrorsIs(err error, target error) bool {
	if err != nil && errors.Is(target, ErrAnyType) {
		return true
	}

	return errors.Is(err, target)
}
