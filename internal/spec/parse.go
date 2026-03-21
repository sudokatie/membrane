// Package spec handles OCI specification parsing and validation.
package spec

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sudokatie/membrane/pkg/oci"
)

var (
	// ErrBundleNotFound is returned when the bundle directory doesn't exist.
	ErrBundleNotFound = errors.New("bundle not found")
	// ErrConfigNotFound is returned when config.json doesn't exist.
	ErrConfigNotFound = errors.New("config.json not found")
	// ErrInvalidSpec is returned when the spec is invalid.
	ErrInvalidSpec = errors.New("invalid spec")
)

// LoadSpec loads and parses an OCI spec from a bundle directory.
func LoadSpec(bundlePath string) (*oci.Spec, error) {
	// Check bundle exists
	info, err := os.Stat(bundlePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: %s", ErrBundleNotFound, bundlePath)
		}
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%w: %s is not a directory", ErrBundleNotFound, bundlePath)
	}

	// Load config.json
	configPath := filepath.Join(bundlePath, "config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: %s", ErrConfigNotFound, configPath)
		}
		return nil, err
	}

	// Parse JSON
	var spec oci.Spec
	if err := json.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidSpec, err)
	}

	// Validate required fields
	if err := ValidateSpec(&spec); err != nil {
		return nil, err
	}

	return &spec, nil
}

// ValidateSpec validates an OCI spec has required fields.
func ValidateSpec(spec *oci.Spec) error {
	if spec.Version == "" {
		return fmt.Errorf("%w: ociVersion is required", ErrInvalidSpec)
	}
	if spec.Root == nil {
		return fmt.Errorf("%w: root is required", ErrInvalidSpec)
	}
	if spec.Root.Path == "" {
		return fmt.Errorf("%w: root.path is required", ErrInvalidSpec)
	}
	if spec.Process == nil {
		return fmt.Errorf("%w: process is required", ErrInvalidSpec)
	}
	if len(spec.Process.Args) == 0 {
		return fmt.Errorf("%w: process.args is required", ErrInvalidSpec)
	}
	if spec.Process.Cwd == "" {
		return fmt.Errorf("%w: process.cwd is required", ErrInvalidSpec)
	}
	return nil
}

// LoadSpecFromFile loads a spec directly from a config.json file.
func LoadSpecFromFile(configPath string) (*oci.Spec, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: %s", ErrConfigNotFound, configPath)
		}
		return nil, err
	}

	var spec oci.Spec
	if err := json.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidSpec, err)
	}

	if err := ValidateSpec(&spec); err != nil {
		return nil, err
	}

	return &spec, nil
}

// WriteSpec writes an OCI spec to a bundle directory.
func WriteSpec(bundlePath string, spec *oci.Spec) error {
	if err := os.MkdirAll(bundlePath, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		return err
	}

	configPath := filepath.Join(bundlePath, "config.json")
	return os.WriteFile(configPath, data, 0644)
}

// DefaultSpec returns a minimal default OCI spec.
func DefaultSpec() *oci.Spec {
	return &oci.Spec{
		Version: oci.Version,
		Root: &oci.Root{
			Path:     "rootfs",
			Readonly: false,
		},
		Process: &oci.Process{
			Terminal: false,
			User: oci.User{
				UID: 0,
				GID: 0,
			},
			Args: []string{"/bin/sh"},
			Env: []string{
				"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
				"TERM=xterm",
			},
			Cwd:             "/",
			NoNewPrivileges: true,
		},
		Hostname: "container",
		Mounts: []oci.Mount{
			{
				Destination: "/proc",
				Type:        "proc",
				Source:      "proc",
				Options:     []string{"nosuid", "noexec", "nodev"},
			},
			{
				Destination: "/dev",
				Type:        "tmpfs",
				Source:      "tmpfs",
				Options:     []string{"nosuid", "strictatime", "mode=755", "size=65536k"},
			},
			{
				Destination: "/dev/pts",
				Type:        "devpts",
				Source:      "devpts",
				Options:     []string{"nosuid", "noexec", "newinstance", "ptmxmode=0666", "mode=0620"},
			},
			{
				Destination: "/dev/shm",
				Type:        "tmpfs",
				Source:      "shm",
				Options:     []string{"nosuid", "noexec", "nodev", "mode=1777", "size=65536k"},
			},
			{
				Destination: "/sys",
				Type:        "sysfs",
				Source:      "sysfs",
				Options:     []string{"nosuid", "noexec", "nodev", "ro"},
			},
		},
		Linux: &oci.Linux{
			Namespaces: []oci.LinuxNamespace{
				{Type: oci.PIDNamespace},
				{Type: oci.MountNamespace},
				{Type: oci.IPCNamespace},
				{Type: oci.UTSNamespace},
				{Type: oci.NetworkNamespace},
			},
			MaskedPaths:   oci.DefaultMaskedPaths,
			ReadonlyPaths: oci.DefaultReadonlyPaths,
		},
	}
}
