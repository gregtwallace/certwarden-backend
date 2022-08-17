package auth

import (
	"errors"
	"legocerthub-backend/pkg/datatypes"
)

var errInvalidUuid = errors.New("invalid uuid")
var errAddExisting = errors.New("cannot add existing uuid again, terminating all sessions for this subject")

// stores session data
type sessionManager struct {
	devMode  bool
	sessions *datatypes.SafeMap
}

// newSessionManager creates a new session manager
func newSessionManager(devMode bool) *sessionManager {
	sm := new(sessionManager)
	sm.devMode = devMode
	sm.sessions = datatypes.NewSafeMap()

	return sm
}

// new adds the session to the map of open sessions. If session already exists
// an error is returned and all sessions for the specific subject (user) are
// removed.
func (sm *sessionManager) new(session sessionClaims) error {
	// parse uuid to a sane string for map elementName
	uuidString := session.UUID.String()
	if uuidString == "" {
		sm.closeSubject(session.Subject)
		return errInvalidUuid
	}

	// check if session already exists
	exists, _ := sm.sessions.Add(uuidString, session)
	if exists {
		sm.closeSubject(session.Subject)
		return errAddExisting
	}

	return nil
}

// close removes the session from the map of open sessions. If session
// doesn't exist an error is returned and all sessions for the specific
// subject (user) are removed.
func (sm *sessionManager) close(session sessionClaims) error {
	// parse uuid to a sane string for map elementName
	uuidString := session.UUID.String()
	if uuidString == "" {
		sm.closeSubject(session.Subject)
		return errInvalidUuid
	}

	// check if trying to remove non-existent
	err := sm.sessions.Delete(uuidString)
	if err != nil {
		sm.closeSubject(session.Subject)
		return err
	}

	return nil
}

// refresh confirms the oldSession is present, removes it, and then adds the new
// session in its place. If the session doesn't exist or the new session
// already exists an error is returned and all sessions for the specific subject
// (user) are removed.
func (sm *sessionManager) refresh(oldSession, newSession sessionClaims) error {
	// remove old session (error if doesn't exist, so this is validation)
	err := sm.close(oldSession)
	// Only error if not in devMove. This accounts for React doing a double
	// call with StrictMode enabled in React dev. That is, if a stale session
	// is used (the 2nd call), instead of failing to refreah a 2nd time sm.close()
	// will close  the session created by React's first call and then NOT return an error
	// here. This will lead to the 2nd refresh call succeeding and providing a
	// newly created session. Because of the behavior of sm.close() the 1st
	// refresh call session is destroyed thus preventing extra sessions from piling
	// up from the double React calls.
	// This should NOT be the production behavior as this disables security measure
	// of checking for refresh token re-use.
	if err != nil && !sm.devMode {
		// closeSubject already called by sm.close()
		return err
	}

	// add new session (error if already exists)
	err = sm.new(newSession)
	if err != nil {
		// closeSubject already called by sm.close()
		return err
	}

	return nil
}

// closeSubject locks the safe map and ranges through it removing any sessions
// belonging to the specified subject (user)
func (sm *sessionManager) closeSubject(subject string) {
	sm.sessions.Lock()
	defer sm.sessions.Unlock()

	// range through all sessions
	for elementName, session := range sm.sessions.Map {
		if session.(sessionClaims).Subject == subject {
			// if subject matches, delete element
			delete(sm.sessions.Map, elementName)
		}
	}
}
