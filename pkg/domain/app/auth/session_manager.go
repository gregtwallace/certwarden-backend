package auth

import (
	"context"
	"errors"
	"legocerthub-backend/pkg/datatypes"
	"sync"
	"time"
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

	if err != nil {
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

// cleaner starts a goroutine that is an indefinite for loop
// that checks for expired sessions and removes them. This is to
// prevent the accumulation of expired sessions that were never
// formally logged out of.
func (service *Service) startCleanerService(ctx context.Context, wg *sync.WaitGroup) {
	// log start and update wg
	service.logger.Info("starting auth session cleaner service")
	wg.Add(1)

	go func() {
		// wait time is based on expiration of sessions (refresh)
		waitTime := 2 * refreshTokenExpiration
		for {
			select {
			case <-ctx.Done():
				// exit
				service.logger.Info("auth session cleaner service shutdown complete")
				wg.Done()
				return

			case <-time.After(waitTime):
				// continue and run
			}

			// lock sessions for cleaning
			service.sessionManager.sessions.Lock()

			// range through all sessions
			for elementName, session := range service.sessionManager.sessions.Map {
				if session.(sessionClaims).ExpiresAt.Unix() <= time.Now().Unix() {
					// if expiration has passed, delete element
					delete(service.sessionManager.sessions.Map, elementName)
				}
			}

			// done cleaning, unlock
			service.sessionManager.sessions.Unlock()
		}
	}()
}
