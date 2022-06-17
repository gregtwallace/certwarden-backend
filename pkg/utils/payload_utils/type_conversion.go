package payload_utils

import "strconv"

// StringToInt turns an html string payload value into an int
// it uses pointers so a null pointer can be passed if the key
// was not part of the payload.
func StringToInt(s *string) (*int, error) {
	i := new(int)

	// i is a nil pointer if s is a nil pointer
	if s == nil {
		i = nil
	} else {
		// else try to convert s into an int and return the pointer to that
		var err error

		*i, err = strconv.Atoi(*s)
		if err != nil {
			return i, err
		}
	}

	return i, nil
}

// StringToBool turns an html string payload value into a bool
// If the string is null or anything other than true, the bool
// is set to false.
func StringToBool(s *string) *bool {
	b := new(bool)

	// b is a nil pointer if s is a nil pointer
	if s == nil {
		b = nil
	} else {
		// if s isn't nil "true" is true, everything else is false
		if *s == "true" {
			*b = true
		} else {
			*b = false
		}
	}

	return b
}
