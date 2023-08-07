package http01internal

// Provision adds a keyAuth/token pair to those being hosted
func (service *Service) Provision(token string, keyAuth string) (err error) {
	// add new entry
	service.mu.Lock()
	defer service.mu.Unlock()

	service.tokens[token] = keyAuth

	return nil
}

// Deprovision removes a keyAuth/token pair to those being hosted
func (service *Service) Deprovision(token string, keyAuth string) (err error) {
	// keyAuth is unused in this function

	service.mu.Lock()
	defer service.mu.Unlock()

	delete(service.tokens, token)

	return nil
}
