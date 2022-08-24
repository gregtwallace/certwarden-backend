package auth

import (
	"legocerthub-backend/pkg/output"
	"net/http"

	"github.com/golang-jwt/jwt/v4"
)

const authHeader = "Authorization"

// ValidAccessToken validates that the header contains a valid
// access token. If valid, it also returns the validated claims.
// It also writes to r to indicate the response was impacted by
// the relevant header.
func (service *Service) ValidAuthHeader(header http.Header, w http.ResponseWriter) (claims jwt.MapClaims, err error) {
	// indicate Authorization header influenced the response
	w.Header().Add("Vary", authHeader)

	// get token string from header
	accessToken := accessToken(header.Get(authHeader))

	// anonymous user
	if accessToken == "" {
		return nil, output.ErrUnauthorized
	}

	// validate token
	claims, err = accessToken.valid(service.accessJwtSecret)
	if err != nil {
		return nil, err
	}

	return claims, nil
}
