package session_manager

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

var errInvalidSessionID = errors.New("invalid session id")
var errAddExisting = errors.New("cannot add session (duplicate id)")

// session contains information about a given session
type session struct {
	username      string
	authorization *authorization
}

// SessionManager stores and manages session data
type SessionManager struct {
	https         bool
	corsPermitted bool
	logger        *zap.SugaredLogger

	sessions map[string]*session // map[uuid]*session
	mu       sync.RWMutex
}

// newSessionManager creates a new sessionManager
func NewSessionManager(https bool, corsPermitted bool, logger *zap.SugaredLogger) *SessionManager {
	sm := &SessionManager{
		https:         https,
		corsPermitted: corsPermitted,
		logger:        logger,

		sessions: make(map[string]*session),
	}

	return sm
}

// NewSession creates a new session for the specified username and returns
// the newly created authorization.
func (sm *SessionManager) NewSession(username string) (*authorization, error) {
	auth, err := sm.newAuthorization()
	if err != nil {
		return nil, err
	}

	session := &session{
		username:      username,
		authorization: auth,
	}

	// make a session id
	uuid := uuid.New()
	uuidString := uuid.String()

	// add session
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// should never already exist but check anyway
	_, exists := sm.sessions[uuidString]
	if exists {
		return nil, errAddExisting
	}

	// if not, add the key and value
	sm.sessions[uuidString] = session

	return session.authorization, nil
}

// DeleteSession extracts the access token from r and deletes the matching session. If
// the access token is missing or does not match a session, an error is returned.
func (sm *SessionManager) DeleteSession(r *http.Request) (_username string, _ error) {
	// get access token string from header
	clientAccessToken := r.Header.Get(authHeader)

	// no token
	if clientAccessToken == "" {
		return "", errors.New("access token is missing")
	}

	// find matching session and remove it
	sm.mu.Lock()
	defer sm.mu.Unlock()

	var sessionID string
	for sid, session := range sm.sessions {
		if session.authorization.AccessToken == clientAccessToken {
			sessionID = sid
			break
		}
	}
	if sessionID == "" {
		return "", errors.New("invalid access token")
	}

	// found, delete it
	username := sm.sessions[sessionID].username
	delete(sm.sessions, sessionID)

	return username, nil
}

// // refresh confirms the oldSession is present, removes it, and then adds the new
// // session in its place. If the session doesn't exist or the new session
// // already exists an error is returned and all sessions for the specific subject
// // (user) are removed.
// func (sm *sessionManager) refresh(oldSession, newSession tokenClaims) error {
// 	// remove old session (error if doesn't exist, so this is validation)
// 	err := sm.close(oldSession)

// 	if err != nil {
// 		// closeSubject already called by sm.close()
// 		return err
// 	}

// 	// add new session (error if already exists)
// 	err = sm.new(newSession)
// 	if err != nil {
// 		// closeSubject already called by sm.close()
// 		return err
// 	}

// 	return nil
// }

// // closeSubject deletes all sessions where the session's Subject is equal to
// // the specified sessionClaims' Subject
// func (sm *sessionManager) closeSubject(sc tokenClaims) {
// 	// delete func for close subject
// 	deleteFunc := func(k string, v tokenClaims) bool {
// 		// if map value Subject == this func param's (sc's) Subject, return true
// 		return v.Subject == sc.Subject
// 	}

// 	// run func against sessions map
// 	_ = sm.sessions.DeleteFunc(deleteFunc)
// }

// StartCleanerService starts a goroutine that is an indefinite for loop
// that checks for expired sessions and removes them. This is to
// prevent the accumulation of expired sessions that were never
// formally logged out of.
func (sm *SessionManager) StartCleanerService(ctx context.Context, wg *sync.WaitGroup) {
	// log start and update wg
	sm.logger.Info("auth: starting auth session cleaner service")

	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			// wait time is based on expiration of session token
			delayTimer := time.NewTimer(2 * sessionExp)

			select {
			case <-ctx.Done():
				// ensure timer releases resources
				if !delayTimer.Stop() {
					<-delayTimer.C
				}

				// exit
				sm.logger.Info("auth session cleaner service shutdown complete")
				return

			case <-delayTimer.C:
				// continue and run
			}

			// delete expired sessions
			sm.mu.Lock()
			now := time.Now()

			for k, v := range sm.sessions {
				if now.After(time.Time(v.authorization.SessionExpiration)) {
					delete(sm.sessions, k)
				}
			}

			sm.mu.Unlock()
		}
	}()
}
