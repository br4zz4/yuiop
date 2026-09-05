// Package resolve maps a canonical package name to the per-manager package.
package resolve

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// Table is the curated, embedded mapping of canonical names to provider names.
type Table struct {
	Packages map[string]map[string]string `yaml:"packages"`
}

// Load parses the embedded packages table.
func Load(data []byte) (*Table, error) {
	var t Table
	if err := yaml.Unmarshal(data, &t); err != nil {
		return nil, fmt.Errorf("parse packages table: %w", err)
	}
	return &t, nil
}

// Resolve returns the provider-specific package name for a canonical name.
// The second return value is false when the canonical is unknown or has no
// mapping for the given provider.
func (t *Table) Resolve(canonical, provider string) (string, bool) {
	m, ok := t.Packages[canonical]
	if !ok {
		return "", false
	}
	pkg, ok := m[provider]
	if !ok {
		return "", false
	}
	return pkg, true
}
