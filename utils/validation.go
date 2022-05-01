package utils

import (
	"errors"
	"regexp"
	"strconv"
)

// idIdMatch returns an error if the URI and payload don't match
// otherwise it returns nil
func isIdMatch(idParam string, idPayload string) error {
	if idParam != idPayload {
		return errors.New("id param/payload mismatch")
	}
	return nil
}

// IsIdNewValid returns an error if not valid, nil if valid
// -1 is the only valid new ID
func IsIdValidNew(idParam string, idPayload string) error {
	// check the payload and the URI match
	err := isIdMatch(idParam, idPayload)
	if err != nil {
		return err
	}

	idPayloadInt, err := strconv.Atoi(idPayload)
	if err != nil {
		return err
	}

	// check id equals new value of -1
	if idPayloadInt != -1 {
		return errors.New("invalid id for new")
	}
	return nil
}

// IsIdExistingValid returns an error if not valid, nil if valid
// we'll generally assume the id is valid if >= 0
func IsIdValidExisting(idParam string, idPayload string) error {
	// check the payload and the URI match
	err := isIdMatch(idParam, idPayload)
	if err != nil {
		return err
	}

	idPayloadInt, err := strconv.Atoi(idPayload)
	if err != nil {
		return err
	}

	// check id equals new value of -1
	if idPayloadInt < 0 {
		return errors.New("invalid id")
	}

	// check id is in the db TODO: Check if this is needed
	// _, err = privateKeysApp.dbGetOnePrivateKey(id)
	// if err != nil {
	// 	privateKeysApp.Logger.Printf("privatekeys: PutOne: invalid payload id -- err: %s", err)
	// 	utils.WriteErrorJSON(w, err)
	// 	return
	// }

	return nil
}

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
