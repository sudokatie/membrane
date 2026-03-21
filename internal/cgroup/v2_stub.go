//go:build !linux

package cgroup

import "errors"

var errNotLinux = errors.New("cgroups require Linux")

// V2Manager stub for non-Linux systems.
type V2Manager struct {
	config *Config
	path   string
}

// NewV2Manager creates a new cgroups v2 manager stub.
func NewV2Manager(config *Config) *V2Manager {
	path := "/sys/fs/cgroup" + config.Parent + "/" + config.Name
	return &V2Manager{
		config: config,
		path:   path,
	}
}

// Create is not supported on non-Linux systems.
func (m *V2Manager) Create() error {
	return errNotLinux
}

// AddProcess is not supported on non-Linux systems.
func (m *V2Manager) AddProcess(pid int) error {
	return errNotLinux
}

// SetResources is not supported on non-Linux systems.
func (m *V2Manager) SetResources(resources *Resources) error {
	return errNotLinux
}

// GetResources is not supported on non-Linux systems.
func (m *V2Manager) GetResources() (*Resources, error) {
	return nil, errNotLinux
}

// Delete is not supported on non-Linux systems.
func (m *V2Manager) Delete() error {
	return errNotLinux
}

// Path returns the cgroup path.
func (m *V2Manager) Path() string {
	return m.path
}

// Exists returns false on non-Linux systems.
func (m *V2Manager) Exists() bool {
	return false
}

// GetPids is not supported on non-Linux systems.
func (m *V2Manager) GetPids() ([]int, error) {
	return nil, errNotLinux
}

// IsCgroupV2 returns false on non-Linux systems.
func IsCgroupV2() bool {
	return false
}
