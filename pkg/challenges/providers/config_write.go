package providers

import (
	"certwarden-backend/pkg/challenges/providers/dns01acmedns"
	"certwarden-backend/pkg/challenges/providers/dns01acmesh"
	"certwarden-backend/pkg/challenges/providers/dns01cloudflare"
	"certwarden-backend/pkg/challenges/providers/dns01goacme"
	"certwarden-backend/pkg/challenges/providers/dns01manual"
	"certwarden-backend/pkg/challenges/providers/http01internal"
	"errors"
	"io/fs"
	"os"

	"gopkg.in/yaml.v3"
)

// unsafeWriteProvidersConfig saves the current provider configuration to the
// specified filename. It MUST be called from a thread that is already at
// minimum RLocked.
func (mgr *Manager) unsafeWriteProvidersConfig() error {
	// get manager config (to write it to file)
	mgrCfg := &Config{}

	// append every provider's config
	for _, p := range mgr.providers {
		switch realCfg := p.Config.(type) {
		case *http01internal.Config:
			mgrCfg.Http01InternalConfigs = append(mgrCfg.Http01InternalConfigs,
				ConfigManagerHttp01Internal{
					InternalConfig: InternalConfig{
						Domains:              p.Domains,
						PreCheckWaitSeconds:  p.PreCheckWaitSeconds,
						PostCheckWaitSeconds: p.PostCheckWaitSeconds,
					},
					Config: realCfg,
				},
			)

		case *dns01manual.Config:
			mgrCfg.Dns01ManualConfigs = append(mgrCfg.Dns01ManualConfigs,
				ConfigManagerDns01Manual{
					InternalConfig: InternalConfig{
						Domains:              p.Domains,
						PreCheckWaitSeconds:  p.PreCheckWaitSeconds,
						PostCheckWaitSeconds: p.PostCheckWaitSeconds,
					},
					Config: realCfg,
				},
			)

		case *dns01acmedns.Config:
			mgrCfg.Dns01AcmeDnsConfigs = append(mgrCfg.Dns01AcmeDnsConfigs,
				ConfigManagerDns01AcmeDns{
					InternalConfig: InternalConfig{
						Domains:              p.Domains,
						PreCheckWaitSeconds:  p.PreCheckWaitSeconds,
						PostCheckWaitSeconds: p.PostCheckWaitSeconds,
					},
					Config: realCfg,
				},
			)

		case *dns01acmesh.Config:
			mgrCfg.Dns01AcmeShConfigs = append(mgrCfg.Dns01AcmeShConfigs,
				ConfigManagerDns01AcmeSh{
					InternalConfig: InternalConfig{
						Domains:              p.Domains,
						PreCheckWaitSeconds:  p.PreCheckWaitSeconds,
						PostCheckWaitSeconds: p.PostCheckWaitSeconds,
					},
					Config: realCfg,
				},
			)

		case *dns01cloudflare.Config:
			mgrCfg.Dns01CloudflareConfigs = append(mgrCfg.Dns01CloudflareConfigs,
				ConfigManagerDns01Cloudflare{
					InternalConfig: InternalConfig{
						Domains:              p.Domains,
						PreCheckWaitSeconds:  p.PreCheckWaitSeconds,
						PostCheckWaitSeconds: p.PostCheckWaitSeconds,
					},
					Config: realCfg,
				},
			)

		case *dns01goacme.Config:
			mgrCfg.Dns01GoAcmeConfigs = append(mgrCfg.Dns01GoAcmeConfigs,
				ConfigManagerDns01GoAcme{
					InternalConfig: InternalConfig{
						Domains:              p.Domains,
						PreCheckWaitSeconds:  p.PreCheckWaitSeconds,
						PostCheckWaitSeconds: p.PostCheckWaitSeconds,
					},
					Config: realCfg,
				},
			)

		default:
			mgr.logger.Errorf("provider mgr couldn't append provider config for provider id %d, report as bug to developer", p.ID)
		}
	}

	// open and unmarshal config file
	fCfg, err := os.ReadFile(mgr.configFile)
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
	err = os.WriteFile(mgr.configFile, newCfg, 0600)
	if err != nil {
		return err
	}

	return nil
}
