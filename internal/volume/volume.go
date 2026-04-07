// Package volume handles container volume management.
package volume

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// VolumeDriver identifies the volume driver type.
type VolumeDriver string

const (
	// DriverLocal is the local filesystem driver.
	DriverLocal VolumeDriver = "local"
)

// Volume represents a named volume.
type Volume struct {
	// Name is the unique volume name.
	Name string `json:"name"`
	// Driver is the volume driver.
	Driver VolumeDriver `json:"driver"`
	// Mountpoint is the path on the host.
	Mountpoint string `json:"mountpoint"`
	// CreatedAt is when the volume was created.
	CreatedAt time.Time `json:"created_at"`
	// Labels are user-defined labels.
	Labels map[string]string `json:"labels,omitempty"`
	// Options are driver-specific options.
	Options map[string]string `json:"options,omitempty"`
}

// BindMount represents a bind mount configuration.
type BindMount struct {
	// Source is the host path.
	Source string
	// Target is the container path.
	Target string
	// ReadOnly makes the mount read-only.
	ReadOnly bool
	// Propagation is the mount propagation mode.
	Propagation string
}

// Validate checks if the bind mount configuration is valid.
func (b *BindMount) Validate() error {
	if b.Source == "" {
		return fmt.Errorf("bind mount source is required")
	}
	if b.Target == "" {
		return fmt.Errorf("bind mount target is required")
	}
	if !filepath.IsAbs(b.Target) {
		return fmt.Errorf("bind mount target must be absolute: %s", b.Target)
	}
	return nil
}

// VolumeMount represents a volume mount configuration.
type VolumeMount struct {
	// Name is the volume name.
	Name string
	// Target is the container path.
	Target string
	// ReadOnly makes the mount read-only.
	ReadOnly bool
}

// Validate checks if the volume mount configuration is valid.
func (v *VolumeMount) Validate() error {
	if v.Name == "" {
		return fmt.Errorf("volume name is required")
	}
	if v.Target == "" {
		return fmt.Errorf("volume target is required")
	}
	if !filepath.IsAbs(v.Target) {
		return fmt.Errorf("volume target must be absolute: %s", v.Target)
	}
	return nil
}

// Manager manages named volumes.
type Manager struct {
	// Root is the directory where volumes are stored.
	Root string
	// Volumes maps volume name to volume.
	volumes map[string]*Volume
	mu      sync.RWMutex
}

// NewManager creates a new volume manager.
func NewManager(root string) (*Manager, error) {
	if err := os.MkdirAll(root, 0700); err != nil {
		return nil, fmt.Errorf("create volume root: %w", err)
	}

	m := &Manager{
		Root:    root,
		volumes: make(map[string]*Volume),
	}

	// Load existing volumes
	if err := m.load(); err != nil {
		return nil, err
	}

	return m, nil
}

// load reads volumes from disk.
func (m *Manager) load() error {
	metaPath := filepath.Join(m.Root, "volumes.json")
	data, err := os.ReadFile(metaPath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read volumes metadata: %w", err)
	}

	var volumes []*Volume
	if err := json.Unmarshal(data, &volumes); err != nil {
		return fmt.Errorf("parse volumes metadata: %w", err)
	}

	for _, v := range volumes {
		m.volumes[v.Name] = v
	}
	return nil
}

// save writes volumes to disk.
func (m *Manager) save() error {
	volumes := make([]*Volume, 0, len(m.volumes))
	for _, v := range m.volumes {
		volumes = append(volumes, v)
	}

	data, err := json.MarshalIndent(volumes, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal volumes: %w", err)
	}

	metaPath := filepath.Join(m.Root, "volumes.json")
	if err := os.WriteFile(metaPath, data, 0600); err != nil {
		return fmt.Errorf("write volumes metadata: %w", err)
	}
	return nil
}

// Create creates a new named volume.
func (m *Manager) Create(name string, opts map[string]string) (*Volume, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.volumes[name]; exists {
		return nil, fmt.Errorf("volume %q already exists", name)
	}

	mountpoint := filepath.Join(m.Root, name, "_data")
	if err := os.MkdirAll(mountpoint, 0755); err != nil {
		return nil, fmt.Errorf("create volume directory: %w", err)
	}

	vol := &Volume{
		Name:       name,
		Driver:     DriverLocal,
		Mountpoint: mountpoint,
		CreatedAt:  time.Now(),
		Options:    opts,
	}

	m.volumes[name] = vol
	if err := m.save(); err != nil {
		return nil, err
	}

	return vol, nil
}

// Get retrieves a volume by name.
func (m *Manager) Get(name string) (*Volume, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	vol, exists := m.volumes[name]
	if !exists {
		return nil, fmt.Errorf("volume %q not found", name)
	}
	return vol, nil
}

// GetOrCreate gets an existing volume or creates it.
func (m *Manager) GetOrCreate(name string, opts map[string]string) (*Volume, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if vol, exists := m.volumes[name]; exists {
		return vol, nil
	}

	mountpoint := filepath.Join(m.Root, name, "_data")
	if err := os.MkdirAll(mountpoint, 0755); err != nil {
		return nil, fmt.Errorf("create volume directory: %w", err)
	}

	vol := &Volume{
		Name:       name,
		Driver:     DriverLocal,
		Mountpoint: mountpoint,
		CreatedAt:  time.Now(),
		Options:    opts,
	}

	m.volumes[name] = vol
	if err := m.save(); err != nil {
		return nil, err
	}

	return vol, nil
}

// Remove removes a named volume.
func (m *Manager) Remove(name string, force bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	vol, exists := m.volumes[name]
	if !exists {
		return fmt.Errorf("volume %q not found", name)
	}

	// Remove volume directory
	volDir := filepath.Join(m.Root, name)
	if err := os.RemoveAll(volDir); err != nil && !force {
		return fmt.Errorf("remove volume directory: %w", err)
	}

	delete(m.volumes, name)
	_ = vol // silence unused warning

	return m.save()
}

// List returns all volumes.
func (m *Manager) List() []*Volume {
	m.mu.RLock()
	defer m.mu.RUnlock()

	volumes := make([]*Volume, 0, len(m.volumes))
	for _, v := range m.volumes {
		volumes = append(volumes, v)
	}
	return volumes
}

// Prune removes unused volumes.
func (m *Manager) Prune() ([]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// For now, just return empty - would need container tracking
	// to know which volumes are in use
	return nil, nil
}
