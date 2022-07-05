package utils

import (
	"encoding/json"
	"net/http"
)

// Response backend sends in response to PUT/POST
type JsonResp struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

func WrapJSON(data interface{}, wrap string) map[string]interface{} {
	wrapper := make(map[string]interface{})
	wrapper[wrap] = data

	return wrapper
}

func WriteJSON(w http.ResponseWriter, status int, data interface{}, wrap string) error {
	wrappedData := WrapJSON(data, wrap)

	// TO-DO: Replace with regular Marshal (and/or add logic for dev vs. prod)
	json, err := json.MarshalIndent(wrappedData, "", "\t")
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(json)

	return nil
}

type jsonError struct {
	Message string `json:"message"`
}

func WriteErrorJSON(w http.ResponseWriter, err error) {
	currentError := jsonError{
		Message: err.Error(),
	}

	WriteJSON(w, http.StatusBadRequest, currentError, "error")
}

// WriteErrorStatusJSON writes an error with the specified status and error
func WriteErrorStatusJSON(w http.ResponseWriter, status int, err error) {
	currentError := jsonError{
		Message: err.Error(),
	}

	WriteJSON(w, status, currentError, "error")
}
