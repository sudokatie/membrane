// Package namespace handles Linux namespace configuration and setup.
package namespace

import (
	"fmt"

	"github.com/sudokatie/membrane/pkg/oci"
)

// Config holds namespace configuration for a container.
type Config struct {
	// Namespaces is the list of namespaces to create or join.
	Namespaces []Namespace
}

// Namespace represents a single namespace configuration.
type Namespace struct {
	// Type is the namespace type.
	Type oci.NamespaceType
	// Path is the path to an existing namespace to join.
	// If empty, a new namespace is created.
	Path string
}

// FromSpec creates a namespace config from an OCI spec.
func FromSpec(linux *oci.Linux) *Config {
	if linux == nil {
		return &Config{}
	}

	config := &Config{
		Namespaces: make([]Namespace, len(linux.Namespaces)),
	}

	for i, ns := range linux.Namespaces {
		config.Namespaces[i] = Namespace{
			Type: ns.Type,
			Path: ns.Path,
		}
	}

	return config
}

// CloneFlags returns the combined clone flags for all namespaces.
func (c *Config) CloneFlags() uintptr {
	var flags uintptr
	for _, ns := range c.Namespaces {
		if ns.Path == "" {
			// Only add flag if creating new namespace (not joining existing)
			if f, ok := oci.CloneFlags[ns.Type]; ok {
				flags |= f
			}
		}
	}
	return flags
}

// HasNamespace returns true if the config includes the specified namespace type.
func (c *Config) HasNamespace(t oci.NamespaceType) bool {
	for _, ns := range c.Namespaces {
		if ns.Type == t {
			return true
		}
	}
	return false
}

// GetNamespace returns the namespace config for the specified type.
func (c *Config) GetNamespace(t oci.NamespaceType) *Namespace {
	for i := range c.Namespaces {
		if c.Namespaces[i].Type == t {
			return &c.Namespaces[i]
		}
	}
	return nil
}

// HasUserNamespace returns true if a user namespace is configured.
func (c *Config) HasUserNamespace() bool {
	return c.HasNamespace(oci.UserNamespace)
}

// Validate checks that the namespace configuration is valid.
func (c *Config) Validate() error {
	// Check for duplicate namespace types
	seen := make(map[oci.NamespaceType]bool)
	for _, ns := range c.Namespaces {
		if seen[ns.Type] {
			return fmt.Errorf("duplicate namespace type: %s", ns.Type)
		}
		seen[ns.Type] = true
	}

	// USER namespace should be first if present (for proper privilege handling)
	if c.HasUserNamespace() {
		if len(c.Namespaces) > 0 && c.Namespaces[0].Type != oci.UserNamespace {
			return fmt.Errorf("user namespace must be first in namespace list")
		}
	}

	return nil
}

// SortForClone reorders namespaces for proper clone ordering.
// USER namespace must be first if present.
func (c *Config) SortForClone() {
	if !c.HasUserNamespace() {
		return
	}

	// Find user namespace and move to front
	for i, ns := range c.Namespaces {
		if ns.Type == oci.UserNamespace && i > 0 {
			// Swap with first element
			c.Namespaces[0], c.Namespaces[i] = c.Namespaces[i], c.Namespaces[0]
			break
		}
	}
}

// DefaultConfig returns a default namespace configuration.
func DefaultConfig() *Config {
	return &Config{
		Namespaces: []Namespace{
			{Type: oci.PIDNamespace},
			{Type: oci.MountNamespace},
			{Type: oci.IPCNamespace},
			{Type: oci.UTSNamespace},
			{Type: oci.NetworkNamespace},
			{Type: oci.CgroupNamespace},
		},
	}
}
