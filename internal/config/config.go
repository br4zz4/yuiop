// Package config reads and writes the optional per-machine config.yml.
package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config is the per-machine configuration. In v1 the only key is platform.
type Config struct {
	Platform string `yaml:"platform"`
}

// Load reads the config file. A missing file yields an empty config, not an error.
func Load(explicit string) (*Config, error) {
	path := explicit
	if path == "" {
		path = DefaultPath()
	}
	if path == "" {
		return &Config{}, nil
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &Config{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}
	var cfg Config
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("parse config %s: %w", path, err)
	}
	return &cfg, nil
}

// SetPlatform persists a platform override to the config file.
func SetPlatform(name, explicit string) error {
	path := explicit
	if path == "" {
		path = DefaultPath()
	}
	if path == "" {
		return fmt.Errorf("cannot determine config path")
	}
	data, err := yaml.Marshal(&Config{Platform: name})
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// DefaultPath returns $YUIOP_CONFIG, else $XDG_CONFIG_HOME/yuiop/config.yml,
// else ~/.config/yuiop/config.yml. Empty if no home can be determined.
func DefaultPath() string {
	if p := os.Getenv("YUIOP_CONFIG"); p != "" {
		return p
	}
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "yuiop", "config.yml")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "yuiop", "config.yml")
}
