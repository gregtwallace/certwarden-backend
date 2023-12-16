package orders

import (
	"legocerthub-backend/pkg/output"
	"net/http"
)

// GetAllWorkStatus returns all jobs/orders currently with fulfiller
func (service *Service) GetAllWorkStatus(w http.ResponseWriter, r *http.Request) *output.Error {
	err := service.output.WriteJSON(w, service.orderFulfiller.allWorkStatus())
	if err != nil {
		service.logger.Errorf("failed to write json (%s)", err)
		return output.ErrWriteJsonError
	}
	return nil
}
