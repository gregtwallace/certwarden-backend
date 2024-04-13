package app

import (
	"certwarden-backend/pkg/domain/app/auth"
	"certwarden-backend/pkg/output"
	"fmt"
	"net/http"
)

// middlewareApplyAuthJWT applies middleware that validates the jwt access token
// contained in the auth header. If it is not valid, an error is returned instead of
// executing next.
func middlewareApplyAuthJWT(next handlerFunc, auth *auth.Service) handlerFunc {
	return func(w http.ResponseWriter, r *http.Request) *output.Error {
		// shorten URI for logging
		trimmedURI := loggableRequestURI(r)

		_, err := auth.ValidateAuthHeader(r, w, fmt.Sprintf("%s %s", r.Method, trimmedURI))
		if err != nil {
			return output.ErrUnauthorized
		}

		// if valid, do next
		return next(w, r)
	}
}
