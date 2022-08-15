package auth

import "net/http"

type session struct {
	AccessToken    accessToken     `json:"access_token"`
	refreshCookie  *refreshCookie  `json:"-"`
	loggedInCookie *loggedInCookie `json:"-"` // not used on backend
}

// createSession creates all of the necessary pieces of information
// for a session and then returns the new session
func createSession(username string) (newSession session, err error) {
	// make access token
	newSession.AccessToken, err = createAccessToken(username)
	if err != nil {
		return session{}, err
	}

	// refresh token / cookie
	refreshToken, err := createRefreshToken(username)
	if err != nil {
		return session{}, err
	}

	newSession.refreshCookie = createRefreshCookie(refreshToken)

	// logged in cookie
	newSession.loggedInCookie = createLoggedInCookie()

	return newSession, nil
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
