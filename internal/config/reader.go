package config

import (
	"fmt"
	"os"

	"go.yaml.in/yaml/v4"
)

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	if err := validate(&cfg); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

func validate(cfg *Config) error {
	if cfg.Version < 1 {
		return fmt.Errorf("invalid version: %d", cfg.Version)
	}

	for name, grp := range cfg.Groups {
		if len(grp.Paths) == 0 {
			return fmt.Errorf("group %s has no paths configured", name)
		}

		for i, pc := range grp.Paths {
			if pc.Dir == "" {
				return fmt.Errorf("group %s paths[%d] has empty directory", name, i)
			}
		}
	}

	return nil
}

func LoadFromBytes(data []byte) (*Config, error) {
	var cfg Config
	err := yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	if err := validate(&cfg); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}