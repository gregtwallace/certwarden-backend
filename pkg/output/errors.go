package output

import "fmt"

var (
	// generic
	ErrNotFound     = Error{Status: 404, Message: "not found"}
	ErrInternal     = Error{Status: 500, Message: "internal error"}
	ErrUnauthorized = Error{Status: 401, Message: "unauthorized"}

	// storage errors
	ErrStorageGeneric = Error{Status: 500, Message: "storage error"}
	ErrDeleteInUse    = Error{Status: 409, Message: "record in use, can't delete"}

	// write
	ErrWriteJsonFailed = Error{Status: 500, Message: "json response write failed"}
	ErrWritePemFailed  = Error{Status: 500, Message: "pem write failed"}

	// validation
	ErrValidationFailed = Error{Status: 400, Message: "request validation (param or payload) invalid"}

	// order
	ErrOrderInvalid     = Error{Status: 400, Message: "order status is invalid (which cannot be recovered from)"}
	ErrOrderCantFulfill = Error{Status: 400, Message: "failed to order from acme (it is likely this order is already currently being processed)"}
)

// Error is the standardized error structure, it is the same as a regular message but also
// implements Error()
type Error JsonResponse

// Error() implements the error interface
func (e Error) Error() string {
	// if there is a type use it
	if e.Type != "" {
		return fmt.Sprintf("%d: %s (%s)", e.Status, e.Message, e.Type)
	}

	// else omit it
	return fmt.Sprintf("%d: %s", e.Status, e.Message)
}
