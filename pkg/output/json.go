package output

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap/zapcore"
)

// JsonResponse is the standard response to clients
type JsonResponse struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
}

func (jr *JsonResponse) HttpStatusCode() int {
	return jr.StatusCode
}

type jsonData interface {
	HttpStatusCode() int
}

// WriteJSON marshalls data and then writes it to the ResponseWriter. An error is
// returned if writing failed.
func (service *Service) WriteJSON(w http.ResponseWriter, data jsonData) error {
	// marshalling: make it pretty if doing debug logging
	var jsonBytes []byte
	var err error
	if service.logger.Level() == zapcore.DebugLevel {
		jsonBytes, err = json.MarshalIndent(data, "", "\t")
	} else {
		jsonBytes, err = json.Marshal(data)
	}
	if err != nil {
		service.logger.Errorf("error marshalling json (%s)", err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(data.HttpStatusCode())

	_, err = w.Write(jsonBytes)
	if err != nil {
		service.logger.Errorf("error writing json (%s)", err)
		return err
	}

	return nil
}
