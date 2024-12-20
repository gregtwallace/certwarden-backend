package dns01acmedns

import (
	"certwarden-backend/pkg/validation"
	"fmt"
	"strings"
)

// Configuration options
type Config struct {
	HostAddress string            `yaml:"acme_dns_address" json:"acme_dns_address"`
	Resources   []acmeDnsResource `yaml:"resources" json:"resources"`
}

// acmeDnsResource contains the needed configuration to update
// and acme-dns record and also the corresponding 'real' domain
// that will have a certificate issued for it.
// type acmeDnsResource struct {
// 	RealDomain string `yaml:"real_domain" json:"real_domain"`
// 	FullDomain string `yaml:"full_domain" json:"full_domain"`
// 	Username   string `yaml:"username" json:"username"`
// 	Password   string `yaml:"password" json:"password"`
// }

// validateConfig verifies the config meets requirements and returns an error if it does not
func validateConfig(cfg *Config) error {
	// must receive a config
	if cfg == nil {
		return errServiceComponent
	}

	// collect all validation errors (to return as a list)
	errStrings := []string{}

	// ACME DNS Host

	// require acme dns host to be specified as https://
	if !validation.HttpsUrlValid(cfg.HostAddress) {
		errStrings = append(errStrings, fmt.Sprintf("acme dns host address (%s) must be a properly formatted url and start with 'https://'", cfg.HostAddress))
	}

	// Resources - Real & Full Domain values + username / pw
	invalidDomains := []string{}
	userOrPwErr := false
	for _, resource := range cfg.Resources {
		// real
		valid := validation.DomainValid(resource.RealDomain, false)
		if !valid {
			invalidDomains = append(invalidDomains, resource.RealDomain)
		}

		// full
		valid = validation.DomainValid(resource.FullDomain, false)
		if !valid {
			invalidDomains = append(invalidDomains, resource.FullDomain)
		}

		// user / pw cant be blank
		if resource.Username == "" || resource.Password == "" {
			userOrPwErr = true
		}
	}
	if len(invalidDomains) != 0 {
		errStrings = append(errStrings, fmt.Sprintf("domains (%s) are not valid", strings.Join(invalidDomains, ", ")))
	}
	if userOrPwErr {
		errStrings = append(errStrings, "username and password must be specified on all resources")
	}

	// combine any errors and return

	if len(errStrings) != 0 {
		return fmt.Errorf("dns01acmedns: invalid config (%s)", strings.Join(errStrings, ", "))
	}

	return nil
}
