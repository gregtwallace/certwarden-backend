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
	AccessToken AccessToken `json:"access_token"`
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

	// user and password now verified, make token pair
	tokens, err := createTokenPair(user.Username)
	if err != nil {
		service.logger.Error(err)
		return output.ErrInternal
	}

	// TODO: save session (map[username]uuid of refresh token)

	// set refresh token cookie
	refreshCookie := http.Cookie(*tokens.refreshCookie)
	http.SetCookie(w, &refreshCookie)

	// return response to client
	response := authResponse{}
	response.Status = http.StatusOK
	response.Message = "authenticated"
	response.AccessToken = tokens.accessToken

	_, err = service.output.WriteJSON(w, response.Status, response, "response")
	if err != nil {
		service.logger.Error(err)
		return output.ErrWriteJsonFailed
	}

	return nil
}

// Refresh takes the RefreshToken, confirms it is valid and from a valid session
// and then returns a new Access Token to the client.
func (service *Service) Refresh(w http.ResponseWriter, r *http.Request) (err error) {
	// get the refresh token cookie from request
	cookie, err := r.Cookie(refreshCookieName)
	if err != nil {
		service.logger.Info(err)
		return output.ErrUnauthorized
	}

	// validate cookie and get claims
	refreshCookie := RefreshCookie(*cookie)
	claims, err := refreshCookie.valid()
	if err != nil {
		service.logger.Info(err)
		return err
	}

	// TODO
	// verify Refresh Token claimed uuid is in the active list
	// if valid and not in list, revoke all for this user
	// use claims from .valid() method

	// refresh token verified, make new Access Token
	subject, ok := claims["sub"].(string)
	if !ok {
		return output.ErrBadRequest
	}
	accessToken, err := createAccessToken(subject)
	if err != nil {
		service.logger.Error(err)
		return output.ErrInternal
	}

	// return response to client
	response := authResponse{}
	response.Status = http.StatusOK
	response.Message = "refreshed"
	response.AccessToken = accessToken

	_, err = service.output.WriteJSON(w, response.Status, response, "response")
	if err != nil {
		service.logger.Error(err)
		return output.ErrWriteJsonFailed
	}

	return nil
}
