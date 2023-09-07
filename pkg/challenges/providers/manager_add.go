package providers

import (
	"errors"
	"fmt"
	"legocerthub-backend/pkg/challenges/providers/dns01acmedns"
	"legocerthub-backend/pkg/challenges/providers/dns01acmesh"
	"legocerthub-backend/pkg/challenges/providers/dns01cloudflare"
	"legocerthub-backend/pkg/challenges/providers/dns01manual"
	"legocerthub-backend/pkg/challenges/providers/http01internal"
	"legocerthub-backend/pkg/randomness"
	"legocerthub-backend/pkg/validation"
	"reflect"
	"strings"
)

// addProvider creates the provider specified in cfg and adds it to manager
func (mgr *Manager) addProvider(cfg providerConfig) error {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	// verify every domain ir properly formatted, or verify this is wildcard cfg (* only)
	// and also verify all domains are available in manager
	domains := cfg.Domains()

	// validate domain names
	for _, domain := range domains {
		// check validity -or- wildcard
		if !validation.DomainValid(domain, false) && !(len(domains) == 1 && domains[0] == "*") {
			if domain == "*" {
				return errors.New("when using wildcard domain * it must be the only specified domain on the provider")
			}
			return fmt.Errorf("domain %s is not a validly formatted domain", domain)
		}

		// check manager availability
		_, exists := mgr.dP[domain]
		if exists {
			return fmt.Errorf("failed to configure domain %s, each domain can only be configured once", domain)
		}
	}

	// make provider service (switch based on cfg type (and thus which pkg to use))
	var serv Service
	var err error

	switch realCfg := cfg.(type) {
	case *http01internal.Config:
		serv, err = http01internal.NewService(mgr.childApp, realCfg)

	case *dns01manual.Config:
		serv, err = dns01manual.NewService(mgr.childApp, realCfg)

	case *dns01acmedns.Config:
		serv, err = dns01acmedns.NewService(mgr.childApp, realCfg)

	case *dns01acmesh.Config:
		serv, err = dns01acmesh.NewService(mgr.childApp, realCfg)

	case *dns01cloudflare.Config:
		serv, err = dns01cloudflare.NewService(mgr.childApp, realCfg)

	default:
		// default fail
		return errors.New("cannot create provider service, unsupported provider cfg")
	}

	// common err check
	if err != nil {
		return err
	}

	// all valid, good to add provider to mgr

	// create Provider from service and config
	typeOf, _ := strings.CutPrefix(reflect.TypeOf(cfg).String(), "*")
	typeOf, _ = strings.CutSuffix(typeOf, ".Config")

	p := &provider{
		ID:      len(mgr.pD),
		Tag:     randomness.GenerateInsecureString(10),
		Type:    typeOf,
		Config:  cfg,
		Service: serv,
	}

	// add provider to providers map with empty domains
	mgr.pD[p] = []string{}

	// add each domain
	for _, domain := range domains {
		// add domain to domains map
		mgr.dP[domain] = p

		// append domain to providers map
		mgr.pD[p] = append(mgr.pD[p], domain)
	}

	return nil
}
