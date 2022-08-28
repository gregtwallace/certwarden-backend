package auth

import (
	"errors"
	"net/http"
)

const refreshCookieName = "refresh_token"
const loggedInCookieName = "logged_in"

const cookieMaxAge = refreshTokenExpiration

// cookie types
type refreshCookie http.Cookie
type loggedInCookie http.Cookie

// createRefreshCookie creates the refresh cookie based on the
// specified refresh token and maxAge
func createRefreshCookie(refreshToken refreshToken) *refreshCookie {
	return &refreshCookie{
		Name:     refreshCookieName,
		Value:    string(refreshToken),
		MaxAge:   int(cookieMaxAge.Seconds()),
		HttpOnly: true,
	}
}

// Valid (RefreshCookie) returns the refresh cookie's token's claims if
// it the token is valid, otherwise an error is returned if there is any
// issue (e.g. token not valid)
func (cookie *refreshCookie) valid(jwtSecret []byte) (claims *sessionClaims, err error) {
	// confirm cookie name (should never trigger)
	if cookie.Name != refreshCookieName {
		return nil, errors.New("bad cookie name")
	}

	// parse and validate refresh token
	refreshToken := refreshToken(cookie.Value)
	claims, err = refreshToken.valid(jwtSecret)
	if err != nil {
		return nil, err
	}

	return claims, nil
}

// createLoggedInCookie() http.Cookie
func createLoggedInCookie() *loggedInCookie {
	return &loggedInCookie{
		Name:     loggedInCookieName,
		Value:    "true",
		Path:     "/",
		MaxAge:   int(cookieMaxAge.Seconds()),
		HttpOnly: false,
	}
}
