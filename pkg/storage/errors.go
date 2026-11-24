package storage

import (
	"errors"
	"fmt"
)

// errors in generic storage package so there are no dependencies on sql or
// sql error types

var (
	ErrInUse                 = errors.New("record in use")
	ErrWrongRowCount         = errors.New("got wrong row count")
	ErrWrongAffectedRowCount = errors.New("wrong affected row count")
)

func errorWrongRowCount(expectedCount, gotCount int64) error {
	return fmt.Errorf("%w (expected: '%d', got: '%d')", ErrWrongRowCount, expectedCount, gotCount)
}

func errorWrongAffectedRowCount(expected, got int64) error {
	return fmt.Errorf("%w (expected: '%d', got: '%d')", ErrWrongAffectedRowCount, expected, got)
}
