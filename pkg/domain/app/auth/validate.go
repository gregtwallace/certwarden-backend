package auth

import (
	"certwarden-backend/pkg/output"
	"fmt"
	"net/http"
)

const authHeader = "Authorization"

// ValidateAuthHeader validates that the header contains a valid access token. If valid,
// it also returns the validated claims. It also writes to w to indicate the response
// was impacted by the relevant header.
func (service *Service) ValidateAuthHeader(r *http.Request, w http.ResponseWriter, logTaskName string) (*tokenClaims, error) {
	// wrap to easily check err and delete cookies
	claims, err := func() (*tokenClaims, error) {
		// indicate Authorization header influenced the response
		w.Header().Add("Vary", authHeader)

		// if logTaskName unspecified, use a default
		if logTaskName == "" {
			logTaskName = "validation of jwt in auth header"
		}

		// get token string from header
		accessToken := accessToken(r.Header.Get(authHeader))

		// anonymous user
		if accessToken == "" {
			err := fmt.Errorf("client %s: %s failed (access token is missing)", r.RemoteAddr, logTaskName)
			service.logger.Debug(err)
			return nil, err
		}

		// validate token
		claims, err := validateTokenString(string(accessToken), service.accessJwtSecret)
		if err != nil {
			err = fmt.Errorf("client %s: %s failed (%s)", r.RemoteAddr, logTaskName, err)
			service.logger.Debug(err)
			return nil, err
		}

		return claims, nil
	}()

	// if err, delete session cookie and return err
	if err != nil {
		service.deleteSessionCookie(w)
		return nil, err
	}

	return claims, nil
}

// validateSessionCookie validates that r contains a valid cookie and that the session ID
// contained in the cookie's claims is for a valid session. If so, it returns the validated
// claims.
func (service *Service) validateSessionCookie(r *http.Request, w http.ResponseWriter, logTaskName string) (*tokenClaims, *output.JsonError) {
	// wrap to easily check err and delete cookies
	claims, outErr := func() (*tokenClaims, *output.JsonError) {
		// if logTaskName unspecified, use a default
		if logTaskName == "" {
			logTaskName = "validation of session cookie"
		}

		// get the session token cookie from request
		cookie, err := r.Cookie(sessionCookieName)
		if err != nil {
			service.logger.Infof("client %s: %s failed (bad cookie: %s)", r.RemoteAddr, logTaskName, err)
			return nil, output.JsonErrUnauthorized
		}

		// validate cookie and get claims
		sessionCookie := sessionCookie(*cookie)
		claims, err := sessionCookie.valid(service.sessionJwtSecret)
		if err != nil {
			service.logger.Infof("client %s: %s failed (bad cookie: %s)", r.RemoteAddr, logTaskName, err)
			return nil, output.JsonErrUnauthorized
		}

		// verify session is still valid
		_, exists := service.sessionManager.sessions.Read(claims.SessionID.String())
		if !exists {
			service.logger.Infof("client %s: %s failed (session no longer valid)", r.RemoteAddr, logTaskName)
			return nil, output.JsonErrUnauthorized
		}

		return claims, nil
	}()

	// if err, delete session cookie and return err
	if outErr != nil {
		service.deleteSessionCookie(w)
		return nil, outErr
	}

	return claims, nil
}
