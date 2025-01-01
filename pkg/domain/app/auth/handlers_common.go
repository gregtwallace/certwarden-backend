package auth

import (
	"certwarden-backend/pkg/domain/app/auth/session_manager"
	"certwarden-backend/pkg/output"
	"fmt"
	"net/http"
)

// RefreshUsingCookie validates the SessionToken cookie and confirms its UUID is for a valid
// session. If so, it generates a new AccessToken and new SessionToken cookie and then sends both
// to the client.
func (service *Service) RefreshUsingCookie(w http.ResponseWriter, r *http.Request) *output.JsonError {
	service.logger.Infof("client %s: attempting session refresh", r.RemoteAddr)

	username, auth, outErr := service.sessionManager.RefreshSession(r, w)
	if outErr != nil {
		return outErr
	}

	// return response to client
	response := &session_manager.AuthResponse{}
	response.StatusCode = http.StatusOK
	response.Message = fmt.Sprintf("user '%s' session refreshed", username)
	response.Authorization = auth

	// write response
	auth.WriteSessionCookie(w)
	err := service.output.WriteJSON(w, response)
	if err != nil {
		service.logger.Errorf("failed to write json (%s)", err)
		// detailed error is OK here because the user passed auth checks
		return output.JsonErrWriteJsonError(err)
	}

	// log success
	service.logger.Infof("client %s: session refresh for user '%s' succeeded", r.RemoteAddr, username)

	return nil
}

// Logout logs the client out and removes the session from session manager
func (service *Service) Logout(w http.ResponseWriter, r *http.Request) *output.JsonError {
	// log attempt
	service.logger.Infof("client %s: attempting logout", r.RemoteAddr)

	// immediately delete any session cookie
	service.sessionManager.DeleteSessionCookie(w)

	// delete session matching the access token
	username, err := service.sessionManager.DeleteSession(r)
	if err != nil {
		service.logger.Errorf("client %s: logout failed (%s)", r.RemoteAddr, err)
		return output.JsonErrUnauthorized
	}

	// log success
	service.logger.Infof("client %s: logout for user '%s' succeeded", r.RemoteAddr, username)

	// return response (logged out)
	response := &output.JsonResponse{}
	response.StatusCode = http.StatusOK
	response.Message = fmt.Sprintf("user '%s' logged out", username)

	err = service.output.WriteJSON(w, response)
	if err != nil {
		service.logger.Errorf("failed to write json (%s)", err)
		// detailed error is OK here because the user passed auth checks
		return output.JsonErrWriteJsonError(err)
	}

	return nil
}
