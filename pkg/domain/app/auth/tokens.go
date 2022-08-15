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
const accessTokenExpiration = 5 * time.Minute
const refreshTokenExpiration = 30 * time.Minute

// token types
type accessToken string
type refreshToken string

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
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
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
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
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
	token, err := jwt.Parse(string(*tokenString), func(token *jwt.Token) (interface{}, error) {
		return accessJwtSecret, nil
	})
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
	token, err := jwt.Parse(string(*tokenString), func(token *jwt.Token) (interface{}, error) {
		return refreshJwtSecret, nil
	})
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
