package auth

import (
	"errors"
	"net/http"
)

const sessionCookieName = "session_token"
const sessionCookieMaxAge = sessionTokenExpiration

// cookie types
type sessionCookie http.Cookie

// createSessionCookie creates the session cookie using the session token
func (service *Service) createSessionCookie(token sessionToken) *sessionCookie {
	// make cookie secure if secure https channel is available
	secureCookie := service.https

	// strict same site for security, unless configured for cross origins
	sameSiteMode := http.SameSiteStrictMode

	// cross origin requires same site None
	// if same site is set to None, cookie must be secure per browser spec
	// see: https://developers.google.com/search/blog/2020/01/get-ready-for-new-samesitenone-secure
	// this will require user to log back in every 2 minutes when server is running
	// in http mode but this is unavoidable
	if service.allowCrossOrigin {
		sameSiteMode = http.SameSiteNoneMode
		secureCookie = true
	}

	return &sessionCookie{
		Name:     sessionCookieName,
		Value:    string(token),
		MaxAge:   int(sessionCookieMaxAge.Seconds()),
		Secure:   secureCookie,
		HttpOnly: true,
		SameSite: sameSiteMode,
	}
}

// valid returns the cookie's token's claims if the token is valid, otherwise an error
// is returned (e.g. token not valid)
func (cookie *sessionCookie) valid(jwtSecret []byte) (claims *tokenClaims, err error) {
	// confirm cookie name (should never trigger)
	if cookie.Name != sessionCookieName {
		return nil, errors.New("bad cookie name")
	}

	// parse and validate session token
	token := sessionToken(cookie.Value)
	claims, err = validateTokenString(string(token), jwtSecret)
	if err != nil {
		return nil, err
	}

	return claims, nil
}

// writeSessionCookie writes the auth's session cookie to w
func (auth *authorization) writeSessionCookie(w http.ResponseWriter) {
	cookie := http.Cookie(*auth.sessionCookie)
	http.SetCookie(w, &cookie)
}

// deleteSessionCookie writes a dummy session cookie with max age -1 (delete now) to w
func (service *Service) deleteSessionCookie(w http.ResponseWriter) {
	// make dummy cookie
	sCookie := service.createSessionCookie("")
	sCookie.MaxAge = -1

	// write cookie
	cookie := http.Cookie(*sCookie)
	http.SetCookie(w, &cookie)
}
