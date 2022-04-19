package main

import (
	"encoding/json"
	"net/http"
)

func (app *application) WriteJSON(w http.ResponseWriter, status int, data interface{}, wrap string) error {
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
