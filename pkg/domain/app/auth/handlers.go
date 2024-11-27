package auth

import (
	"certwarden-backend/pkg/output"
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

// authResponse contains the JSON response for both
// login and session (session token is in a cookie
// so the JSON struct doesn't change)
type authResponse struct {
	output.JsonResponse
	Authorization authorization `json:"authorization"`
}

// loginPayload is the payload client's send to login
type loginPayload struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginUsingUserPwPayload takes the loginPayload, looks up the username in storage
// and validates the password. If so, an Access Token is returned in JSON and a refresh
// token is sent in a cookie.
func (service *Service) LoginUsingUserPwPayload(w http.ResponseWriter, r *http.Request) *output.JsonError {
	// wrap handler to easily check err and delete cookies
	outErr := func() *output.JsonError {
		var payload loginPayload

		// log attempt
		service.logger.Infof("client %s: attempting login", r.RemoteAddr)

		// decode body into payload
		err := json.NewDecoder(r.Body).Decode(&payload)
		if err != nil {
			service.logger.Infof("client %s: login failed (payload error: %s)", r.RemoteAddr, err)
			return output.JsonErrUnauthorized
		}

		// fetch the password hash from storage
		user, err := service.storage.GetOneUserByName(payload.Username)
		if err != nil {
			service.logger.Infof("client %s: login failed (bad username: %s)", r.RemoteAddr, err)
			return output.JsonErrUnauthorized
		}

		// compare
		err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(payload.Password))
		if err != nil {
			service.logger.Infof("client %s: login failed (bad password: %s)", r.RemoteAddr, err)
			return output.JsonErrUnauthorized
		}

		// user and password now verified, make auth
		auth, err := service.newAuthorization(user.Username)
		if err != nil {
			service.logger.Errorf("client %s: login failed (internal error: %s)", r.RemoteAddr, err)
			return output.JsonErrInternal(nil)
		}

		// save auth's session in manager
		err = service.sessionManager.new(auth.SessionTokenClaims)
		if err != nil {
			service.logger.Errorf("client %s: login failed (internal error: %s)", r.RemoteAddr, err)
			return output.JsonErrUnauthorized
		}

		// return response to client
		response := &authResponse{}
		response.StatusCode = http.StatusOK
		response.Message = fmt.Sprintf("user '%s' logged in", auth.SessionTokenClaims.Subject)
		response.Authorization = auth

		// write response
		auth.writeSessionCookie(w)
		err = service.output.WriteJSON(w, response)
		if err != nil {
			service.logger.Errorf("failed to write json (%s)", err)
			// detailed error is OK here because the user passed auth checks
			return output.JsonErrWriteJsonError(err)
		}

		// log success
		service.logger.Infof("client %s: user '%s' logged in", r.RemoteAddr, auth.SessionTokenClaims.Subject)

		return nil
	}()

	// if err, delete session cookie and return err
	if outErr != nil {
		service.deleteSessionCookie(w)
		return outErr
	}

	return nil
}

// RefreshUsingCookie validates the SessionToken cookie and confirms its UUID is for a valid
// session. If so, it generates a new AccessToken and new SessionToken cookie and then sends both
// to the client.
func (service *Service) RefreshUsingCookie(w http.ResponseWriter, r *http.Request) *output.JsonError {
	// wrap to easily check err and delete cookies
	outErr := func() *output.JsonError {
		// log attempt
		service.logger.Infof("client %s: attempting access token refresh", r.RemoteAddr)

		// validate cookie
		oldClaims, outErr := service.validateSessionCookie(r, w, "access token refresh")
		if outErr != nil {
			// error logged in validateCookieSession func and nice output error returned
			return outErr
		}

		// cookie & session verified, make new auth
		auth, err := service.newAuthorization(oldClaims.Subject)
		if err != nil {
			service.logger.Errorf("client %s: access token refresh failed (internal error: %s)", r.RemoteAddr, err)
			return output.JsonErrInternal(nil)
		}

		// refresh session in manager (remove old, add new)
		err = service.sessionManager.refresh(*oldClaims, auth.SessionTokenClaims)
		if err != nil {
			service.logger.Errorf("client %s: access token refresh failed (internal error: %s)", r.RemoteAddr, err)
			return output.JsonErrUnauthorized
		}

		// return response (new auth) to client
		response := &authResponse{}
		response.StatusCode = http.StatusOK
		response.Message = fmt.Sprintf("user '%s' access token refreshed", auth.SessionTokenClaims.Subject)
		response.Authorization = auth

		// write response
		auth.writeSessionCookie(w)
		err = service.output.WriteJSON(w, response)
		if err != nil {
			service.logger.Errorf("failed to write json (%s)", err)
			// detailed error is OK here because the user passed auth checks
			return output.JsonErrWriteJsonError(err)
		}

		// log success
		service.logger.Infof("client %s: access token refresh for user '%s' succeeded", r.RemoteAddr, auth.SessionTokenClaims.Subject)

		return nil
	}()

	// if err, delete cookies and return err
	if outErr != nil {
		service.deleteSessionCookie(w)
		return outErr
	}

	return nil
}

