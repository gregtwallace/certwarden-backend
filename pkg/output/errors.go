package output

import (
	"fmt"
)

// JsonError is the standardized error structure, it is the same as a regular message but also
// implements the Error() interface
type JsonError JsonResponse

// HttpStatusCode() implements the jsonData interface
func (je *JsonError) HttpStatusCode() int {
	return je.StatusCode
}

// Error() implements the error interface
func (je JsonError) Error() string {
	return fmt.Sprintf("%d: %s", je.StatusCode, je.Message)
}

//
// Funcs to make various Output Errors
//

// generic
func ErrorJsonErrNotFound(err error) *JsonError {
	return &JsonError{
		StatusCode: 404,
		Message:    fmt.Sprintf("error: not found (%s)", err),
	}
}

func ErrorJsonErrInternal(err error) *JsonError {
	return &JsonError{
		StatusCode: 500,
		Message:    fmt.Sprintf("error: internal (%s)", err),
	}
}

var ErrJsonUnauthorized = &JsonError{StatusCode: 401, Message: "unauthorized"}

// storage
func ErrorJsonErrStorageGeneric(err error) *JsonError {
	return &JsonError{
		StatusCode: 500,
		Message:    fmt.Sprintf("error: storage error (%s)", err),
	}
}

func ErrorJsonErrDeleteInUse(recordType string) *JsonError {
	return &JsonError{
		StatusCode: 409,
		Message:    fmt.Sprintf("error: record (%s) in use, can't delete", recordType),
	}
}

// write
func ErrorJsonErrWriteJsonError(err error) *JsonError {
	return &JsonError{
		StatusCode: 500,
		Message:    fmt.Sprintf("error: json response write failed (%s)", err),
	}
}

// validation
func ErrorJsonErrValidationFailed(err error) *JsonError {
	return &JsonError{
		StatusCode: 400,
		Message:    fmt.Sprintf("error: request validation (param or payload) invalid (%s)", err),
	}
}
