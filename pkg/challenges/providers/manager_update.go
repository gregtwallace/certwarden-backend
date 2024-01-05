package providers

// unsafeUpdateProviderDomains updates the domains serviced by a provider, if no domains
// are specified, no modification is performed
func (mgr *Manager) unsafeUpdateProviderDomains(p *provider, newDomains []string) {
	// no domains == no-op
	if newDomains == nil || len(newDomains) < 1 {
		return
	}

	// remove existing domain -> p mappings
	for _, oldDomain := range p.Domains {
		delete(mgr.dP, oldDomain)
	}

	// update p's domains
	p.Domains = newDomains

	// add new blank domains slice (to overwrite pD)
	mgr.pD[p] = []string{}

	// add each domain
	for _, newDomain := range newDomains {
		// add domain to domains map
		mgr.dP[newDomain] = p

		// append domain to providers map
		mgr.pD[p] = append(mgr.pD[p], newDomain)
	}
}
