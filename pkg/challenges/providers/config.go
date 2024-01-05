package providers

import (
	"legocerthub-backend/pkg/challenges/providers/dns01acmedns"
	"legocerthub-backend/pkg/challenges/providers/dns01acmesh"
	"legocerthub-backend/pkg/challenges/providers/dns01cloudflare"
	"legocerthub-backend/pkg/challenges/providers/dns01manual"
	"legocerthub-backend/pkg/challenges/providers/http01internal"
)

// provider manager configs
type ConfigManagerHttp01Internal struct {
	Domains                []string `yaml:"domains"`
	*http01internal.Config `yaml:",inline"`
}

type ConfigManagerDns01Manual struct {
	Domains             []string `yaml:"domains"`
	*dns01manual.Config `yaml:",inline"`
}

type ConfigManagerDns01AcmeDns struct {
	Domains              []string `yaml:"domains"`
	*dns01acmedns.Config `yaml:",inline"`
}

type ConfigManagerDns01AcmeSh struct {
	Domains             []string `yaml:"domains"`
	*dns01acmesh.Config `yaml:",inline"`
}

type ConfigManagerDns01Cloudflare struct {
	Domains                 []string `yaml:"domains"`
	*dns01cloudflare.Config `yaml:",inline"`
}

// Config contains configurations for all provider types with domains
type Config struct {
	Http01InternalConfigs  []ConfigManagerHttp01Internal  `yaml:"http_01_internal,omitempty"`
	Dns01ManualConfigs     []ConfigManagerDns01Manual     `yaml:"dns_01_manual,omitempty"`
	Dns01AcmeDnsConfigs    []ConfigManagerDns01AcmeDns    `yaml:"dns_01_acme_dns,omitempty"`
	Dns01AcmeShConfigs     []ConfigManagerDns01AcmeSh     `yaml:"dns_01_acme_sh,omitempty"`
	Dns01CloudflareConfigs []ConfigManagerDns01Cloudflare `yaml:"dns_01_cloudflare,omitempty"`
}

// Len returns the total number of Provider Configs, regardless of type.
func (cfg Config) Len() int {
	return len(cfg.Http01InternalConfigs) +
		len(cfg.Dns01ManualConfigs) +
		len(cfg.Dns01AcmeDnsConfigs) +
		len(cfg.Dns01AcmeShConfigs) +
		len(cfg.Dns01CloudflareConfigs)
}

// managerProviderConfig is a provider config and additional config for
// the manager
type managerProviderConfig struct {
	domains     []string
	providerCfg providerConfig
}

// All returns a slice of manager provider configs
func (cfg Config) All() []managerProviderConfig {
	all := []managerProviderConfig{}
	for _, mgrCfg := range cfg.Dns01AcmeDnsConfigs {
		all = append(all, managerProviderConfig{
			domains:     mgrCfg.Domains,
			providerCfg: mgrCfg.Config,
		})
	}
	for _, mgrCfg := range cfg.Dns01AcmeShConfigs {
		all = append(all, managerProviderConfig{
			domains:     mgrCfg.Domains,
			providerCfg: mgrCfg.Config,
		})
	}
	for _, mgrCfg := range cfg.Dns01CloudflareConfigs {
		all = append(all, managerProviderConfig{
			domains:     mgrCfg.Domains,
			providerCfg: mgrCfg.Config,
		})
	}
	for _, mgrCfg := range cfg.Dns01ManualConfigs {
		all = append(all, managerProviderConfig{
			domains:     mgrCfg.Domains,
			providerCfg: mgrCfg.Config,
		})
	}
	for _, mgrCfg := range cfg.Http01InternalConfigs {
		all = append(all, managerProviderConfig{
			domains:     mgrCfg.Domains,
			providerCfg: mgrCfg.Config,
		})
	}

	return all
}
