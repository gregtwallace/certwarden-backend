package auth

import (
	"encoding/json"
	"legocerthub-backend/pkg/output"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

// authResponse contains the JSON response for both
// login and session (session token is in a cookie
// so the JSON struct doesn't change)
type authResponse struct {
	output.JsonResponse
	authorization
}

// loginPayload is the payload client's send to login
type loginPayload struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginUsingUserPwPayload takes the loginPayload, looks up the username in storage
// and validates the password. If so, an Access Token is returned in JSON and a refresh
// token is sent in a cookie.
func (service *Service) LoginUsingUserPwPayload(w http.ResponseWriter, r *http.Request) (err error) {
	// wrap handler to easily check err and delete cookies
	err = func() error {
		var payload loginPayload

		// log attempt
		service.logger.Infof("client %s: attempting login", r.RemoteAddr)

		// decode body into payload
		err = json.NewDecoder(r.Body).Decode(&payload)
		if err != nil {
			service.logger.Infof("client %s: login failed (payload error: %s)", r.RemoteAddr, err)
			return output.ErrUnauthorized
		}

		// fetch the password hash from storage
		user, err := service.storage.GetOneUserByName(payload.Username)
		if err != nil {
			service.logger.Infof("client %s: login failed (bad username: %s)", r.RemoteAddr, err)
			return output.ErrUnauthorized
		}

		// compare
		err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(payload.Password))
		if err != nil {
			service.logger.Infof("client %s: login failed (bad password: %s)", r.RemoteAddr, err)
			return output.ErrUnauthorized
		}

		// user and password now verified, make auth
		auth, err := service.newAuthorization(user.Username)
		if err != nil {
			service.logger.Errorf("client %s: login failed (internal error: %s)", r.RemoteAddr, err)
			return output.ErrInternal
		}

		// save auth's session in manager
		err = service.sessionManager.new(auth.SessionTokenClaims)
		if err != nil {
			service.logger.Errorf("client %s: login failed (internal error: %s)", r.RemoteAddr, err)
			return output.ErrUnauthorized
		}

		// return response to client
		response := authResponse{}
		response.Status = http.StatusOK
		response.Message = "authenticated"
		response.authorization = auth

		// write auth cookies (part of response)
		auth.writeSessionCookie(w)

		err = service.output.WriteJSON(w, response.Status, response, "response")
		if err != nil {
			return err
		}

		// log success
		service.logger.Infof("client %s: user '%s' logged in", r.RemoteAddr, auth.SessionTokenClaims.Subject)

		return nil
	}()

	// if err, delete session cookie and return err
	if err != nil {
		service.deleteSessionCookie(w)
		return err
	}

	return nil
}

// RefreshUsingCookie validates the SessionToken cookie and confirms its UUID is for a valid
// session. If so, it generates a new AccessToken and new SessionToken cookie and then sends both
// to the client.
func (service *Service) RefreshUsingCookie(w http.ResponseWriter, r *http.Request) (err error) {
	// wrap to easily check err and delete cookies
	err = func() error {
		// log attempt
		service.logger.Infof("client %s: attempting access token refresh", r.RemoteAddr)

		// validate cookie
		oldClaims, err := service.validateSessionCookie(r, w, "access token refresh")
		if err != nil {
			// error logged in validateCookieSession func and nice output error returned
			return err
		}

		// cookie & session verified, make new auth
		auth, err := service.newAuthorization(oldClaims.Subject)
		if err != nil {
			service.logger.Errorf("client %s: access token refresh failed (internal error: %s)", r.RemoteAddr, err)
			return output.ErrInternal
		}

		// refresh session in manager (remove old, add new)
		err = service.sessionManager.refresh(*oldClaims, auth.SessionTokenClaims)
		if err != nil {
			service.logger.Errorf("client %s: access token refresh failed (internal error: %s)", r.RemoteAddr, err)
			return output.ErrUnauthorized
		}

		// return response (new auth) to client
		response := authResponse{}
		response.Status = http.StatusOK
		response.Message = "refreshed"
		response.authorization = auth
		// write auth cookies (part of response)
		auth.writeSessionCookie(w)

		err = service.output.WriteJSON(w, response.Status, response, "response")
		if err != nil {
			return err
		}

		// log success
		service.logger.Infof("client %s: access token refresh for user '%s' succeeded", r.RemoteAddr, auth.SessionTokenClaims.Subject)

		return nil
	}()

	// if err, delete cookies and return err
	if err != nil {
		service.deleteSessionCookie(w)
		return err
	}

	return nil
}

