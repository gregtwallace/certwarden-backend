package utils

import (
	"encoding/json"
	"net/http"
)

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

func WriteErrorJSON(w http.ResponseWriter, err error) {
	type jsonError struct {
		Message string `json:"message"`
	}

	currentError := jsonError{
		Message: err.Error(),
	}

	WriteJSON(w, http.StatusBadRequest, currentError, "error")
}