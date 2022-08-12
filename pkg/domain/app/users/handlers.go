package users

import (
	"encoding/json"
	"legocerthub-backend/pkg/output"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

// TODO: move jwt secret
var jwtKey = []byte("2dce505d96a53c5768052ee90f3df2055657518dad489160df9913f66042e160")

// loginPayload is the payload client's send to login
type loginPayload struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// login allows a user to login to the API
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

	// user and password now verified, build jwt
	expirationTime := time.Now().Add(5 * time.Minute)
	claims := &jwt.StandardClaims{
		Subject:   payload.Username,
		ExpiresAt: expirationTime.Unix(),
		NotBefore: time.Now().Unix(),
		IssuedAt:  time.Now().Unix(),
		// TODO: Issuer / Audiences domains
	}

	// create token and then signed token string
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		service.logger.Error(err)
		return output.ErrInternal
	}

	// return response to client
	_, err = service.output.WriteJSON(w, http.StatusOK, tokenString, "jwt")
	if err != nil {
		service.logger.Error(err)
		return output.ErrWriteJsonFailed
	}

	return nil
}
