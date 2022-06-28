package acme

import (
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
