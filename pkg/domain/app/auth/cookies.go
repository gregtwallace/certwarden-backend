package auth

import (
	"legocerthub-backend/pkg/output"
	"net/http"

	"github.com/golang-jwt/jwt/v4"
)

const refreshCookieName = "refresh_token"

type RefreshCookie http.Cookie

// createRefreshCookie returns the refresh token cookie
func createRefreshCookie(refreshToken RefreshToken) *RefreshCookie {
	return &RefreshCookie{
		Name:     refreshCookieName,
		Value:    string(refreshToken),
		MaxAge:   int(refreshTokenExpiration.Seconds()),
		HttpOnly: true,
	}
}

// Valid (RefreshCookie) returns the refresh cookie's token's claims if
// it the token is valid, otherwise an error is returned if there is any
// issue (e.g. token not valid)
func (cookie *RefreshCookie) valid() (claims jwt.MapClaims, err error) {
	// confirm cookie name (should never trigger)
	if cookie.Name != refreshCookieName {
		return nil, output.ErrInternal
	}

	// parse and validate refresh token
	refreshToken := RefreshToken(cookie.Value)
	claims, err = refreshToken.valid()
	if err != nil {
		return nil, err
	}

	return claims, nil
}
