package config

import (
	"fmt"

	"go.yaml.in/yaml/v4"
)

type Config struct {
	Version int              `yaml:"version"`
	Groups  map[string]*Group `yaml:"groups"`
}

type Group struct {
	Paths        PathConfigs   `yaml:"paths"`
	Dependencies *Dependencies `yaml:"dependencies,omitempty"`
}

type PathConfigs []PathConfig

type PathConfig struct {
	Dir       string   `yaml:"dir"`
	Exclude   []string `yaml:"exclude"`
	Include   []string `yaml:"include"`
	Recursive bool     `yaml:"recursive"`
}

func (p *PathConfigs) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == yaml.ScalarNode {
		*p = []PathConfig{{Dir: value.Value}}
		return nil
	}

	if value.Kind == yaml.SequenceNode {
		for _, node := range value.Content {
			switch node.Kind {
			case yaml.ScalarNode:
				*p = append(*p, PathConfig{Dir: node.Value})
			case yaml.MappingNode:
				var pc PathConfig
				if err := node.Decode(&pc); err != nil {
					return fmt.Errorf("failed to decode path config: %w", err)
				}
				*p = append(*p, pc)
			default:
				return fmt.Errorf("unexpected node type for path: %v", node.Kind)
			}
		}
		return nil
	}

	if value.Kind == yaml.MappingNode {
		var pc PathConfig
		if err := value.Decode(&pc); err != nil {
			return fmt.Errorf("failed to decode path config: %w", err)
		}
		*p = []PathConfig{pc}
		return nil
	}

	return fmt.Errorf("path must be a string, object, or array")
}

func (p PathConfigs) AsStrings() []string {
	var result []string
	for _, pc := range p {
		result = append(result, pc.Dir)
	}
	return result
}

type Dependencies struct {
	Allow *DependencyRule `yaml:"allow,omitempty"`
	Deny  *DependencyRule `yaml:"deny,omitempty"`
}

type DependencyRule struct {
	Relative    []string `yaml:"relative,omitempty"`
	Groups      []string `yaml:"groups,omitempty"`
	Patterns    []string `yaml:"patterns,omitempty"`
	SubPackages bool     `yaml:"subPackages,omitempty"`
}
