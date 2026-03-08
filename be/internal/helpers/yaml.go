package helpers

import (
	"os"

	"gopkg.in/yaml.v3"
)

// LoadYAML loads a YAML file into the provided struct
func LoadYAML[T any](path string) (*T, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config T
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// SaveYAML saves a struct to a YAML file
func SaveYAML[T any](path string, config *T) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
