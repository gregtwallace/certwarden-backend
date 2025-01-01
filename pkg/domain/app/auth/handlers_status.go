package auth

import (
	"certwarden-backend/pkg/output"
	"net/http"
)

// statusResponse is used to provide information about the authentication process so the
// frontend / client can determine appropriate login method(s)
type statusResponse struct {
	output.JsonResponse
	AuthorizationStatus struct {
		Local struct {
			Enabled bool `json:"enabled"`
		} `json:"local"`
		OIDC struct {
			Enabled bool `json:"enabled"`
		} `json:"oidc"`
	} `json:"auth_status"`
}

// Status returns a response indicating info about Auth status. The route is NOT secure and
// therefore should NOT leak senstive information.
func (service *Service) Status(w http.ResponseWriter, r *http.Request) *output.JsonError {
	// return response to client
	response := &statusResponse{}
	response.StatusCode = http.StatusOK
	response.Message = "ok"
	// TODO: Actually check what's available
	response.AuthorizationStatus.Local.Enabled = service.methodLocalEnabled()
	response.AuthorizationStatus.OIDC.Enabled = service.methodOIDCEnabled()

	// write response
	err := service.output.WriteJSON(w, response)
	if err != nil {
		service.logger.Errorf("failed to write json (%s)", err)
		// err not detailed as this route will not be secured
		return output.JsonErrWriteJsonError(nil)
	}

	return nil
}
