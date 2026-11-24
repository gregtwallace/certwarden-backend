package validation

const newID = -1

// IsNewID returns true if the id is the new id value
func IsNewID(id int) bool {
	return id == newID
}

// IsValidIDValue returns true if the ID is a usable value (i.e., it is greater
// than or equal to 0 and it is not the newID)
func IsValidIDValue(id int) bool {
	return id != newID && id >= 0
}
