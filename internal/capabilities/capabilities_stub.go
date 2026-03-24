//go:build !linux

package capabilities

import "fmt"

// Apply is a stub for non-Linux systems.
func Apply(config *Config) error {
	return fmt.Errorf("capabilities not supported on this platform")
}

// SetNoNewPrivs is a stub for non-Linux systems.
func SetNoNewPrivs() error {
	return fmt.Errorf("no_new_privs not supported on this platform")
}

// DropAllCapabilities is a stub for non-Linux systems.
func DropAllCapabilities() error {
	return fmt.Errorf("capabilities not supported on this platform")
}
