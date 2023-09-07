package providers

import (
	"errors"
	"io/fs"
	"legocerthub-backend/pkg/challenges/providers/dns01acmedns"
	"legocerthub-backend/pkg/challenges/providers/dns01acmesh"
	"legocerthub-backend/pkg/challenges/providers/dns01cloudflare"
	"legocerthub-backend/pkg/challenges/providers/dns01manual"
	"legocerthub-backend/pkg/challenges/providers/http01internal"
	"os"

	"gopkg.in/yaml.v3"
)

var errFailedToAssertCfg = errors.New("failed to assert provider config type")

// Config contains configurations for all provider types
type Config struct {
	Http01InternalConfigs  []*http01internal.Config  `yaml:"http_01_internal" json:"http_01_internal"`
	Dns01ManualConfigs     []*dns01manual.Config     `yaml:"dns_01_manual" json:"dns_01_manual"`
	Dns01AcmeDnsConfigs    []*dns01acmedns.Config    `yaml:"dns_01_acme_dns" json:"dns_01_acme_dns"`
	Dns01AcmeShConfigs     []*dns01acmesh.Config     `yaml:"dns_01_acme_sh" json:"dns_01_acme_sh"`
	Dns01CloudflareConfigs []*dns01cloudflare.Config `yaml:"dns_01_cloudflare" json:"dns_01_cloudflare"`
}

// Len returns the total number of Provider Configs, regardless of type.
func (cfg Config) Len() int {
	return len(cfg.Dns01AcmeDnsConfigs) +
		len(cfg.Dns01AcmeShConfigs) +
		len(cfg.Dns01CloudflareConfigs) +
		len(cfg.Dns01ManualConfigs) +
		len(cfg.Http01InternalConfigs)
}

// config returns the providers Config for manager's current configuration
func (mgr *Manager) config() (*Config, error) {
	mgr.mu.RLock()
	defer mgr.mu.RUnlock()

	// fail if not usable
	if !mgr.usable {
		return nil, errMgrUnusable
	}

	cfg := &Config{}

	// add each provider config to Config
	for _, h01i := range mgr.tP["http01internal"] {
		c, ok := h01i.Config.(*http01internal.Config)
		if !ok {
			return nil, errFailedToAssertCfg
		}
		cfg.Http01InternalConfigs = append(cfg.Http01InternalConfigs, c)
	}

	for _, d01m := range mgr.tP["dns01manual"] {
		c, ok := d01m.Config.(*dns01manual.Config)
		if !ok {
			return nil, errFailedToAssertCfg
		}
		cfg.Dns01ManualConfigs = append(cfg.Dns01ManualConfigs, c)
	}

	for _, d01ad := range mgr.tP["dns01acmedns"] {
		c, ok := d01ad.Config.(*dns01acmedns.Config)
		if !ok {
			return nil, errFailedToAssertCfg
		}
		cfg.Dns01AcmeDnsConfigs = append(cfg.Dns01AcmeDnsConfigs, c)
	}

	for _, d01as := range mgr.tP["dns01acmesh"] {
		c, ok := d01as.Config.(*dns01acmesh.Config)
		if !ok {
			return nil, errFailedToAssertCfg
		}
		cfg.Dns01AcmeShConfigs = append(cfg.Dns01AcmeShConfigs, c)
	}

	for _, d01cf := range mgr.tP["dns01cloudflare"] {
		c, ok := d01cf.Config.(*dns01cloudflare.Config)
		if !ok {
			return nil, errFailedToAssertCfg
		}
		cfg.Dns01CloudflareConfigs = append(cfg.Dns01CloudflareConfigs, c)
	}

	return cfg, nil
}

// SaveProvidersConfig saves the current provider configuration to the
// specified filename
func (mgr *Manager) SaveProvidersConfig(filename string) error {
	// get manager current config (will fail of unusable)
	mgrCfg, err := mgr.config()
	if err != nil {
		return err
	}

	// open and unmarshal config
	fCfg, err := os.ReadFile(filename)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return err
		}
	}

	var fullCfgFile yaml.Node
	err = yaml.Unmarshal(fCfg, &fullCfgFile)
	if err != nil {
		return err
	}

	// if no config read in / decoded, make root node
	if fullCfgFile.Kind == 0 {
		fullCfgFile = yaml.Node{
			Kind: yaml.DocumentNode,
			Content: []*yaml.Node{
				{
					Kind: yaml.MappingNode,
					Tag:  "!!map",
				},
			},
		}
	}

	// find challenges content node
	challValIndex := -1
	for i, n := range fullCfgFile.Content[0].Content {
		if n.Value == "challenges" {
			challValIndex = i + 1
			break
		}
	}
	// if no challenges node, create one
	if challValIndex == -1 {
		nameNode := &yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!str",
			Value: "challenges",
		}
		valNode := &yaml.Node{
			Kind: yaml.MappingNode,
			Tag:  "!!map",
		}
		fullCfgFile.Content[0].Content =
			append(fullCfgFile.Content[0].Content, nameNode, valNode)
		// index is now last member, so len -1
		challValIndex = len(fullCfgFile.Content[0].Content) - 1
	}

	// find providers content node
	providersValIndex := -1
	for i, n := range fullCfgFile.Content[0].Content[challValIndex].Content {
		if n.Value == "providers" {
			providersValIndex = i + 1
			break
		}
	}
	// if no providers node, create one
	if providersValIndex == -1 {
		nameNode := &yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!str",
			Value: "providers",
		}
		valNode := &yaml.Node{
			Kind: yaml.MappingNode,
			Tag:  "!!map",
		}
		fullCfgFile.Content[0].Content[challValIndex].Content =
			append(fullCfgFile.Content[0].Content[challValIndex].Content, nameNode, valNode)
		// index is now last member, so len -1
		providersValIndex = len(fullCfgFile.Content[0].Content[challValIndex].Content) - 1
	}

	// set providers to the mgr Config
	newNode := &yaml.Node{}
	newNode.Encode(mgrCfg)
	fullCfgFile.Content[0].Content[challValIndex].Content[providersValIndex] = newNode

	// Marshall new completed config
	newCfg, err := yaml.Marshal(fullCfgFile.Content[0])
	if err != nil {
		return err
	}

	// Write new config to file
	err = os.WriteFile(filename, newCfg, 0600)
	if err != nil {
		return err
	}

	return nil
}
