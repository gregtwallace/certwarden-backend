package output

import (
	"encoding/json"
	"legocerthub-backend/pkg/acme"
	"net/http"
)

// WriteJSON wraps data in the specified wrap, marshals it to json, and then writes the
// json to the ResponseWriter with the specified status code. Error is returned if the
// data can't be marshaled
func WriteJSON(w http.ResponseWriter, status int, data interface{}, wrap string) error {
	wrapper := make(map[string]interface{})
	wrapper[wrap] = data

	// TO-DO: Replace with regular Marshal (and/or add logic for dev vs. prod)
	json, err := json.MarshalIndent(wrapper, "", "\t")
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(json)

	return nil
}

// jsonError is the standardized error structure
type jsonError struct {
	Status  int    `json:"status"`
	Type    string `json:"type,omitempty"`
	Message string `json:"message"`
}

// WriteErrorJSON marshals the specified error and then writes it to the ResponseWriter
// there are some special error types that are addressed on a case by case basis, otherwise
// there is a generic format. Error is returned if the data can't be marshaled (which should
// never happen)
func WriteErrorJSON(w http.ResponseWriter, err error) error {
	// special cases for specific error structs
	switch err := err.(type) {
	// ACME error repsonse (from ACME upstream server)
	case acme.AcmeErrorResponse:
		errToWrite := jsonError{
			Status:  err.Status,
			Type:    err.Type,
			Message: err.Detail,
		}

		return WriteJSON(w, err.Status, errToWrite, "error")
	default:
		// break to generic
	}

	// generic error
	currentError := jsonError{
		Status:  http.StatusBadRequest,
		Message: err.Error(),
	}

	return WriteJSON(w, currentError.Status, currentError, "error")
}
