package auth

import (
	"encoding/json"
	"legocerthub-backend/pkg/output"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

// authResponse contains the JSON response for both
// login and refresh (refresh token is in a cookie
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

// Login takes the loginPayload and determines if the username is in
// storage and if the password matches the hash. If so, an Access Token
// is returned in JSON and a refresh token is sent in a cookie.
func (service *Service) Login(w http.ResponseWriter, r *http.Request) (err error) {
	var payload loginPayload

	// decode body into payload
	err = json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		service.logger.Info(err)
		return output.ErrUnauthorized
	}

	// fetch the password hash from storage
	user, err := service.storage.GetOneUserByName(payload.Username)
	if err != nil {
		service.logger.Info(err)
		return output.ErrUnauthorized
	}

	// compare
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(payload.Password))
	if err != nil {
		service.logger.Info(err)
		return output.ErrUnauthorized
	}

	// user and password now verified, make auth
	auth, err := service.createAuth(user.Username)
	if err != nil {
		service.logger.Error(err)
		return output.ErrInternal
	}

	// save auth's session in manager
	err = service.sessionManager.new(auth.SessionClaims)
	if err != nil {
		service.logger.Error(err)
		return output.ErrUnauthorized
	}

	// return response to client
	response := authResponse{}
	response.Status = http.StatusOK
	response.Message = "authenticated"
	response.AccessToken = auth.AccessToken
	response.SessionClaims = auth.SessionClaims

	service.logger.Infof("user '%s' logged in", auth.SessionClaims.Subject)

	// write auth cookies (part of response)
	auth.writeCookies(w)

	_, err = service.output.WriteJSON(w, response.Status, response, "response")
	if err != nil {
		service.logger.Error(err)
		return output.ErrWriteJsonFailed
	}

	return nil
}

// Refresh takes the RefreshToken, confirms it is valid and from a valid auth
// and then returns a new Access Token to the client.
func (service *Service) Refresh(w http.ResponseWriter, r *http.Request) (err error) {
	// get the refresh token cookie from request
	cookie, err := r.Cookie(refreshCookieName)
	if err != nil {
		service.logger.Info(err)
		return output.ErrUnauthorized
	}

	// validate cookie and get claims
	refreshCookie := refreshCookie(*cookie)
	oldClaims, err := refreshCookie.valid(service.refreshJwtSecret)
	if err != nil {
		service.logger.Info(err)
		return output.ErrUnauthorized
	}

	// refresh token verified, make new auth
	auth, err := service.createAuth(oldClaims.Subject)
	if err != nil {
		service.logger.Error(err)
		return output.ErrInternal
	}

	// refresh session in manager (remove old, add new)
	err = service.sessionManager.refresh(*oldClaims, auth.SessionClaims)
	if err != nil {
		service.logger.Error(err)
		return output.ErrUnauthorized
	}

	// return response (new auth) to client
	response := authResponse{}
	response.Status = http.StatusOK
	response.Message = "refreshed"
	response.AccessToken = auth.AccessToken
	response.SessionClaims = auth.SessionClaims
	// write auth cookies (part of response)
	auth.writeCookies(w)

	_, err = service.output.WriteJSON(w, response.Status, response, "response")
	if err != nil {
		service.logger.Error(err)
		return output.ErrWriteJsonFailed
	}

	return nil
}

// Logout deletes the client cookies.
func (service *Service) Logout(w http.ResponseWriter, r *http.Request) (err error) {
	// logic flow different from logout, still want succesful logout even
	// if cookie doesn't parse or token is expired.
	// However, if there is no cookie, don't try to remove session from
	// manager since there are no claims to work with.
	// get the refresh token cookie from request
	cookie, err := r.Cookie(refreshCookieName)
	// if cookie okay
	if err == nil {
		// validate cookie and get claims
		refreshCookie := refreshCookie(*cookie)
		oldClaims, err := refreshCookie.valid(service.refreshJwtSecret)
		// if token okay
		if err == nil {
			// remove session in manager
			err = service.sessionManager.close(*oldClaims)
			if err != nil {
				service.logger.Errorf("logout error for '%s': %s", oldClaims.Subject, err)
				// do return, there was an error with the token received
				return output.ErrUnauthorized
			}
			service.logger.Infof("user '%s' logged out", oldClaims.Subject)
		} else {
			service.logger.Info(err)
			// don't return, keep working
		}
	} else {
		service.logger.Debug(err)
		// don't return, keep working
	}

	// return response (logged out)
	response := output.JsonResponse{}
	response.Status = http.StatusOK
	response.Message = "logged out"
	// delete auth cookies (part of response)
	deleteAuthCookies(w)

	_, err = service.output.WriteJSON(w, response.Status, response, "response")
	if err != nil {
		service.logger.Error(err)
		return output.ErrWriteJsonFailed
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
	// if not running https, error
	if !service.https && !service.devMode {
		return output.ErrUnavailableHttp
	}

	var payload passwordChangePayload

	// This route will be unsecured in the router because the claims need to be accessed.
	// validate jwt and get the claims (to confirm the username)
	claims, err := service.ValidAuthHeader(r.Header, w)
	if err != nil {
		return output.ErrUnauthorized
	}
	username := claims["sub"].(string)

	// decode body into payload
	err = json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		service.logger.Info(err)
		return output.ErrUnauthorized
	}

	// fetch the password hash from storage
	user, err := service.storage.GetOneUserByName(username)
	if err != nil {
		service.logger.Info(err)
		return output.ErrUnauthorized
	}

	// confirm current password is correct
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(payload.CurrentPassword))
	if err != nil {
		service.logger.Info(err)
		return output.ErrUnauthorized
	}

	// verify new password matches
	if payload.NewPassword != payload.ConfirmNewPassword {
		service.logger.Info("new password did not match confirm new password")
		return output.ErrValidationFailed
	}

	// TODO: Additional password complexity requirements?
	// verify password is long enough
	if len(payload.NewPassword) < 10 {
		service.logger.Info("password did not meet length requirement")
		return output.ErrValidationFailed
	}

	// generate new password hash
	newPasswordHash, err := bcrypt.GenerateFromPassword([]byte(payload.NewPassword), BcryptCost)
	if err != nil {
		service.logger.Error(err)
		return output.ErrInternal
	}

	// update password in storage
	userId, err := service.storage.UpdateUserPassword(username, string(newPasswordHash))
	if err != nil {
		service.logger.Error(err)
		return output.ErrStorageGeneric
	}

	// return response to client
	response := output.JsonResponse{}
	response.Status = http.StatusOK
	response.Message = "password changed"
	response.ID = userId

	_, err = service.output.WriteJSON(w, response.Status, response, "response")
	if err != nil {
		service.logger.Error(err)
		return output.ErrWriteJsonFailed
	}

	return nil
}
