package providers

// unsafeUpdateProviderDomains updates the domains serviced by a provider, if no domains
// are specified, no modification is performed
func (mgr *Manager) unsafeUpdateProviderDomains(p *provider, newDomains []string) {
	// no domains == no-op
	if len(newDomains) <= 0 {
		return
	}

	// remove existing domain -> p mappings
	for _, oldDomain := range p.Domains {
		delete(mgr.dP, oldDomain)
	}

	// update p's domains
	p.Domains = newDomains

	// add each new domain to map
	for _, newDomain := range newDomains {
		mgr.dP[newDomain] = p
	}
}
