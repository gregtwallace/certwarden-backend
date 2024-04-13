package auth

import (
	"certwarden-backend/pkg/datatypes/safemap"
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

var errInvalidSessionID = errors.New("invalid session id")
var errAddExisting = errors.New("cannot add existing session id again, terminating all sessions for this subject")

// sessionManager stores and manages session data
type sessionManager struct {
	sessions *safemap.SafeMap[tokenClaims] // map[uuid]sessionClaims
}

// newSessionManager creates a new sessionManager
func newSessionManager() *sessionManager {
	sm := &sessionManager{
		sessions: safemap.NewSafeMap[tokenClaims](),
	}

	return sm
}

// new adds the session to the map of open sessions. If session already exists
// an error is returned and all sessions for the specific subject (user) are
// removed.
func (sm *sessionManager) new(session tokenClaims) error {
	// parse uuid to a sane string for map key
	uuidString := session.SessionID.String()
	if uuidString == "" {
		sm.closeSubject(session)
		return errInvalidSessionID
	}

	// check if session already exists
	exists, _ := sm.sessions.Add(uuidString, session)
	if exists {
		sm.closeSubject(session)
		return errAddExisting
	}

	return nil
}

// close removes the session from the map of open sessions. If session
// doesn't exist an error is returned and all sessions for the specific
// subject (user) are removed.
func (sm *sessionManager) close(session tokenClaims) error {
	// parse uuid to a sane string for map key
	uuidString := session.SessionID.String()
	if uuidString == "" {
		sm.closeSubject(session)
		return errInvalidSessionID
	}

	// remove and check if trying to remove non-existent
	delFunc := func(key string, _ tokenClaims) bool {
		return key == uuidString
	}

	deletedOk := sm.sessions.DeleteFunc(delFunc)
	if !deletedOk {
		sm.closeSubject(session)
		return fmt.Errorf("app auth: failed to close session %s, closing all sessions for %s", uuidString, session.Subject)
	}

	return nil
}

// refresh confirms the oldSession is present, removes it, and then adds the new
// session in its place. If the session doesn't exist or the new session
// already exists an error is returned and all sessions for the specific subject
// (user) are removed.
func (sm *sessionManager) refresh(oldSession, newSession tokenClaims) error {
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

// closeSubject deletes all sessions where the session's Subject is equal to
// the specified sessionClaims' Subject
func (sm *sessionManager) closeSubject(sc tokenClaims) {
	// delete func for close subject
	deleteFunc := func(k string, v tokenClaims) bool {
		// if map value Subject == this func param's (sc's) Subject, return true
		return v.Subject == sc.Subject
	}

	// run func against sessions map
	_ = sm.sessions.DeleteFunc(deleteFunc)
}

// startCleanerService starts a goroutine that is an indefinite for loop
// that checks for expired sessions and removes them. This is to
// prevent the accumulation of expired sessions that were never
// formally logged out of.
func (service *Service) startCleanerService(ctx context.Context, wg *sync.WaitGroup) {
	// log start and update wg
	service.logger.Info("starting auth session cleaner service")

	// delete func that checks values for expired session
	deleteFunc := func(k string, v tokenClaims) bool {
		if v.ExpiresAt.Unix() <= time.Now().Unix() {
			// if expiration has passed, delete
			service.logger.Infof("user '%s' logged out (expired)", v.Subject)
			return true
		}

		// else don't delete (valid)
		return false
	}

	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			// wait time is based on expiration of session token
			delayTimer := time.NewTimer(2 * sessionTokenExpiration)

			select {
			case <-ctx.Done():
				// ensure timer releases resources
				if !delayTimer.Stop() {
					<-delayTimer.C
				}

				// exit
				service.logger.Info("auth session cleaner service shutdown complete")
				return

			case <-delayTimer.C:
				// continue and run
			}

			// run delete func against sessions map
			_ = service.sessionManager.sessions.DeleteFunc(deleteFunc)
		}
	}()
}
