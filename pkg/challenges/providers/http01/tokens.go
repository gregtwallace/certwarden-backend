package http01

// AddToken adds a token to the slice of hosted tokens
func (service *Service) AddToken(token string, keyAuth string) {
	// add new entry
	service.mu.Lock()
	defer service.mu.Unlock()

	service.tokens[token] = keyAuth
}

// RemoveToken removes the specified token from the slice of
// hosted tokens
func (service *Service) RemoveToken(token string) {
	service.mu.Lock()
	defer service.mu.Unlock()

	delete(service.tokens, token)
}
