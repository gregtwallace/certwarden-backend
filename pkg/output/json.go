package output

import (
	"encoding/json"
	"legocerthub-backend/pkg/acme"
	"net/http"
)

// JsonResponse is the standard response to clients
type JsonResponse struct {
	Status  int    `json:"status"`
	Type    string `json:"type,omitempty"`
	Message any    `json:"message"`
}

// WrapJSON wraps data in the specified wrapper
func WrapJSON(data interface{}, wrap string) map[string]interface{} {
	wrapper := make(map[string]interface{})
	wrapper[wrap] = data

	return wrapper
}

// WriteJSON wraps data in the specified wrap, marshals it to json, and then writes the
// json to the ResponseWriter with the specified status code. The string of json that was
// written is returned, or an error if writing failed.
func WriteJSON(w http.ResponseWriter, status int, data interface{}, wrap string) (jsonWritten string, err error) {
	wrappedData := WrapJSON(data, wrap)

	// TO-DO: Replace with regular Marshal (and/or add logic for dev vs. prod)
	jsonBytes, err := json.MarshalIndent(wrappedData, "", "\t")
	// jsonBytes, err := json.Marshal(wrappedData)
	if err != nil {
		return "", err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(jsonBytes)

	return string(jsonBytes), nil
}

// WriteErrorJSON marshals the specified error and then writes it to the ResponseWriter.
// There are some special error types that are addressed on a case by case basis, otherwise
// there is a generic format.
// The string of json that was written is returned, or an error if writing failed.
func WriteErrorJSON(w http.ResponseWriter, err error) (jsonWritten string, writeErr error) {
	// special cases for specific error structs
	switch err := err.(type) {
	// output Error
	case Error:
		return WriteJSON(w, err.Status, err, "error")
	// ACME error repsonse (from ACME upstream server)
	case acme.AcmeErrorResponse:
		errToWrite := Error{
			Status:  err.Status,
			Type:    err.Type,
			Message: err.Detail,
		}

		return WriteJSON(w, err.Status, errToWrite, "error")
	default:
		// break to generic
	}

	// generic error
	currentError := Error{
		Status:  http.StatusBadRequest,
		Message: err.Error(),
	}

	return WriteJSON(w, currentError.Status, currentError, "error")
}
