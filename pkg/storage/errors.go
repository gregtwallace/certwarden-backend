package storage

import (
	"errors"
	"fmt"
)

// errors in generic storage package so there are no dependencies on sql or
// sql error types

var (
	ErrInUse               = errors.New("record in use")
	ErrWrongUpdateRowCount = errors.New("wrong record update row count")
)

func errorWrongUpdateRowCount(expected int64, got int64) error {
	return fmt.Errorf("%w (expected: '%d', got: '%d')", ErrWrongUpdateRowCount, expected, got)
}
