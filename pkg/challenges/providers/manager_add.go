package providers

import (
	"errors"
	"legocerthub-backend/pkg/challenges/providers/dns01acmedns"
	"legocerthub-backend/pkg/challenges/providers/dns01acmesh"
	"legocerthub-backend/pkg/challenges/providers/dns01cloudflare"
	"legocerthub-backend/pkg/challenges/providers/dns01manual"
	"legocerthub-backend/pkg/challenges/providers/http01internal"
	"legocerthub-backend/pkg/randomness"
	"reflect"
	"strings"
)

// unsafeAddProvider creates the provider specified in cfg and adds it to
// manager. It MUST be called from a Locked state OR during initial Manager
// creation which is single threaded (and thus safe)
func (mgr *Manager) unsafeAddProvider(domains []string, cfg providerConfig) (*provider, error) {
	// verify every domain ir properly formatted, or verify this is wildcard cfg (* only)
	// and also verify all domains are available in manager
	err := mgr.unsafeValidateDomains(domains, nil)
	if err != nil {
		return nil, err
	}

	// make provider service (switch based on cfg type (and thus which pkg to use))
	var serv Service

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
		return nil, errors.New("cannot create provider service, unsupported provider cfg")
	}

	// common err check
	if err != nil {
		return nil, err
	}

	// all valid, good to add provider to mgr

	// create Provider from service and config
	typeOf, _ := strings.CutPrefix(reflect.TypeOf(cfg).String(), "*")
	typeOf, _ = strings.CutSuffix(typeOf, ".Config")

	p := &provider{
		ID:      mgr.nextId,
		Tag:     randomness.GenerateInsecureString(10),
		Domains: domains,
		Type:    typeOf,
		Config:  cfg,
		Service: serv,
	}

	// increment next id
	mgr.nextId++

	// add provider to provider slice
	mgr.providers = append(mgr.providers, p)

	// add each domain to domain map
	for _, domain := range domains {
		// add domain to domains map
		mgr.dP[domain] = p
	}

	return p, nil
}
