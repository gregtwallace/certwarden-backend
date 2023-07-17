package auth

import (
	"errors"
	"net/http"
)

const refreshCookieName = "refresh_token"
const cookieMaxAge = refreshTokenExpiration

// cookie types
type refreshCookie http.Cookie

// createRefreshCookie creates the refresh cookie based on the
// specified refresh token and maxAge
func (service *Service) createRefreshCookie(refreshToken refreshToken) *refreshCookie {
	// strict same site for security, unless configured for cross origins
	sameSiteMode := http.SameSiteStrictMode
	if service.allowCrossOrigin {
		sameSiteMode = http.SameSiteNoneMode
	}

	return &refreshCookie{
		Name:   refreshCookieName,
		Value:  string(refreshToken),
		MaxAge: int(cookieMaxAge.Seconds()),
		// disable secure when server is in http mode
		// however, if same site is set to None, secure becomes true regardless of what
		// this value is set to and will require user to log back in every 2 minutes when
		// server is running in http mode.
		Secure:   service.https,
		HttpOnly: true,
		SameSite: sameSiteMode,
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
