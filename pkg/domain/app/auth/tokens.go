package auth

import (
	"legocerthub-backend/pkg/output"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// TODO: move jwt secrets
var accessJwtSecret = []byte("17842911225de55706cb6e417418c7a0d21c9ccaf1c4ec271e187b9bea951b03")
var refreshJwtSecret = []byte("de0bce3589c282acc4e917eb1af6f85521624681e7dded2542004d26d1f5e87b")

// expiration times
// const accessTokenExpiration = 2 * time.Minute
// const refreshTokenExpiration = 30 * time.Minute

const accessTokenExpiration = 10 * time.Second
const refreshTokenExpiration = 30 * time.Second

// token types
type accessToken string
type refreshToken string

// jwt signature method
var tokenSignatureMethod = jwt.SigningMethodHS256

// keyFunc provides the jwt.KeyFunc
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
func createAccessToken(username string) (accessToken, error) {
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
	tokenString, err := token.SignedString(accessJwtSecret)
	if err != nil {
		return "", err
	}

	return accessToken(tokenString), nil
}

// create refresh token
func createRefreshToken(username string) (refreshToken, error) {
	// make refresh token
	refreshExpiration := time.Now().Add(refreshTokenExpiration)
	claims := &jwt.RegisteredClaims{
		Subject:   username,
		ExpiresAt: jwt.NewNumericDate(refreshExpiration),
		NotBefore: jwt.NewNumericDate(time.Now()),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		// TODO: Issuer / Audiences domains
	}

	// create token and then signed token string
	token := jwt.NewWithClaims(tokenSignatureMethod, claims)
	refreshString, err := token.SignedString(refreshJwtSecret)
	if err != nil {
		return "", err
	}

	return refreshToken(refreshString), nil
}

// Valid (AccessToken) returns the token's claims if it is valid, otherwise
// an error is returned if there is any issue (e.g. token not valid)
func (tokenString *accessToken) valid() (claims jwt.MapClaims, err error) {
	// parse and validate token
	token, err := jwt.Parse(string(*tokenString), makeKeyFunc(accessJwtSecret))
	if err != nil {
		return nil, output.ErrUnauthorized
	}

	if !token.Valid {
		return nil, output.ErrUnauthorized
	}

	// map claims
	var ok bool
	if claims, ok = token.Claims.(jwt.MapClaims); !ok {
		return nil, output.ErrBadRequest
	}

	return claims, nil
}

// Valid (RefreshToken) returns the refresh token's claims if
// it the token is valid, otherwise an error is returned if there is any
// issue (e.g. token not valid)
func (tokenString *refreshToken) valid() (claims jwt.MapClaims, err error) {
	// parse and validate token
	token, err := jwt.Parse(string(*tokenString), makeKeyFunc(refreshJwtSecret))
	if err != nil {
		return nil, output.ErrUnauthorized
	}

	if !token.Valid {
		return nil, output.ErrUnauthorized
	}

	// map claims
	var ok bool
	if claims, ok = token.Claims.(jwt.MapClaims); !ok {
		return nil, output.ErrBadRequest
	}

	return claims, nil
}
