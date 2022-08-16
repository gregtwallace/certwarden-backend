package auth

import (
	"net/http"
)

type session struct {
	AccessToken    accessToken     `json:"access_token"`
	Claims         sessionClaims   `json:"claims"`
	refreshCookie  *refreshCookie  `json:"-"`
	loggedInCookie *loggedInCookie `json:"-"` // not used on backend
}

// createSession creates all of the necessary pieces of information
// for a session and then returns the new session
func (service *Service) createSession(username string) (newSession session, claims sessionClaims, err error) {
	// make access token
	newSession.AccessToken, _, err = service.createAccessToken(username)
	if err != nil {
		return session{}, sessionClaims{}, err
	}

	// refresh token / cookie
	refreshToken, claims, err := service.createRefreshToken(username)
	if err != nil {
		return session{}, sessionClaims{}, err
	}

	newSession.refreshCookie = createRefreshCookie(refreshToken)

	// logged in cookie
	newSession.loggedInCookie = createLoggedInCookie()

	return newSession, claims, nil
}

// write cookies writes the session cookies to the specified writer
func (s *session) writeCookies(w http.ResponseWriter) {
	// write logged in cookie
	loggedInCookie := http.Cookie(*s.loggedInCookie)
	http.SetCookie(w, &loggedInCookie)

	// write refresh token cookie
	refreshCookie := http.Cookie(*s.refreshCookie)
	http.SetCookie(w, &refreshCookie)
}

// delete cookies writes dummy cookies with max age -1 (delete now)
func deleteSessionCookies(w http.ResponseWriter) {
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
