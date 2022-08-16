package auth

import (
	"net/http"
)

type authorization struct {
	AccessToken    accessToken     `json:"access_token"`
	SessionClaims  sessionClaims   `json:"session"`
	refreshCookie  *refreshCookie  `json:"-"`
	loggedInCookie *loggedInCookie `json:"-"` // not used on backend
}

// createAuth creates all of the necessary pieces of information
// for an auth response
func (service *Service) createAuth(username string) (auth authorization, err error) {
	// make access token
	auth.AccessToken, _, err = service.createAccessToken(username)
	if err != nil {
		return authorization{}, err
	}

	// refresh token / cookie
	var refreshToken refreshToken
	refreshToken, auth.SessionClaims, err = service.createRefreshToken(username)
	if err != nil {
		return authorization{}, err
	}

	auth.refreshCookie = createRefreshCookie(refreshToken)

	// logged in cookie
	auth.loggedInCookie = createLoggedInCookie()

	return auth, nil
}

// writeCookies writes the auth's cookies to the specified ResponseWriter
func (auth *authorization) writeCookies(w http.ResponseWriter) {
	// write logged in cookie
	loggedInCookie := http.Cookie(*auth.loggedInCookie)
	http.SetCookie(w, &loggedInCookie)

	// write refresh token cookie
	refreshCookie := http.Cookie(*auth.refreshCookie)
	http.SetCookie(w, &refreshCookie)
}

// deleteAuthCookies writes dummy cookies with max age -1 (delete now)
// to the specified ResponseWriter
func deleteAuthCookies(w http.ResponseWriter) {
	// logged in cookie
	loggedIn := createLoggedInCookie()
	loggedIn.MaxAge = -1

	// write logged in cookie
	loggedInCookie := http.Cookie(*loggedIn)
	http.SetCookie(w, &loggedInCookie)

	// refresh token cookie
	refresh := createRefreshCookie("")
	refresh.MaxAge = -1

	// write refresh token cookie
	refreshCookie := http.Cookie(*refresh)
	http.SetCookie(w, &refreshCookie)
}
