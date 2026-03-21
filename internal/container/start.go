package container

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sudokatie/membrane/internal/cgroup"
	"github.com/sudokatie/membrane/internal/namespace"
	"github.com/sudokatie/membrane/internal/spec"
	"github.com/sudokatie/membrane/internal/state"
	"github.com/sudokatie/membrane/pkg/oci"
)

// StartOptions holds options for starting a container.
type StartOptions struct {
	// ID is the container ID.
	ID string
}

// Start starts a created container.
func (m *Manager) Start(opts *StartOptions) error {
	// Load state
	st, err := m.store.Load(opts.ID)
	if err != nil {
		return err
	}

	// Check state
	if !st.CanStart() {
		return fmt.Errorf("cannot start container in %s state", st.Status)
	}

	// Load spec
	containerSpec, err := spec.LoadSpec(st.Bundle)
	if err != nil {
		return fmt.Errorf("load spec: %w", err)
	}

	// Set up cgroup
	cgroupConfig := cgroup.DefaultConfig(opts.ID)
	if containerSpec.Linux != nil && containerSpec.Linux.Resources != nil {
		cgroupConfig.Resources = cgroup.FromSpec(containerSpec.Linux.Resources)
	}
	cgroupMgr := cgroup.NewV2Manager(cgroupConfig)
	if err := cgroupMgr.Create(); err != nil {
		// Non-fatal on non-Linux
	}

	// Set up namespace config
	nsConfig := namespace.DefaultConfig()
	if containerSpec.Linux != nil {
		nsConfig = namespace.FromSpec(containerSpec.Linux)
	}
	nsConfig.SortForClone()

	// Fork child process
	pid, err := namespace.CloneChild(nsConfig)
	if err != nil {
		return fmt.Errorf("clone: %w", err)
	}

	if pid == 0 {
		// In child process
		if err := m.initChild(containerSpec, st.Bundle); err != nil {
			fmt.Fprintf(os.Stderr, "init error: %v\n", err)
			os.Exit(1)
		}
		// initChild calls exec, so we shouldn't reach here
		os.Exit(0)
	}

	// In parent process

	// Add child to cgroup
	if err := cgroupMgr.AddProcess(pid); err != nil {
		// Non-fatal
	}

	// Update state
	st.Status = state.StatusRunning
	st.Pid = pid
	if err := m.store.Save(st); err != nil {
		return fmt.Errorf("save state: %w", err)
	}

	return nil
}

// initChild runs in the child process after fork.
func (m *Manager) initChild(containerSpec *oci.Spec, bundle string) error {
	// Set up hostname
	if containerSpec.Hostname != "" {
		if err := namespace.SetHostname(containerSpec.Hostname); err != nil {
			return fmt.Errorf("set hostname: %w", err)
		}
	}

	// Resolve rootfs path
	rootfs := containerSpec.Root.Path
	if !filepath.IsAbs(rootfs) {
		rootfs = filepath.Join(bundle, rootfs)
	}

	// Set up filesystem and pivot_root
	// This is done in the filesystem package

	// Change to working directory
	if containerSpec.Process != nil && containerSpec.Process.Cwd != "" {
		if err := os.Chdir(containerSpec.Process.Cwd); err != nil {
			return fmt.Errorf("chdir: %w", err)
		}
	}

	// Execute the container process
	return m.execProcess(containerSpec.Process)
}

// Run creates and starts a container in one operation.
func (m *Manager) Run(createOpts *CreateOptions) error {
	// Create container
	container, err := m.Create(createOpts)
	if err != nil {
		return err
	}

	// Start it
	return m.Start(&StartOptions{ID: container.ID})
}
