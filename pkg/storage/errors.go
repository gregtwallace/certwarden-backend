package storage

import "errors"

// errors in generic storage package so there are no dependencies on sql or
// sql error types

var (
	ErrInUse = errors.New("record in use")
)
