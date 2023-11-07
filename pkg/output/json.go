package output

import (
	"encoding/json"
	"legocerthub-backend/pkg/acme"
	"net/http"

	"go.uber.org/zap/zapcore"
)

// JsonResponse is the standard response to clients
type JsonResponse struct {
	Status  int    `json:"status"`
	Type    string `json:"type,omitempty"`
	ID      int    `json:"record_id,omitempty"`
	Message string `json:"message"`
}

// wrapJSON wraps data in the specified wrapper
func wrapJSON(data interface{}, wrap string) map[string]interface{} {
	wrapper := make(map[string]interface{})
	wrapper[wrap] = data

	return wrapper
}

// WriteJSON wraps data in the specified wrap, marshals it to json, and then writes the
// json to the ResponseWriter with the specified status code. An error is returned if
// writing failed.
func (service *Service) WriteJSON(w http.ResponseWriter, status int, data interface{}, wrap string) error {
	var jsonBytes []byte
	var err error

	// wrap the data
	wrappedData := wrapJSON(data, wrap)

	// make it pretty if doing debug logging
	if service.logger.Level() == zapcore.DebugLevel {
		jsonBytes, err = json.MarshalIndent(wrappedData, "", "\t")
	} else {
		jsonBytes, err = json.Marshal(wrappedData)
	}
	if err != nil {
		service.logger.Errorf("error marshalling json (%s)", err)
		return errWriteJsonError
	}

	return service.WriteMarshalledJSON(w, status, jsonBytes)
}

// WriteMarshalledJSON assumes the data is already marshalled correctly. It just writes the json
// to the ResponseWriter with the specified status code.
// An error is returned if writing failed.
func (service *Service) WriteMarshalledJSON(w http.ResponseWriter, status int, marshalledData []byte) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_, err := w.Write(marshalledData)
	if err != nil {
		service.logger.Errorf("error writing json (%s)", err)
		return errWriteJsonError
	}

	return nil
}

// WriteErrorJSON marshals the specified error and then writes it to the ResponseWriter.
// There are some special error types that are addressed on a case by case basis, otherwise
// there is a generic format.
// An error is returned if writing failed.
func (service *Service) WriteErrorJSON(w http.ResponseWriter, err error) error {
	// special cases for specific error structs
	switch err := err.(type) {
	// output Error
	case Error:
		return service.WriteJSON(w, err.Status, err, "error")
	// ACME error repsonse (from ACME upstream server)
	case acme.Error:
		errToWrite := Error{
			Status:  err.Status,
			Type:    err.Type,
			Message: err.Detail,
		}

		return service.WriteJSON(w, err.Status, errToWrite, "error")
	default:
		// break to generic
	}

	// generic error
	currentError := Error{
		Status:  http.StatusBadRequest,
		Message: err.Error(),
	}

	return service.WriteJSON(w, currentError.Status, currentError, "error")
}
