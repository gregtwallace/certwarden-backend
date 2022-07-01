package acme

import (
	"encoding/json"
	"errors"
	"strconv"
)

// LE error response
type AcmeErrorResponse struct {
	Status int    `json:"status"`
	Type   string `json:"type"`
	Detail string `json:"detail"`
}

// Error() returns an error type for the acme error response
func (acmeError *AcmeErrorResponse) Error() error {
	status := strconv.Itoa(acmeError.Status)

	return errors.New("response error: status: " + status + ", type: " + acmeError.Type + ", detail: " + acmeError.Detail)
}

// unmarshalErrorResponse attempts to unmarshal into the error response object
// Note: This function returns err when an error response COULD NOT be decoded.
// That is, the function returns an error type when the response did NOT decode
// to an error.
func unmarshalErrorResponse(bodyBytes []byte) (response AcmeErrorResponse, err error) {
	err = json.Unmarshal(bodyBytes, &response)
	// if error decoding was not succesful
	if err != nil {
		return AcmeErrorResponse{}, err
	}

	// if we did get an error response from ACME
	return response, nil
}
