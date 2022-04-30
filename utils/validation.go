package utils

import (
	"errors"
	"regexp"
)

// IsNameValid returns an error if not valid, nil if valid
// to be valid: must only contain symbols - _ . ~ letters and numbers
// name is also not allowed to be blank (len <= 0)
func IsNameValid(name string) error {
	regex, err := regexp.Compile("[^-_.~A-z0-9]|[\\^]")
	if err != nil {
		return err
	}

	invalid := regex.Match([]byte(name))
	if invalid || len(name) <= 0 {
		return errors.New("invalid name")
	}
	return nil
}
