package storage

import (
	"encoding/json"
)

// sliceToJsonString marshals the slice s into a json string. If nullable is false
// and s is nil, an empty array `[]` is returned instead of nil.
func sliceToJsonString[S []E, E any](s S, nullable bool) (*string, error) {
	val, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}

	// nullable?
	if nullable && s == nil {
		return nil, nil
	}

	// obj is empty and can't be nil
	if len(s) == 0 {
		return new("[]"), nil
	}

	// non-empty slice
	return new(string(val)), nil
}

// The below two methods are required due to the may Marshal/Unmarshal interfaces
// deal with nil values (i.e., they return a 0 length byte array instead of nil)

// structToNullableJsonString marshals the struct o into a json string. If val is nil,
// nil is returned instead of an 0 length byte slice.
func structToNullableJsonString[E any](val *E) (*string, error) {
	if val == nil {
		return nil, nil
	}

	res, err := json.Marshal(val)
	if err != nil {
		return nil, err
	}

	// non-nil obj
	return new(string(res)), nil
}

// jsonStringToNullableStruct unmarshals the string s into the object pointer E. If s
// is nil, nil is returned
func jsonStringToNullableStruct[E any](s *string) (*E, error) {
	if s == nil {
		return nil, nil
	}

	res := new(E)
	err := json.Unmarshal([]byte(*s), res)

	return res, err
}
