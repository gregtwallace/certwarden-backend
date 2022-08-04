package acme

import (
	"encoding/json"
	"fmt"
)

// ACME error
type Error struct {
	Status int    `json:"status"`
	Type   string `json:"type"`
	Detail string `json:"detail"`
}

// Error() implements the error interface
func (e Error) Error() string {
	return fmt.Sprintf("status: %d; type: %s; detail: %s", e.Status, e.Type, e.Detail)
}

// unmarshalErrorResponse attempts to unmarshal into the error response object
// Note: This function returns err when an error response COULD NOT be decoded.
// That is, the function returns an error type when the response did NOT decode
// to an error.
func unmarshalErrorResponse(bodyBytes []byte) (response Error, err error) {
	err = json.Unmarshal(bodyBytes, &response)
	// if error decoding was not succesful
	if err != nil {
		return Error{}, err
	}

	// if we did get an error response from ACME
	return response, nil
}
