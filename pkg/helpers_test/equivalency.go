package helpers_test

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrTimePointerNotEqual = errors.New("time not equal")
)

// TimePointerEquals checks if the two underlying time values are equivelant
func TimePointerEquals(tim, expectedTim *time.Time) error {
	// nil checks
	if tim == nil && expectedTim == nil {
		return nil
	}
	if tim == nil && expectedTim != nil {
		return fmt.Errorf("%w: tim is nil but expectedTim is non-nil", ErrTimePointerNotEqual)
	}
	if tim != nil && expectedTim == nil {
		return fmt.Errorf("%w: tim is non-nil but expectedTim is nil", ErrTimePointerNotEqual)
	}

	// both are set - check value
	if !tim.Equal(*expectedTim) {
		return fmt.Errorf("%w: expected '%s' but got '%s'", ErrTimePointerNotEqual, expectedTim.UTC(), tim.UTC())
	}

	return nil
}

// StringPointerEquals checks if the two underlying string values are equivelant
func StringPointerEquals(s, expectedS *string) error {
	// nil checks
	if s == nil && expectedS == nil {
		return nil
	}
	if s == nil && expectedS != nil {
		return fmt.Errorf("%w: tim is nil but expectedTim is non-nil", ErrTimePointerNotEqual)
	}
	if s != nil && expectedS == nil {
		return fmt.Errorf("%w: tim is non-nil but expectedTim is nil", ErrTimePointerNotEqual)
	}

	// both are set - check value
	if *s != *expectedS {
		return fmt.Errorf("%w: expected '%s' but got '%s'", ErrTimePointerNotEqual, *expectedS, *s)
	}

	return nil
}
