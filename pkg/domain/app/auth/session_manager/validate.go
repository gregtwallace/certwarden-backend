package session_manager

import (
	"fmt"
	"net/http"
	"time"
)

const authHeader = "Authorization"

// ValidateAuthHeader validates that the header contains a valid access token. If invalid,
// an error is returned. It also writes to w to indicate the response was impacted by the
// relevant header.
func (sm *SessionManager) ValidateAuthHeader(r *http.Request, w http.ResponseWriter, logTaskName string) (username string, _ error) {
	// indicate Authorization header influenced the response
	w.Header().Add("Vary", authHeader)

	// if logTaskName unspecified, use a default
	if logTaskName == "" {
		logTaskName = "validation of auth header"
	}

	// get token string from header
	clientAccessToken := r.Header.Get(authHeader)

	// anonymous user
	if clientAccessToken == "" {
		err := fmt.Errorf("client %s: %s failed (access token is missing)", r.RemoteAddr, logTaskName)
		sm.logger.Debug(err)
		return "", err
	}

	// validate token
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var session *session
	for _, s := range sm.sessions {
		if s.authorization.AccessToken == clientAccessToken {
			session = s
			break
		}
	}
	if session == nil {
		err := fmt.Errorf("client %s: %s failed (invalid access token)", r.RemoteAddr, logTaskName)
		sm.logger.Debug(err)
		return "", err
	}

	// found, check if expired
	if time.Now().After(time.Time(session.authorization.AccessTokenExpiration)) {
		err := fmt.Errorf("client %s: %s failed (access token expired)", r.RemoteAddr, logTaskName)
		sm.logger.Debug(err)
		return "", err
	}

	return session.username, nil
}
