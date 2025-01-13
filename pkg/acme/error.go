package acme

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ACME error
type Error struct {
	Status int    `json:"status"`
	Type   string `json:"type"`
	Detail string `json:"detail"`
}

// Error() implements the error interface
func (e *Error) Error() string {
	return fmt.Sprintf("status: %d; type: %s; detail: %s", e.Status, e.Type, e.Detail)
}

// unmarshalErrorResponse attempts to unmarshal into the error response object. If
// it returns nil, the bodyBytes are not an ACME error.
func unmarshalErrorResponse(bodyBytes []byte) *Error {
	acmeErr := new(Error)
	err := json.Unmarshal(bodyBytes, acmeErr)
	// if error decoding was not succesful, not an error
	if err != nil {
		return nil
	}

	// validate the unmarshalled thing is an error, and not just something else that
	// unmarshalled without golang error
	if !strings.HasPrefix(acmeErr.Type, "urn:ietf:params:acme:error") {
		return nil
	}

	// if we did get an error response from ACME
	return acmeErr
}

// MarshalledString returns a JSON object as a string. This is useful to
// store an ACME error in storage (e.g. a database string)
func (e *Error) MarshalledString() (*string, error) {
	if e == nil {
		return nil, nil
	}

	errBytes, err := json.Marshal(*e)
	if err != nil {
		return nil, err
	}

	s := new(string)
	*s = string(errBytes)

	return s, nil
}

// NewAcmeError converts a json acme.Error object into an acme.Error
func NewAcmeError(acmeErrorJson *string) (e *Error) {
	if acmeErrorJson == nil {
		return nil
	}

	err := json.Unmarshal([]byte(*acmeErrorJson), e)
	if err != nil {
		// if unmarshal fails, return nil
		return nil
	}

	return e
}
