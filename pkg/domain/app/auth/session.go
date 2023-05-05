package auth

import (
	"net/http"
)

type authorization struct {
	AccessToken   accessToken    `json:"access_token"`
	SessionClaims sessionClaims  `json:"session"`
	refreshCookie *refreshCookie `json:"-"`
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

	auth.refreshCookie = service.createRefreshCookie(refreshToken)

	return auth, nil
}

// writeCookies writes the auth's cookies to the specified ResponseWriter
func (auth *authorization) writeCookies(w http.ResponseWriter) {
	// write refresh token cookie
	refreshCookie := http.Cookie(*auth.refreshCookie)
	http.SetCookie(w, &refreshCookie)
}

// deleteAuthCookies writes dummy cookies with max age -1 (delete now)
// to the specified ResponseWriter
func (service *Service) deleteAuthCookies(w http.ResponseWriter) {
	// refresh token cookie
	refresh := service.createRefreshCookie("")
	refresh.MaxAge = -1

	// write refresh token cookie
	refreshCookie := http.Cookie(*refresh)
	http.SetCookie(w, &refreshCookie)
}
