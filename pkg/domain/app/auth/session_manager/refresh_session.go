package session_manager

import (
	"errors"
	"fmt"
	"net/http"
	"time"
)

// RefreshSession validates that r contains a valid cookie. If it does, the session in session manager
// is updated with a new authorization and the username and new authorization are returned. If it does
// not, any existing cookie is deleted using w and an error is returned.
func (sm *SessionManager) RefreshSession(r *http.Request, w http.ResponseWriter) (_username string, _ *authorization, _ error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// wrap to easily check err and delete cookies
	session, err := func() (*session, error) {
		// get the session token cookie from request
		clientSessionCookie, err := r.Cookie(sessionCookieName)
		if err != nil {
			return nil, fmt.Errorf("bad cookie: %s", err)
		}

		// check for matching session token
		var session *session
		for _, s := range sm.sessions {
			if s.authorization.sessionCookie.Value == clientSessionCookie.Value {
				session = s
				break
			}
		}
		if session == nil {
			return nil, errors.New("invalid cookie value")
		}

		// found, check if expired
		if time.Now().After(time.Time(session.authorization.SessionExpiration)) {
			return nil, errors.New("session expired")
		}

		// not expired, return session
		return session, nil
	}()

	// if err, delete session cookie and return err
	if err != nil {
		sm.DeleteSessionCookie(w)
		return "", nil, err
	}

	// session was found, update it and return username and new auth
	session.authorization, err = sm.newAuthorization()
	if err != nil {
		return "", nil, fmt.Errorf("couldn't make new auth: %s", err)
	}

	return session.username, session.authorization, nil
}
