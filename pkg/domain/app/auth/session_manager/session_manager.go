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

type extraFuncs interface {
	// RefreshCheck performs additional validation prior to returning a succesful refresh
	RefreshCheck() error
}

// session contains information about a given session
type session struct {
	authorization *authorization
	extraFuncs    extraFuncs
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
func (sm *SessionManager) NewSession(username string, usertype userType, extraFuncs extraFuncs) (*authorization, error) {
	auth, err := sm.newAuthorization(username, usertype)
	if err != nil {
		return nil, err
	}

	session := &session{
		authorization: auth,
	}
	if extraFuncs != nil {
		session.extraFuncs = extraFuncs
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
func (sm *SessionManager) DeleteSession(r *http.Request) (_deletedAuthorization *authorization, _ error) {
	// get access token string from header
	clientAccessToken := r.Header.Get(authHeader)

	// no token
	if clientAccessToken == "" {
		return nil, errors.New("access token is missing")
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
		return nil, errors.New("invalid access token")
	}

	// found, delete it
	deletedAuthorization := sm.sessions[sessionID].authorization
	delete(sm.sessions, sessionID)

	return deletedAuthorization, nil
}

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
