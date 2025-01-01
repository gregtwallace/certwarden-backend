package auth

import (
	"certwarden-backend/pkg/domain/app/auth/session_manager"
	"certwarden-backend/pkg/output"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

const localUsernamePrefix = "local|"

// loginPayload is the payload client's send to login
type loginPayload struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LocalPostLogin takes the loginPayload, looks up the username in storage
// and validates the password. If so, an Access Token is returned in JSON and a refresh
// token is sent in a cookie.
func (service *Service) LocalPostLogin(w http.ResponseWriter, r *http.Request) *output.JsonError {
	if !service.methodLocalEnabled() {
		return output.JsonErrNotFound(errors.New("auth: local login is not configured"))
	}

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
		user, err := service.local.storage.GetOneUserByName(payload.Username)
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

		// user and password now verified, make new session
		username := localUsernamePrefix + user.Username
		auth, err := service.sessionManager.NewSession(username)
		if err != nil {
			service.logger.Errorf("client %s: login failed (internal error: %s)", r.RemoteAddr, err)
			return output.JsonErrInternal(nil)
		}

		// return response to client
		response := &session_manager.AuthResponse{}
		response.StatusCode = http.StatusOK
		response.Message = fmt.Sprintf("user '%s' logged in", username)
		response.Authorization = auth

		// write response
		auth.WriteSessionCookie(w)
		err = service.output.WriteJSON(w, response)
		if err != nil {
			service.logger.Errorf("failed to write json (%s)", err)
			// detailed error is OK here because the user passed auth checks
			return output.JsonErrWriteJsonError(err)
		}

		// log success
		service.logger.Infof("client %s: user '%s' logged in", username)

		return nil
	}()

	// if err, delete session cookie and return err
	if outErr != nil {
		service.sessionManager.DeleteSessionCookie(w)
		return outErr
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
func (service *Service) LocalChangePassword(w http.ResponseWriter, r *http.Request) *output.JsonError {
	if !service.methodLocalEnabled() {
		return output.JsonErrNotFound(errors.New("auth: change password if for local logins, which are not configured"))
	}

	// log attempt
	service.logger.Infof("client %s: attempting password change", r.RemoteAddr)

	// validate access token and get username
	username, err := service.ValidateAuthHeader(r, w, "password change")
	if err != nil {
		service.logger.Infof("client %s: password change failed (bad auth header: %s)", r.RemoteAddr, err)
		return output.JsonErrUnauthorized
	}

	// if not a local user, error
	if !strings.HasPrefix(username, localUsernamePrefix) {
		err = fmt.Errorf("client %s: password change failed (user '%s' is not a local user)", r.RemoteAddr, username)
		service.logger.Info(err)
		return output.JsonErrValidationFailed(err)
	}

	// decode body into payload
	var payload passwordChangePayload
	err = json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		err = fmt.Errorf("client %s: password change for user '%s' failed (payload error: %s)", r.RemoteAddr, username, err)
		service.logger.Info(err)
		return output.JsonErrValidationFailed(err)
	}

	// fetch the password hash from storage
	user, err := service.local.storage.GetOneUserByName(username)
	if err != nil {
		// shouldn't be possible since header was valid
		err = fmt.Errorf("client %s: password change for user '%s' failed (bad username: %s)", r.RemoteAddr, username, err)
		service.logger.Error(err)
		return output.JsonErrInternal(err)
	}

	// confirm current password is correct
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(payload.CurrentPassword))
	if err != nil {
		err = fmt.Errorf("client %s: password change for user '%s' failed (bad password: %s)", r.RemoteAddr, username, err)
		service.logger.Info(err)
		return output.JsonErrValidationFailed(err)
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
	userId, err := service.local.storage.UpdateUserPassword(username, string(newPasswordHash))
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
