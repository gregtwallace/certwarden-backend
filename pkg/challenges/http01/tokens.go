package http01

// AddToken adds a token to the slice of hosted tokens
func (service *Service) AddToken(token string) {
	service.mu.Lock()
	defer service.mu.Unlock()

	service.tokens = append(service.tokens, token)
}

// RemoveToken removes the specified token from the slice of
// hosted tokens
func (service *Service) RemoveToken(token string) {
	service.mu.Lock()
	defer service.mu.Unlock()

	for i := range service.tokens {
		if service.tokens[i] == token {
			service.tokens = append(service.tokens[:i], service.tokens[i+1:]...)
			break
		}
	}
}
