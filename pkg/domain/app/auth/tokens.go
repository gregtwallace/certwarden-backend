package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

// expiration times
const accessTokenExpiration = 2 * time.Minute
const refreshTokenExpiration = 15 * time.Minute

// expiration times for testing
// const accessTokenExpiration = 10 * time.Second
// const refreshTokenExpiration = 2 * time.Minute

// const accessTokenExpiration = 5 * time.Second
// const refreshTokenExpiration = 15 * time.Second

// token types
type accessToken string
type refreshToken string

// custom session claims (used on refresh token which tracks the session)
type sessionClaims struct {
	jwt.RegisteredClaims
	UUID uuid.UUID `json:"uuid"`
}

// jwt signature method
var tokenSignatureMethod = jwt.SigningMethodHS256

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

// create access token
func (service *Service) createAccessToken(username string) (accessToken, jwt.RegisteredClaims, error) {
	// make claims
	claims := &jwt.RegisteredClaims{
		Subject:   username,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessTokenExpiration)),
		NotBefore: jwt.NewNumericDate(time.Now()),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		// TODO: Issuer / Audiences domains
	}

	// create token and then signed token string
	token := jwt.NewWithClaims(tokenSignatureMethod, claims)
	tokenString, err := token.SignedString(service.accessJwtSecret)
	if err != nil {
		return "", jwt.RegisteredClaims{}, err
	}

	return accessToken(tokenString), *claims, nil
}

// create refresh token
func (service *Service) createRefreshToken(username string) (refreshToken, sessionClaims, error) {
	// make refresh token
	refreshExpiration := time.Now().Add(refreshTokenExpiration)
	claims := &sessionClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   username,
			ExpiresAt: jwt.NewNumericDate(refreshExpiration),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			// TODO: Issuer / Audiences domains
		},
		UUID: uuid.New(),
	}

	// create token and then signed token string
	token := jwt.NewWithClaims(tokenSignatureMethod, claims)
	refreshString, err := token.SignedString(service.refreshJwtSecret)
	if err != nil {
		return "", sessionClaims{}, err
	}

	return refreshToken(refreshString), *claims, nil
}

// Valid (AccessToken) returns the token's claims if it is valid, otherwise
// an error is returned if there is any issue (e.g. token not valid)
func (tokenString *accessToken) valid(jwtSecret []byte) (claims jwt.MapClaims, err error) {
	// parse and validate token
	token, err := jwt.Parse(string(*tokenString), makeKeyFunc(jwtSecret))
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, err
	}

	// map claims
	var ok bool
	if claims, ok = token.Claims.(jwt.MapClaims); !ok {
		return nil, err
	}

	return claims, nil
}

// Valid (RefreshToken) returns the refresh token's claims if
// it the token is valid, otherwise an error is returned if there is any
// issue (e.g. token not valid)
func (tokenString *refreshToken) valid(jwtSecret []byte) (claims *sessionClaims, err error) {
	// parse and validate token
	token, err := jwt.ParseWithClaims(string(*tokenString), &sessionClaims{}, makeKeyFunc(jwtSecret))
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, err
	}

	// map claims
	var ok bool
	if claims, ok = token.Claims.(*sessionClaims); !ok {
		return nil, err
	}

	return claims, nil
}
