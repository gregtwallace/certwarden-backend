package orders

import (
	"legocerthub-backend/pkg/output"
	"net/http"
)

func (service *Service) SimulateAuto(w http.ResponseWriter, r *http.Request) error {
	err := service.orderExpiringCerts()
	if err != nil {
		service.logger.Error(err)
	}

	// return response to client
	response := output.JsonResponse{
		Status:  http.StatusOK,
		Message: "temp did a thing",
	}

	_, err = service.output.WriteJSON(w, response.Status, response, "response")
	if err != nil {
		service.logger.Error(err)
		return output.ErrWriteJsonFailed
	}

	return nil
}
