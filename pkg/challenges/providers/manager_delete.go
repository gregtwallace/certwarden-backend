package providers

// unsafeDeleteProvider deletes the specified provider from manager
// and deletes its domains. It MUST be called from a Locked thread.
func (mgr *Manager) unsafeDeleteProvider(p *provider) {
	// delete each domain that used provider
	for _, domain := range p.Domains {
		delete(mgr.dP, domain)
	}

	// delete provider from provider slice
	for i, oneP := range mgr.providers {
		// when on correct provider, snip it out
		if p == oneP {
			mgr.providers = append(mgr.providers[:i], mgr.providers[i+1:]...)
			break
		}
	}
}
