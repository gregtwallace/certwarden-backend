package auth

import (
	"certwarden-backend/pkg/domain/app/auth/session_manager"
	"certwarden-backend/pkg/output"
	"certwarden-backend/pkg/randomness"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

// OIDCLogin checks for a state parameter. If state is found, the client is redirected to the OIDCLoginFinalize
// handler. If there is no state parameter, the rest of this handler runs directing the client to the OIDC
// login page of the Idp. If OIDC is not configured it returns an error.
func (service *Service) OIDCGetLogin(w http.ResponseWriter, r *http.Request) *output.JsonError {
	// return an error if OIDC is not in use
	if !service.methodOIDCEnabled() {
		return output.JsonErrNotFound(errors.New("auth: OIDC is not configured"))
	}

	// Option 1: Finish login with state/code
	stateVal := r.URL.Query().Get("state")
	if stateVal != "" {
		return service.OIDCLoginFinalize(w, r)
	}

	// Option 2: New login attempt - make state and do redirect
	// Parse redirect param
	redirectUri := r.URL.Query().Get("redirect_uri")
	redirectUrlParsed, err := url.Parse(redirectUri)
	if err != nil {
		err = fmt.Errorf("client %s: oidc redirect_uri '%s' failed to parse (%s)", r.RemoteAddr, redirectUri, err)
		service.logger.Debug(err)
		return output.JsonErrValidationFailed(err)
	}

	// remove any previous error param
	queryValues := redirectUrlParsed.Query()
	queryValues.Del("oidc_error")
	redirectUrlParsed.RawQuery = queryValues.Encode()

	// Validate Redirect is a permitted frontend
	// check if the redirect is to a frontend on the user specified RedirectURL (no err check as parse checked on startup)
	cfgApiRedirectURLParsed, _ := url.Parse(service.oidc.oauth2Config.RedirectURL)
	apiRedirectSchemeHostFrontend := fmt.Sprintf("%s://%s", cfgApiRedirectURLParsed.Scheme, cfgApiRedirectURLParsed.Host) + service.frontendURLPath
	// ok if redirect matches API URL + frontend path
	if !strings.HasPrefix(redirectUri, apiRedirectSchemeHostFrontend) {
		// else, check if origins + frontend path are prefix, which is also ok
		permitted := false
		for _, origin := range service.corsPermittedCrossOrigins {
			if strings.HasPrefix(redirectUri, origin+service.frontendURLPath) {
				permitted = true
				break
			}
		}
		if !permitted {
			// log error here as this shouldn't ever happen unless someone is deliberately misbehaving or there is a bug
			err = fmt.Errorf("auth: oidc redirect_uri '%s' not permitted", redirectUri)
			service.logger.Errorf("client %s: %s", r.RemoteAddr, err)
			return output.JsonErrValidationFailed(err)
		}
	}

	// generate state
	state, err := randomness.GenerateRandomByteSlice(24)
	if err != nil {
		service.logger.Errorf("client %s: failed to generate oidc state (%s)", r.RemoteAddr, err)
		// redirect to frontend to try again
		http.Redirect(w, r, oidcUnauthorizedErrorURL(redirectUrlParsed).String(), http.StatusFound)
		return nil
	}
	stateString := base64.URLEncoding.EncodeToString(state)

	// generate code_verifier
	codeVerifierSecret, err := randomness.Generate32ByteSecret()
	if err != nil {
		service.logger.Errorf("client %s: failed to generate oidc code verifier (%s)", r.RemoteAddr, err)
		// redirect to frontend to try again
		http.Redirect(w, r, oidcUnauthorizedErrorURL(redirectUrlParsed).String(), http.StatusFound)
		return nil
	}
	codeVerifierHex := hex.EncodeToString(codeVerifierSecret)

	// add pending session
	pendingSession := &oidcPendingSession{
		callerRedirectUrl: redirectUrlParsed,
		codeVerifierHex:   codeVerifierHex,
		createdAt:         time.Now(),
	}

	exists, _ := service.oidc.pendingSessions.Add(stateString, pendingSession)
	if exists {
		service.logger.Errorf("client %s: failed to generate new login attempt (oidc state duplicated somehow)", r.RemoteAddr)
		// redirect to frontend to try again
		http.Redirect(w, r, oidcUnauthorizedErrorURL(redirectUrlParsed).String(), http.StatusFound)
		return nil
	}

	// redirect, including code challenge & audience
	opts := []oauth2.AuthCodeOption{}
	opts = append(opts, oauth2.S256ChallengeOption(pendingSession.codeVerifierHex))
	opts = append(opts, oauth2.SetAuthURLParam("audience", fmt.Sprintf("%s://%s%s", cfgApiRedirectURLParsed.Scheme, cfgApiRedirectURLParsed.Host, service.apiURLPath)))
	http.Redirect(w, r, service.oidc.oauth2Config.AuthCodeURL(stateString, opts...), http.StatusFound)

	return nil
}

// OIDCGetCallback is the callback route that is used by the Idp after the user completes authentication
// It obtains id_token from the Idp which is used to create a backend session. A cookie is written
// for the session and then the caller is redirected back to the original frontend route.
func (service *Service) OIDCGetCallback(w http.ResponseWriter, r *http.Request) *output.JsonError {
	query := r.URL.Query()
	qStateString := query.Get("state")

	// pop state out of pending list (it will go back after Exchange)
	oidcStateObj, found := service.oidc.pendingSessions.Pop(qStateString)
	if !found {
		service.logger.Errorf("client %s: oidc state not found", r.RemoteAddr)
		return output.JsonErrUnauthorized
	}

	// if Idp sent an error
	qError := query.Get("error")
	qErrorDescription := query.Get("error_description")
	if qError != "" {
		service.logger.Infof("client %s: oidc callback failed (%s: %s)", r.RemoteAddr, qError, qErrorDescription)
		// redirect to frontend to try again
		http.Redirect(w, r, oidcUnauthorizedErrorURL(oidcStateObj.callerRedirectUrl).String(), http.StatusFound)
		return nil
	}

	// if pending session isn't at the right stage of auth, fail and trash that state
	if oidcStateObj.oauth2Token != nil {
		service.logger.Infof("client %s: oidc callback failed (pending session already has a token saved)", r.RemoteAddr)
		// redirect to frontend to try again
		http.Redirect(w, r, oidcUnauthorizedErrorURL(oidcStateObj.callerRedirectUrl).String(), http.StatusFound)
		return nil
	}

	// oauth2 exchange
	var err error
	qCode := query.Get("code")
	oidcStateObj.oauth2Token, err = service.oidc.oauth2Config.Exchange(
		r.Context(),
		qCode,
		oauth2.VerifierOption(oidcStateObj.codeVerifierHex),
	)
	if err != nil {
		service.logger.Infof("client %s: oidc exchange code for token failed (%s)", r.RemoteAddr, err)
		// redirect to frontend to try again
		http.Redirect(w, r, oidcUnauthorizedErrorURL(oidcStateObj.callerRedirectUrl).String(), http.StatusFound)
		return nil
	}

	// Extract the ID Token from OAuth2 token.
	rawIDToken, ok := oidcStateObj.oauth2Token.Extra("id_token").(string)
	if !ok {
		service.logger.Error("auth: oidc idp did not return an id_token")
		// redirect to frontend to try again
		http.Redirect(w, r, oidcUnauthorizedErrorURL(oidcStateObj.callerRedirectUrl).String(), http.StatusFound)
		return nil
	}

	// Parse and verify ID Token payload.
	// i.e., the AUTHENTICATION step
	oidcStateObj.oidcIDToken, err = service.oidc.idTokenVerifier.Verify(r.Context(), rawIDToken)
	if err != nil {
		service.logger.Infof("client %s: oidc id_token failed verification (%s)", r.RemoteAddr, err)
		// redirect to frontend to try again
		http.Redirect(w, r, oidcUnauthorizedErrorURL(oidcStateObj.callerRedirectUrl).String(), http.StatusFound)
		return nil
	}

	// Validate the required scopes were granted - https://datatracker.ietf.org/doc/html/rfc6749#section-3.3
	// if the scope is omitted, spec requires scope to exactly match the request (i.e., scope was accepted and
	// validation here isn't needed)
	// i.e., this is the AUTHORIZATION step
	responseScopeString, hasScope := oidcStateObj.oauth2Token.Extra("scope").(string)
	if hasScope {
		responseScopes := strings.Split(responseScopeString, " ")
		for _, requiredScope := range oidcRequiredScopes {
			found := false
			for _, responseScope := range responseScopes {
				if requiredScope == responseScope {
					found = true
					break
				}
			}

			if !found {
				service.logger.Infof("client %s: oidc user '%s' required scope '%s' was not granted", r.RemoteAddr, oidcStateObj.oidcIDToken.Subject, requiredScope)
				// redirect to frontend to try again
				http.Redirect(w, r, oidcUnauthorizedErrorURL(oidcStateObj.callerRedirectUrl).String(), http.StatusFound)
				return nil
			}
		}
	}

	// update redirect uri to contain the state and code
	queryValues := oidcStateObj.callerRedirectUrl.Query()
	queryValues.Set("state", qStateString)
	queryValues.Set("code", qCode)
	oidcStateObj.callerRedirectUrl.RawQuery = queryValues.Encode()

	// save code to pending session
	oidcStateObj.idpCode = qCode

	// add pending session back
	exists, _ := service.oidc.pendingSessions.Add(qStateString, oidcStateObj)
	if exists {
		err = errors.New("auth: failed to add pending login with id token (oidc state duplicated somehow)")
		service.logger.Errorf("client %s: %s", r.RemoteAddr, err)
		// redirect to frontend to try again (restore original query first)
		http.Redirect(w, r, oidcErrorURL(oidcStateObj.callerRedirectUrl, err).String(), http.StatusInternalServerError)
		return nil
	}

	// redirect to frontend for final login
	http.Redirect(w, r, oidcStateObj.callerRedirectUrl.String(), http.StatusFound)

	return nil
}

// OIDCLoginFinalize validates the state and code query params and if they're valid it returns a login
// response to the client (the same payload as a succesful local logon). Otherwise, an unauthorized
// response is returned.
func (service *Service) OIDCLoginFinalize(w http.ResponseWriter, r *http.Request) *output.JsonError {
	query := r.URL.Query()
	qStateString := query.Get("state")
	qCode := query.Get("code")

	// pop state out of pending list (its final now)
	oidcStateObj, found := service.oidc.pendingSessions.Pop(qStateString)
	if !found {
		service.logger.Errorf("client %s: oidc state not found", r.RemoteAddr)
		return output.JsonErrUnauthorized
	}

	// validate code matches
	if qCode == "" || qCode != oidcStateObj.idpCode {
		service.logger.Infof("client %s: login failed oidc state's code did not match", r.RemoteAddr)
		return output.JsonErrUnauthorized
	}

	// validation done
	// make extra func obj
	extraFuncs := &oidcExtraFuncs{
		cfg:             service.oidc.oauth2Config,
		idTokenVerifier: service.oidc.idTokenVerifier,
		token: &expectedToken{
			// only RefreshToken is needed for extra funcs
			RefreshToken: oidcStateObj.oauth2Token.RefreshToken,
		},
	}

	// make new session
	username := oidcUsernamePrefix + oidcStateObj.oidcIDToken.Subject
	auth, err := service.sessionManager.NewSession(username, extraFuncs)
	if err != nil {
		service.logger.Errorf("client %s: login failed (internal error: %s)", r.RemoteAddr, err)
		return output.JsonErrInternal(nil)
	}

	// return response to client
	response := &session_manager.AuthResponse{}
	response.StatusCode = http.StatusOK
	response.Message = fmt.Sprintf("user '%s' logged in", username)
	response.Authorization = auth

	// write response
	auth.WriteSessionCookie(w)
	err = service.output.WriteJSON(w, response)
	if err != nil {
		service.logger.Errorf("failed to write json (%s)", err)
		// detailed error is OK here because the user passed auth checks
		return output.JsonErrWriteJsonError(err)
	}

	// log success
	service.logger.Infof("client %s: user '%s' logged in", r.RemoteAddr, username)

	return nil
}
