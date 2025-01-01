package session_manager

import (
	"certwarden-backend/pkg/output"
	"certwarden-backend/pkg/randomness"
	"encoding/base64"
	"net/http"
	"strconv"
	"time"
)

// expiration times
const accessTokenExp = 2 * time.Minute
const sessionExp = 15 * time.Minute

// expiration times for testing
// const accessTokenExp = 10 * time.Second
// const sessionExp = 1 * time.Minute

// const accessTokenExp = 5 * time.Second
// const sessionExp = 15 * time.Second

// cookie param
const sessionCookieName = "session_token"

// JSONTime is used to output time.Time as a unix timestamp
type jsonTime time.Time

func (t jsonTime) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Itoa(int(time.Time(t).Unix()))), nil
}

// AuthResponse contains the JSON response for successful login or refresh
type AuthResponse struct {
	output.JsonResponse
	Authorization *authorization `json:"authorization"`
}

// authorization contains data to send to a client after the client has been
// authenticated & authorized
type authorization struct {
	AccessToken           string       `json:"access_token"`
	AccessTokenExpiration jsonTime     `json:"access_token_exp"`
	SessionExpiration     jsonTime     `json:"session_exp"`
	sessionCookie         *http.Cookie `json:"-"`
}

// newAuthorization creates a new authorization
func (sm *SessionManager) newAuthorization() (*authorization, error) {
	// access token
	accessTokenBytes, err := randomness.Generate32ByteSecret()
	if err != nil {
		return nil, err
	}
	accessToken := base64.RawURLEncoding.EncodeToString(accessTokenBytes)

	// session token
	sessionTokenBytes, err := randomness.Generate32ByteSecret()
	if err != nil {
		return nil, err
	}
	sessionToken := base64.RawURLEncoding.EncodeToString(sessionTokenBytes)

	// assemble auth
	now := time.Now()
	return &authorization{
		AccessToken:           accessToken,
		AccessTokenExpiration: jsonTime(now.Add(accessTokenExp)),
		SessionExpiration:     jsonTime(now.Add(sessionExp)),
		sessionCookie:         sm.createSessionCookie(sessionToken),
	}, nil
}

// cookie

// createSessionCookie creates the session cookie using the session token
func (sm *SessionManager) createSessionCookie(sessionToken string) *http.Cookie {
	// make cookie secure if secure https channel is available
	secureCookie := sm.https

	// strict same site for security, unless configured for cross origins
	sameSiteMode := http.SameSiteStrictMode

	// cross origin requires same site None
	// if same site is set to None, cookie must be secure per browser spec
	// see: https://developers.google.com/search/blog/2020/01/get-ready-for-new-samesitenone-secure
	// this will require user to log back in every 2 minutes when server is running
	// in http mode but this is unavoidable
	if sm.corsPermitted {
		sameSiteMode = http.SameSiteNoneMode
		secureCookie = true
	}

	return &http.Cookie{
		Name:     sessionCookieName,
		Value:    string(sessionToken),
		MaxAge:   int(sessionExp.Seconds()),
		Secure:   secureCookie,
		HttpOnly: true,
		SameSite: sameSiteMode,
	}
}

// writeSessionCookie writes the auth's session cookie to w
func (auth *authorization) WriteSessionCookie(w http.ResponseWriter) {
	cookie := http.Cookie(*auth.sessionCookie)
	http.SetCookie(w, &cookie)
}

// deleteSessionCookie writes a dummy session cookie with max age -1 (delete now) to w
func (sm *SessionManager) DeleteSessionCookie(w http.ResponseWriter) {
	// make dummy cookie
	dCookie := sm.createSessionCookie("")
	dCookie.MaxAge = -1

	// write cookie
	cookie := http.Cookie(*dCookie)
	http.SetCookie(w, &cookie)
}
