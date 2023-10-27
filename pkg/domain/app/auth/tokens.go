package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

// expiration times
const accessTokenExpiration = 2 * time.Minute
const sessionTokenExpiration = 15 * time.Minute

// expiration times for testing
// const accessTokenExpiration = 10 * time.Second
// const sessionTokenExpiration = 2 * time.Minute

// const accessTokenExpiration = 5 * time.Second
// const sessionTokenExpiration = 15 * time.Second

// jwt signature method
var tokenSignatureMethod = jwt.SigningMethodHS256

// token types
type accessToken string
type sessionToken string

// custom claims for tokens
type tokenClaims struct {
	jwt.RegisteredClaims
	SessionID uuid.UUID `json:"session_id"`
}

// newTokenClaims creates tokenClaims
func newTokenClaims(username string, uuid uuid.UUID, expirationDuration time.Duration) tokenClaims {
	return tokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   username,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expirationDuration)),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			// TODO: Issuer / Audiences domains ?
		},
		SessionID: uuid,
	}
}

// keyFunc provides the jwt.KeyFunc. This implementation screens for acceptable signature
// methods
func makeKeyFunc(secret []byte) func(*jwt.Token) (interface{}, error) {
	return func(token *jwt.Token) (interface{}, error) {
		// confirm signature is of proper type
		// see: https://auth0.com/blog/critical-vulnerabilities-in-json-web-token-libraries/
		if token.Method != tokenSignatureMethod {
			return nil, jwt.ErrSignatureInvalid
		}

		return secret, nil
	}
}

// validateTokenString checks the validity of the specified tokenString using the specified
// jwtSecret. if it is valid, the claims are returned.
func validateTokenString(tokenString string, jwtSecret []byte) (*tokenClaims, error) {
	// parse and validate token
	token, err := jwt.ParseWithClaims(tokenString, &tokenClaims{}, makeKeyFunc(jwtSecret))
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, err
	}

	// assert claims type
	claims, ok := token.Claims.(*tokenClaims)
	if !ok {
		return nil, err
	}

	return claims, nil
}
