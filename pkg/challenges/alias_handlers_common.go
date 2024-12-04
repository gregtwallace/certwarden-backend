package challenges

// output as an array to make use in frontend easier
type domainAliasJson struct {
	ChallengeDomain string `json:"challenge_domain"`
	ProvisionDomain string `json:"provision_domain"`
}

// domainAliases returns the current dns aliases as a slice of json
// output objects
func (service *Service) domainAliases() []domainAliasJson {
	m := make(map[string]string)
	service.dnsIDtoDomain.CopyToMap(m)

	// translate into output array
	aliases := []domainAliasJson{}
	for k, v := range m {
		aliases = append(aliases, domainAliasJson{
			ChallengeDomain: k,
			ProvisionDomain: v,
		})
	}

	return aliases
}
