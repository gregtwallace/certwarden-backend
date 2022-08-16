package auth

import (
	"legocerthub-backend/pkg/output"
	"net/http"
)

const authHeader = "Authorization"

// ValidAccessToken validates that the header contains a valid
// access token. It also writes to r to indicate the response was
// impacted by the relevant header
func (service *Service) ValidAuthHeader(header http.Header, w http.ResponseWriter) (err error) {
	// indicate Authorization header influenced the response
	w.Header().Add("Vary", authHeader)

	// get token string from header
	accessToken := accessToken(header.Get(authHeader))

	// anonymous user
	if accessToken == "" {
		return output.ErrUnauthorized
	}

	// validate token
	_, err = accessToken.valid(service.accessJwtSecret)
	if err != nil {
		return err
	}

	return nil
}
