package challenges

import (
	"errors"
	"io/fs"
	"os"

	"gopkg.in/yaml.v3"
)

// writeAliasConfig writes the current domain alias safe map to the config file
func (service *Service) writeAliasConfig() error {
	// open and unmarshal config file
	fCfg, err := os.ReadFile(service.configFile)
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

	// find `domain_aliases` content node
	aliasValIndex := -1
	for i, n := range fullCfgFile.Content[0].Content[challValIndex].Content {
		if n.Value == "domain_aliases" {
			aliasValIndex = i + 1
			break
		}
	}
	// if no `domain_aliases` node, create one
	if aliasValIndex == -1 {
		nameNode := &yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!str",
			Value: "domain_aliases",
		}
		valNode := &yaml.Node{
			Kind: yaml.MappingNode,
			Tag:  "!!map",
		}
		fullCfgFile.Content[0].Content[challValIndex].Content =
			append(fullCfgFile.Content[0].Content[challValIndex].Content, nameNode, valNode)
		// index is now last member, so len -1
		aliasValIndex = len(fullCfgFile.Content[0].Content[challValIndex].Content) - 1
	}

	// set the map value from service
	m := make(map[string]string)
	service.dnsIDtoDomain.CopyToMap(m)

	newNode := &yaml.Node{}
	newNode.Encode(m)
	fullCfgFile.Content[0].Content[challValIndex].Content[aliasValIndex] = newNode

	// Marshall new completed config
	newCfg, err := yaml.Marshal(fullCfgFile.Content[0])
	if err != nil {
		return err
	}

	// Write new config to file
	err = os.WriteFile(service.configFile, newCfg, 0600)
	if err != nil {
		return err
	}

	return nil
}
