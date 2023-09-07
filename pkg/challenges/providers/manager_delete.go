package providers

// unsafeDeleteProvider deletes the specified provider from manager
// and deletes its domains. It MUST be called from a Locked thread.
func (mgr *Manager) unsafeDeleteProvider(p *provider) {
	// get domains list for provider
	domains := mgr.pD[p]
	// delete each domain that used provider
	for _, domain := range domains {
		delete(mgr.dP, domain)
	}

	// delete provider from provider map
	delete(mgr.pD, p)
}
