package http01

import "legocerthub-backend/pkg/acme"

// AddToken adds a token to the slice of hosted tokens
func (service *Service) AddToken(token string, key acme.AccountKey) (err error) {
	// calculate key auth
	keyAuth, err := key.KeyAuthorization(token)
	if err != nil {
		return err
	}

	// add new entry
	service.mu.Lock()
	defer service.mu.Unlock()

	service.tokens[token] = keyAuth

	return nil
}

// RemoveToken removes the specified token from the slice of
// hosted tokens
func (service *Service) RemoveToken(token string) {
	service.mu.Lock()
	defer service.mu.Unlock()

	delete(service.tokens, token)
}
