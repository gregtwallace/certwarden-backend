package auth

import (
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

// authorization contains data to send to a client to become authorized
type authorization struct {
	AccessToken        accessToken    `json:"access_token"`
	AccessTokenClaims  tokenClaims    `json:"access_token_claims"`
	SessionTokenClaims tokenClaims    `json:"session_claims"`
	sessionCookie      *sessionCookie `json:"-"`
}

// createAuth creates all of the necessary pieces of information for an auth response
func (service *Service) newAuthorization(username string) (auth authorization, err error) {
	// generate UUID
	uuid := uuid.New()

	// make access token claims
	auth.AccessTokenClaims = newTokenClaims(username, uuid, accessTokenExpiration)

	// create token and then signed token string
	token := jwt.NewWithClaims(tokenSignatureMethod, auth.AccessTokenClaims)
	tokenString, err := token.SignedString(service.accessJwtSecret)
	if err != nil {
		return authorization{}, err
	}
	auth.AccessToken = accessToken(tokenString)

	// make session token claims
	auth.SessionTokenClaims = newTokenClaims(username, uuid, sessionTokenExpiration)

	// create token and then signed token string
	token = jwt.NewWithClaims(tokenSignatureMethod, auth.SessionTokenClaims)
	tokenString, err = token.SignedString(service.sessionJwtSecret)
	if err != nil {
		return authorization{}, err
	}
	rToken := sessionToken(tokenString)

	// make session cookie
	auth.sessionCookie = service.createSessionCookie(rToken)

	return auth, nil
}