// Logout logs the client out and removes the session from session manager
func (service *Service) Logout(w http.ResponseWriter, r *http.Request) *output.JsonError {
	// log attempt
	service.logger.Infof("client %s: attempting logout", r.RemoteAddr)

	// get claims from auth header
	oldClaims, err := service.ValidateAuthHeader(r, w, "logout")
	if err != nil {
		service.logger.Errorf("client %s: logout failed (%s)", r.RemoteAddr, oldClaims.Subject, err)
		return output.JsonErrUnauthorized
	}

	// remove session in manager
	err = service.sessionManager.close(*oldClaims)
	if err != nil {
		service.logger.Errorf("client %s: logout for user '%s' failed (%s)", r.RemoteAddr, oldClaims.Subject, err)
		return output.JsonErrUnauthorized
	}

	// log success
	service.logger.Infof("client %s: logout for user '%s' succeeded", r.RemoteAddr, oldClaims.Subject)

	// return response (logged out)
	response := &output.JsonResponse{}
	response.StatusCode = http.StatusOK
	response.Message = fmt.Sprintf("user '%s' logged out", oldClaims.Subject)
	// delete session cookie
	service.deleteSessionCookie(w)

	err = service.output.WriteJSON(w, response)
	if err != nil {
		service.logger.Errorf("failed to write json (%s)", err)
		// detailed error is OK here because the user passed auth checks
		return output.JsonErrWriteJsonError(err)
	}

	return nil
}

// passwordChangePayload contains the expected payload fields for
// a user changing their password
type passwordChangePayload struct {
	CurrentPassword    string `json:"current_password"`
	NewPassword        string `json:"new_password"`
	ConfirmNewPassword string `json:"confirm_new_password"`
}

// ChangePassword allows a user to change their password
func (service *Service) ChangePassword(w http.ResponseWriter, r *http.Request) *output.JsonError {
	// log attempt
	service.logger.Infof("client %s: attempting password change", r.RemoteAddr)

	// validate jwt and get the claims (to confirm the username)
	claims, err := service.ValidateAuthHeader(r, w, "password change")
	if err != nil {
		service.logger.Infof("client %s: password change failed (bad auth header: %s)", r.RemoteAddr, err)
		return output.JsonErrUnauthorized
	}
	username := claims.Subject

	// decode body into payload
	var payload passwordChangePayload
	err = json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		service.logger.Infof("client %s: password change for user '%s' failed (payload error: %s)", r.RemoteAddr, username, err)
		return output.JsonErrUnauthorized
	}

	// fetch the password hash from storage
	user, err := service.storage.GetOneUserByName(username)
	if err != nil {
		// shouldn't be possible since header was valid
		service.logger.Errorf("client %s: password change for user '%s' failed (bad username: %s)", r.RemoteAddr, username, err)
		return output.JsonErrUnauthorized
	}

	// confirm current password is correct
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(payload.CurrentPassword))
	if err != nil {
		service.logger.Infof("client %s: password change for user '%s' failed (bad password: %s)", r.RemoteAddr, username, err)
		return output.JsonErrUnauthorized
	}

	// Auth confirmed OK

	// verify new password matches
	if payload.NewPassword != payload.ConfirmNewPassword {
		err = fmt.Errorf("client %s: password change for user '%s' failed (new password did not match confirmation)", r.RemoteAddr, username)
		service.logger.Info(err)
		return output.JsonErrValidationFailed(err)
	}

	// don't enforce any password requirements other than it needs to exist
	if len(payload.NewPassword) < 1 {
		err = fmt.Errorf("client %s: password change for user '%s' failed (new password not specified)", r.RemoteAddr, username)
		service.logger.Info(err)
		return output.JsonErrValidationFailed(err)
	}

	// generate new password hash
	newPasswordHash, err := bcrypt.GenerateFromPassword([]byte(payload.NewPassword), BcryptCost)
	if err != nil {
		err = fmt.Errorf("client %s: password change for user '%s' failed (internal error: %s)", r.RemoteAddr, username, err)
		service.logger.Error(err)
		return output.JsonErrInternal(err)
	}

	// update password in storage
	userId, err := service.storage.UpdateUserPassword(username, string(newPasswordHash))
	if err != nil {
		err = fmt.Errorf("client %s: password change for user '%s' failed (storage error: %s)", r.RemoteAddr, username, err)
		service.logger.Error(err)
		return output.JsonErrStorageGeneric(err)
	}

	// log success (before response since new pw already saved)
	service.logger.Infof("client %s: password change for user '%s' succeeded", r.RemoteAddr, username)

	// return response to client
	response := &output.JsonResponse{}
	response.StatusCode = http.StatusOK
	response.Message = fmt.Sprintf("password changed for user '%s' (id: %d)", username, userId)

	err = service.output.WriteJSON(w, response)
	if err != nil {
		service.logger.Errorf("failed to write json (%s)", err)
		// detailed error is OK here because the user passed auth checks
		return output.JsonErrWriteJsonError(err)
	}

	return nil
}
