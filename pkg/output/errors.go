package output

import "fmt"

var (
	// generic
	ErrBadRequest   = &Error{StatusCode: 400, Message: "error: bad request"}
	ErrNotFound     = &Error{StatusCode: 404, Message: "error: not found"}
	ErrInternal     = &Error{StatusCode: 500, Message: "error: internal error"}
	ErrUnauthorized = &Error{StatusCode: 401, Message: "error: unauthorized"}

	// storage errors
	ErrStorageGeneric = &Error{StatusCode: 500, Message: "error: storage error"}
	ErrDeleteInUse    = &Error{StatusCode: 409, Message: "error: record in use, can't delete"}

	// write
	ErrWriteConfigFailed = &Error{StatusCode: 500, Message: "error: failed to write lego config file"}
	ErrWriteJsonError    = &Error{StatusCode: 500, Message: "error: json response write error"}

	// validation
	ErrValidationFailed = &Error{StatusCode: 400, Message: "error: request validation (param or payload) invalid"}
	ErrBadDirectoryURL  = &Error{StatusCode: 400, Message: "error: specified acme directory url is not https or did not return a valid directory json response"}

	// order
	ErrOrderInvalid = &Error{StatusCode: 400, Message: "error: order status is invalid (which cannot be recovered from)"}
)

// Error is the standardized error structure, it is the same as a regular message but also
// implements Error() interface
type Error JsonResponse

// HttpStatusCode() implements the jsonData interface
func (e *Error) HttpStatusCode() int {
	return e.StatusCode
}

// Error() implements the error interface
func (e Error) Error() string {
	return fmt.Sprintf("%d: %s", e.StatusCode, e.Message)
}