// Logout logs the client out and removes the session from session manager
func (service *Service) Logout(w http.ResponseWriter, r *http.Request) (err error) {
	// log attempt
	service.logger.Infof("client %s: attempting logout", r.RemoteAddr)

	// get claims from auth header
	oldClaims, err := service.ValidateAuthHeader(r, w, "logout")
	if err != nil {
		service.logger.Errorf("client %s: logout failed (%s)", r.RemoteAddr, oldClaims.Subject, err)
		return output.ErrUnauthorized
	}

	// remove session in manager
	err = service.sessionManager.close(*oldClaims)
	if err != nil {
		service.logger.Errorf("client %s: logout for user '%s' failed (%s)", r.RemoteAddr, oldClaims.Subject, err)
		return output.ErrUnauthorized
	}

	// log success
	service.logger.Infof("client %s: logout for user '%s' succeeded", r.RemoteAddr, oldClaims.Subject)

	// return response (logged out)
	response := output.JsonResponse{}
	response.Status = http.StatusOK
	response.Message = "logged out"
	// delete session cookie
	service.deleteSessionCookie(w)

	err = service.output.WriteJSON(w, response.Status, response, "response")
	if err != nil {
		return err
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
func (service *Service) ChangePassword(w http.ResponseWriter, r *http.Request) (err error) {
	// log attempt
	service.logger.Infof("client %s: attempting password change", r.RemoteAddr)

	// validate jwt and get the claims (to confirm the username)
	claims, err := service.ValidateAuthHeader(r, w, "password change")
	if err != nil {
		service.logger.Infof("client %s: password change failed (bad auth header: %s)", r.RemoteAddr, err)
		return output.ErrUnauthorized
	}
	username := claims.Subject

	// decode body into payload
	var payload passwordChangePayload
	err = json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		service.logger.Infof("client %s: password change for user '%s' failed (payload error: %s)", r.RemoteAddr, username, err)
		return output.ErrUnauthorized
	}

	// fetch the password hash from storage
	user, err := service.storage.GetOneUserByName(username)
	if err != nil {
		// shouldn't be possible since header was valid
		service.logger.Errorf("client %s: password change for user '%s' failed (bad username: %s)", r.RemoteAddr, username, err)
		return output.ErrUnauthorized
	}

	// confirm current password is correct
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(payload.CurrentPassword))
	if err != nil {
		service.logger.Infof("client %s: password change for user '%s' failed (bad password: %s)", r.RemoteAddr, username, err)
		return output.ErrUnauthorized
	}

	// verify new password matches
	if payload.NewPassword != payload.ConfirmNewPassword {
		service.logger.Infof("client %s: password change for user '%s' failed (new password did not match confirmation)", r.RemoteAddr, username)
		return output.ErrValidationFailed
	}

	// don't enforce any password requirements other than it needs to exist
	if len(payload.NewPassword) < 1 {
		service.logger.Infof("client %s: password change for user '%s' failed (new password not specified)", r.RemoteAddr, username)
		return output.ErrValidationFailed
	}

	// generate new password hash
	newPasswordHash, err := bcrypt.GenerateFromPassword([]byte(payload.NewPassword), BcryptCost)
	if err != nil {
		service.logger.Errorf("client %s: password change for user '%s' failed (internal error: %s)", r.RemoteAddr, username, err)
		return output.ErrInternal
	}

	// update password in storage
	userId, err := service.storage.UpdateUserPassword(username, string(newPasswordHash))
	if err != nil {
		service.logger.Errorf("client %s: password change for user '%s' failed (internal error: %s)", r.RemoteAddr, username, err)
		return output.ErrStorageGeneric
	}

	// log success (before response since new pw already saved)
	service.logger.Infof("client %s: password change for user '%s' succeeded", r.RemoteAddr, username)

	// return response to client
	response := output.JsonResponse{}
	response.Status = http.StatusOK
	response.Message = "password changed"
	response.ID = userId

	err = service.output.WriteJSON(w, response.Status, response, "response")
	if err != nil {
		return err
	}

	return nil
}
