package validation

const newId = -1

// IsIdNew returns true if the id is the new id value
func IsIdNew(id int) bool {
	return id == newId
}

// IsIdExistingValidRange returns true if the id is greater than or equal
// to 0 and is not the newId.
func IsIdExistingValidRange(id int) bool {
	return id != newId && id >= 0
}
