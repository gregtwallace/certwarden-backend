package output

import "fmt"

var (
	// generic
	ErrBadRequest      = Error{Status: 400, Message: "bad request"}
	ErrNotFound        = Error{Status: 404, Message: "not found"}
	ErrInternal        = Error{Status: 500, Message: "internal error"}
	ErrUnauthorized    = Error{Status: 401, Message: "unauthorized"}
	ErrUnavailableHttp = Error{Status: 503, Message: "server requires upgrade to https (not available over http)"}

	// storage errors
	ErrStorageGeneric = Error{Status: 500, Message: "storage error"}
	ErrDeleteInUse    = Error{Status: 409, Message: "record in use, can't delete"}

	// write
	errWriteJsonError = Error{Status: 500, Message: "json response write error"}
	//errWriteZipError  = Error{Status: 500, Message: "zip write error"}
	//errWritePemError  = Error{Status: 500, Message: "pem write error"}

	// validation
	ErrValidationFailed = Error{Status: 400, Message: "request validation (param or payload) invalid"}
	ErrBadDirectoryURL  = Error{Status: 400, Message: "specified acme directory url is not https or did not return a valid directory json response"}

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
