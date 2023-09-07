package providers

// updateProvider updates the information in manager relating to a provider
func (mgr *Manager) updateProvider(p *provider, newCfg providerConfig) {
	// delete old provider/domains
	mgr.unsafeDeleteProvider(p)

	// update provider cfg to new one
	p.Config = newCfg

	// add new provider/domains

	// add provider to providers map with empty domains
	mgr.pD[p] = []string{}

	// add each domain
	for _, domain := range newCfg.Domains() {
		// add domain to domains map
		mgr.dP[domain] = p

		// append domain to providers map
		mgr.pD[p] = append(mgr.pD[p], domain)
	}
}
