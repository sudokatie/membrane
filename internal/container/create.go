package container

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/sudokatie/membrane/internal/cgroup"
	"github.com/sudokatie/membrane/internal/spec"
	"github.com/sudokatie/membrane/internal/state"
	"github.com/sudokatie/membrane/pkg/oci"
)

var (
	// ErrInvalidID is returned when the container ID is invalid.
	ErrInvalidID = errors.New("invalid container id")
	// ErrBundleNotFound is returned when the bundle path doesn't exist.
	ErrBundleNotFound = errors.New("bundle not found")
	// ErrContainerExists is returned when a container already exists.
	ErrContainerExists = errors.New("container already exists")
)

// CreateOptions holds options for creating a container.
type CreateOptions struct {
	// ID is the unique identifier for the container.
	ID string
	// Bundle is the path to the container bundle.
	Bundle string
	// Annotations are additional metadata.
	Annotations map[string]string
}

// Create creates a new container from a bundle.
func (m *Manager) Create(opts *CreateOptions) (*Container, error) {
	// Validate ID
	if opts.ID == "" {
		return nil, ErrInvalidID
	}

	// Check container doesn't already exist
	if m.store.Exists(opts.ID) {
		return nil, fmt.Errorf("%w: %s", ErrContainerExists, opts.ID)
	}

	// Resolve bundle path to absolute
	bundle, err := filepath.Abs(opts.Bundle)
	if err != nil {
		return nil, fmt.Errorf("resolve bundle path: %w", err)
	}

	// Check bundle exists and is a directory
	info, err := os.Stat(bundle)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: %s", ErrBundleNotFound, bundle)
		}
		return nil, fmt.Errorf("stat bundle: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%w: %s is not a directory", ErrBundleNotFound, bundle)
	}

	// Load and validate spec
	containerSpec, err := spec.LoadSpec(bundle)
	if err != nil {
		return nil, fmt.Errorf("load spec: %w", err)
	}

	// Check rootfs exists
	rootfs := containerSpec.Root.Path
	if !filepath.IsAbs(rootfs) {
		rootfs = filepath.Join(bundle, rootfs)
	}
	if _, err := os.Stat(rootfs); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("rootfs not found: %s", rootfs)
		}
		return nil, fmt.Errorf("stat rootfs: %w", err)
	}

	// Merge annotations
	annotations := make(map[string]string)
	for k, v := range containerSpec.Annotations {
		annotations[k] = v
	}
	for k, v := range opts.Annotations {
		annotations[k] = v
	}

	// Create cgroup
	cgroupConfig := cgroup.DefaultConfig(opts.ID)
	if containerSpec.Linux != nil && containerSpec.Linux.Resources != nil {
		cgroupConfig.Resources = cgroup.FromSpec(containerSpec.Linux.Resources)
	}
	cgroupMgr := cgroup.NewV2Manager(cgroupConfig)
	if err := cgroupMgr.Create(); err != nil {
		// Log but don't fail - cgroups may not be available
		// (e.g., running in a container without cgroup access)
	}

	// Apply resource limits
	if cgroupConfig.Resources != nil {
		if err := cgroupMgr.SetResources(cgroupConfig.Resources); err != nil {
			// Non-fatal
		}
	}

	// Create state
	st := &state.State{
		Version:     oci.Version,
		ID:          opts.ID,
		Status:      state.StatusCreated,
		Pid:         0,
		Bundle:      bundle,
		Annotations: annotations,
		Created:     time.Now(),
	}

	// Save state
	if err := m.store.Create(st); err != nil {
		// Clean up cgroup on failure
		cgroupMgr.Delete()
		return nil, fmt.Errorf("save state: %w", err)
	}

	return &Container{
		ID:     opts.ID,
		Bundle: bundle,
		Spec:   containerSpec,
		State:  st,
	}, nil
}

// Delete deletes a container.
func (m *Manager) Delete(id string, force bool) error {
	// Load state
	st, err := m.store.Load(id)
	if err != nil {
		return err
	}

	// Check if container can be deleted
	if !force && !st.CanDelete() {
		return fmt.Errorf("cannot delete container in %s state", st.Status)
	}

	// Clean up cgroup
	cgroupConfig := cgroup.DefaultConfig(id)
	cgroupMgr := cgroup.NewV2Manager(cgroupConfig)
	if err := cgroupMgr.Delete(); err != nil {
		// Log but continue - cgroup may not exist or may have been cleaned up
	}

	// Delete state
	if err := m.store.Delete(id); err != nil {
		return fmt.Errorf("delete state: %w", err)
	}

	return nil
}
