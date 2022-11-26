package http01internal

// AddToken adds a token to the slice of hosted tokens
func (service *Service) Provision(token string, keyAuth string) (err error) {
	// add new entry
	service.mu.Lock()
	defer service.mu.Unlock()

	service.tokens[token] = keyAuth

	return nil
}

// RemoveToken removes the specified token from the slice of
// hosted tokens
func (service *Service) Deprovision(token string, keyAuth string) (err error) {
	// keyAuth is unused in this function

	service.mu.Lock()
	defer service.mu.Unlock()

	delete(service.tokens, token)

	return nil
}
