package app

import (
	"fmt"
	"legocerthub-backend/pkg/output"
	"net/http"
)

// middlewareApplyAuthJWT applies middleware that validates the jwt access token
// contained in the auth header. If it is not valid, an error is returned instead of
// executing next.
func (app *Application) middlewareApplyAuthJWT(next handlerFunc) handlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		// shorten URI for logging
		trimmedURI := loggableRequestURI(r)

		_, err := app.auth.ValidateAuthHeader(r, w, fmt.Sprintf("%s %s", r.Method, trimmedURI))
		if err != nil {
			return output.ErrUnauthorized
		}

		// if valid, execute next
		err = next(w, r)
		if err != nil {
			return err
		}

		return nil
	}
}
