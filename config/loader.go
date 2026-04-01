package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// LoadConfigFile loads configuration from a file into cfg.
// The format is auto-detected from the file extension:
//   - .json       → JSON
//   - .yaml/.yml  → YAML
func LoadConfigFile(path string, cfg interface{}) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open config file: %w", err)
	}
	defer func() { _ = file.Close() }()

	switch strings.ToLower(filepath.Ext(path)) {
	case ".json":
		if err := json.NewDecoder(file).Decode(cfg); err != nil {
			return fmt.Errorf("decode JSON config: %w", err)
		}
	case ".yaml", ".yml":
		if err := yaml.NewDecoder(file).Decode(cfg); err != nil {
			return fmt.Errorf("decode YAML config: %w", err)
		}
	default:
		return fmt.Errorf("unsupported config file format: %s", filepath.Ext(path))
	}

	return nil
}

// LoadJSON loads configuration specifically from a JSON file into cfg.
func LoadJSON(path string, cfg interface{}) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open JSON config file: %w", err)
	}
	defer func() { _ = file.Close() }()

	if err := json.NewDecoder(file).Decode(cfg); err != nil {
		return fmt.Errorf("decode JSON config: %w", err)
	}
	return nil
}

// LoadYAML loads configuration specifically from a YAML file into cfg.
func LoadYAML(path string, cfg interface{}) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open YAML config file: %w", err)
	}
	defer func() { _ = file.Close() }()

	if err := yaml.NewDecoder(file).Decode(cfg); err != nil {
		return fmt.Errorf("decode YAML config: %w", err)
	}
	return nil
}
