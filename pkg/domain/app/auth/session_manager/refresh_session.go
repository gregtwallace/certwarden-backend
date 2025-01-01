package session_manager

import (
	"certwarden-backend/pkg/output"
	"fmt"
	"net/http"
	"time"
)

// RefreshSession validates that r contains a valid cookie. If it does, the session in session manager
// is updated with a new authorization and the username and new authorization are returned. If it does
// not, any existing cookie is deleted using w and an error is returned.
func (sm *SessionManager) RefreshSession(r *http.Request, w http.ResponseWriter) (_username string, _ *authorization, _ *output.JsonError) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// wrap to easily check err and delete cookies
	session, outErr := func() (*session, *output.JsonError) {
		// get the session token cookie from request
		clientSessionCookie, err := r.Cookie(sessionCookieName)
		if err != nil {
			sm.logger.Infof("client %s: session refresh failed (bad cookie: %s)", r.RemoteAddr, err)
			return nil, output.JsonErrUnauthorized
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
			sm.logger.Infof("client %s: session refresh failed (invalid cookie value)", r.RemoteAddr)
			return nil, output.JsonErrUnauthorized
		}

		// found, check if expired
		if time.Now().After(time.Time(session.authorization.SessionExpiration)) {
			sm.logger.Infof("client %s: session refresh failed (session expired)", r.RemoteAddr)
			return nil, output.JsonErrUnauthorized
		}

		// not expired, return session
		return session, nil
	}()

	// if err, delete session cookie and return err
	if outErr != nil {
		sm.DeleteSessionCookie(w)
		return "", nil, outErr
	}

	// session was found, update it and return username and new auth
	var err error
	session.authorization, err = sm.newAuthorization()
	if err != nil {
		err = fmt.Errorf("client %s: session refresh failed (couldn't make new auth: %s)", r.RemoteAddr, err)
		sm.logger.Error(err)
		return "", nil, output.JsonErrInternal(err)
	}

	return session.username, session.authorization, nil
}
