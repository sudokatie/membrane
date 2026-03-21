// Package container manages container lifecycle operations.
package container

import (
	"github.com/sudokatie/membrane/internal/state"
	"github.com/sudokatie/membrane/pkg/oci"
)

// Container represents a container instance.
type Container struct {
	// ID is the unique identifier for the container.
	ID string
	// Bundle is the path to the container bundle.
	Bundle string
	// Spec is the OCI runtime specification.
	Spec *oci.Spec
	// State is the current container state.
	State *state.State
}

// Config holds configuration for container operations.
type Config struct {
	// StateRoot is the root directory for state storage.
	// Default: /run/membrane
	StateRoot string
}

// DefaultConfig returns the default container configuration.
func DefaultConfig() *Config {
	return &Config{
		StateRoot: "/run/membrane",
	}
}

// Manager handles container operations.
type Manager struct {
	config *Config
	store  state.Store
}

// NewManager creates a new container manager.
func NewManager(config *Config) *Manager {
	if config == nil {
		config = DefaultConfig()
	}
	return &Manager{
		config: config,
		store:  state.NewFileStore(config.StateRoot),
	}
}

// Get retrieves a container by ID.
func (m *Manager) Get(id string) (*Container, error) {
	st, err := m.store.Load(id)
	if err != nil {
		return nil, err
	}
	return &Container{
		ID:     st.ID,
		Bundle: st.Bundle,
		State:  st,
	}, nil
}

// List returns all containers.
func (m *Manager) List() ([]*Container, error) {
	states, err := m.store.List()
	if err != nil {
		return nil, err
	}
	containers := make([]*Container, len(states))
	for i, st := range states {
		containers[i] = &Container{
			ID:     st.ID,
			Bundle: st.Bundle,
			State:  st,
		}
	}
	return containers, nil
}
