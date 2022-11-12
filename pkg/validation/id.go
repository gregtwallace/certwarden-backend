package validation

import "errors"

const newId = -1

var (
	ErrIdBad      = errors.New("bad id")
	ErrIdMismatch = errors.New("id mismatch")
)

// IsIdNew returns true if the id is the new id value
func IsIdNew(id int) bool {
	return id == newId
}

// IsIdExistingValidRange returns true if the id is greater than or equal
// to 0 and is not the newId.
func IsIdExistingValidRange(id int) bool {
	// check if id is the new id or
	// check if id is not >= 0
	if id == newId || !(id >= 0) {
		return false
	}

	return true
}
