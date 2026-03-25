package container

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/sudokatie/membrane/internal/capabilities"
	"github.com/sudokatie/membrane/internal/cgroup"
	"github.com/sudokatie/membrane/internal/filesystem"
	"github.com/sudokatie/membrane/internal/hooks"
	"github.com/sudokatie/membrane/internal/log"
	"github.com/sudokatie/membrane/internal/namespace"
	"github.com/sudokatie/membrane/internal/seccomp"
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
	logger := log.WithField("container", opts.ID)
	logger.Debug("starting container")

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

	// Create hook state for lifecycle hooks
	hookState := &hooks.HookState{
		OCIVersion:  containerSpec.Version,
		ID:          opts.ID,
		Status:      st.Status,
		Bundle:      st.Bundle,
		Annotations: st.Annotations,
	}

	// Run createRuntime hooks (before environment setup)
	if err := hooks.RunCreateRuntime(containerSpec.Hooks, hookState); err != nil {
		return fmt.Errorf("createRuntime hooks: %w", err)
	}

	// Set up cgroup
	cgroupConfig := cgroup.DefaultConfig(opts.ID)
	if containerSpec.Linux != nil && containerSpec.Linux.Resources != nil {
		cgroupConfig.Resources = cgroup.FromSpec(containerSpec.Linux.Resources)
	}
	cgroupMgr := cgroup.NewV2Manager(cgroupConfig)
	if err := cgroupMgr.Create(); err != nil {
		logger.Warnf("cgroup create failed (may be unavailable): %v", err)
	}

	// Set up namespace config
	nsConfig := namespace.DefaultConfig()
	if containerSpec.Linux != nil {
		nsConfig = namespace.FromSpec(containerSpec.Linux)
	}
	nsConfig.SortForClone()

	// Check if user namespace is enabled
	hasUserNS := nsConfig.HasUserNamespace()

	// Set up terminal if requested
	var terminal *Terminal
	var terminalMu sync.Mutex
	if containerSpec.Process != nil && containerSpec.Process.Terminal {
		var err error
		terminal, err = NewTerminal()
		if err != nil {
			return fmt.Errorf("create terminal: %w", err)
		}
		// Terminal cleanup handled after process exits, not in defer
	}

	// Fork child process
	logger.Debug("forking child process")
	pid, err := namespace.CloneChild(nsConfig)
	if err != nil {
		if terminal != nil {
			terminal.Close()
		}
		return fmt.Errorf("clone: %w", err)
	}

	if pid == 0 {
		// In child process
		terminalMu.Lock()
		childTerminal := terminal
		terminalMu.Unlock()

		if err := m.initChild(containerSpec, st.Bundle, childTerminal, hookState); err != nil {
			fmt.Fprintf(os.Stderr, "init error: %v\n", err)
			os.Exit(1)
		}
		// initChild calls exec, so we shouldn't reach here
		os.Exit(0)
	}

	// In parent process
	logger.WithField("pid", pid).Debug("child process started")

	// Write UID/GID mappings for user namespace
	if hasUserNS && containerSpec.Linux != nil {
		if len(containerSpec.Linux.UIDMappings) > 0 {
			logger.Debug("writing UID mappings")
			if err := namespace.WriteUIDMapping(pid, containerSpec.Linux.UIDMappings); err != nil {
				logger.Warnf("write UID mapping failed: %v", err)
			}
		}
		if len(containerSpec.Linux.GIDMappings) > 0 {
			logger.Debug("writing GID mappings")
			if err := namespace.WriteGIDMapping(pid, containerSpec.Linux.GIDMappings); err != nil {
				logger.Warnf("write GID mapping failed: %v", err)
			}
		}
	}

	// Add child to cgroup
	if err := cgroupMgr.AddProcess(pid); err != nil {
		logger.Warnf("add process to cgroup failed: %v", err)
	}

	// Update hook state with PID
	hookState.Pid = pid
	hookState.Status = state.StatusRunning

	// Run poststart hooks
	if err := hooks.RunPoststart(containerSpec.Hooks, hookState); err != nil {
		logger.Warnf("poststart hooks failed: %v", err)
		// Non-fatal per OCI spec
	}

	// Update state
	st.Status = state.StatusRunning
	st.Pid = pid
	if err := m.store.Save(st); err != nil {
		return fmt.Errorf("save state: %w", err)
	}

	logger.Info("container started")
	return nil
}

// initChild runs in the child process after fork.
func (m *Manager) initChild(containerSpec *oci.Spec, bundle string, terminal *Terminal, hookState *hooks.HookState) error {
	// Close extra file descriptors to prevent leaks
	if err := closeExtraFDs(); err != nil {
		log.Warnf("close extra FDs failed: %v", err)
	}

	// Set up terminal if provided
	if terminal != nil {
		if err := terminal.SetupChildTerminal(); err != nil {
			return fmt.Errorf("setup terminal: %w", err)
		}
	}

	// Set up hostname
	if containerSpec.Hostname != "" {
		if err := namespace.SetHostname(containerSpec.Hostname); err != nil {
			return fmt.Errorf("set hostname: %w", err)
		}
	}

	// Apply sysctl settings (must be done before pivot_root in some cases)
	if containerSpec.Linux != nil && len(containerSpec.Linux.Sysctl) > 0 {
		if err := applySysctls(containerSpec.Linux.Sysctl); err != nil {
			return fmt.Errorf("apply sysctls: %w", err)
		}
	}

	// Resolve rootfs path
	rootfs := containerSpec.Root.Path
	if !filepath.IsAbs(rootfs) {
		rootfs = filepath.Join(bundle, rootfs)
	}

	// Build mount list from spec
	mounts := filesystem.DefaultMounts()
	if containerSpec.Mounts != nil {
		specMounts := filesystem.FromSpec(containerSpec, bundle)
		mounts = append(mounts, specMounts.Mounts...)
	}

	// Set up filesystem and pivot_root
	if err := filesystem.SetupRootfs(rootfs, mounts); err != nil {
		return fmt.Errorf("setup rootfs: %w", err)
	}

	// Run createContainer hooks (after pivot_root, before user process)
	if err := hooks.RunCreateContainer(containerSpec.Hooks, hookState); err != nil {
		return fmt.Errorf("createContainer hooks: %w", err)
	}

	// Apply root readonly if specified
	if containerSpec.Root != nil && containerSpec.Root.Readonly {
		if err := filesystem.ReadonlyPath("/"); err != nil {
			return fmt.Errorf("make root readonly: %w", err)
		}
	}

	// Apply masked paths
	if containerSpec.Linux != nil && len(containerSpec.Linux.MaskedPaths) > 0 {
		if err := filesystem.MaskPaths(containerSpec.Linux.MaskedPaths); err != nil {
			return fmt.Errorf("mask paths: %w", err)
		}
	}

	// Apply readonly paths
	if containerSpec.Linux != nil && len(containerSpec.Linux.ReadonlyPaths) > 0 {
		if err := filesystem.ReadonlyPaths(containerSpec.Linux.ReadonlyPaths); err != nil {
			return fmt.Errorf("readonly paths: %w", err)
		}
	}

	// Create devices from spec
	if containerSpec.Linux != nil && len(containerSpec.Linux.Devices) > 0 {
		if err := createDevicesFromSpec(containerSpec.Linux.Devices); err != nil {
			return fmt.Errorf("create devices: %w", err)
		}
	}

	// Apply rlimits
	if containerSpec.Process != nil && len(containerSpec.Process.Rlimits) > 0 {
		if err := applyRlimits(containerSpec.Process.Rlimits); err != nil {
			return fmt.Errorf("apply rlimits: %w", err)
		}
	}

	// Set no_new_privs if requested
	if containerSpec.Process != nil && containerSpec.Process.NoNewPrivileges {
		if err := capabilities.SetNoNewPrivs(); err != nil {
			return fmt.Errorf("set no_new_privs: %w", err)
		}
	}

	// Apply AppArmor profile
	if containerSpec.Process != nil && containerSpec.Process.ApparmorProfile != "" {
		if err := applyAppArmorProfile(containerSpec.Process.ApparmorProfile); err != nil {
			return fmt.Errorf("apply apparmor profile: %w", err)
		}
	}

	// Apply SELinux label
	if containerSpec.Process != nil && containerSpec.Process.SelinuxLabel != "" {
		if err := applySELinuxLabel(containerSpec.Process.SelinuxLabel); err != nil {
			return fmt.Errorf("apply selinux label: %w", err)
		}
	}

	// Apply capabilities
	if containerSpec.Process != nil && containerSpec.Process.Capabilities != nil {
		capConfig := capabilities.FromSpec(containerSpec.Process.Capabilities)
		if err := capabilities.Apply(capConfig); err != nil {
			return fmt.Errorf("apply capabilities: %w", err)
		}
	}

	// Change to working directory
	if containerSpec.Process != nil && containerSpec.Process.Cwd != "" {
		if err := os.Chdir(containerSpec.Process.Cwd); err != nil {
			return fmt.Errorf("chdir: %w", err)
		}
	}

	// Run startContainer hooks (just before exec)
	if err := hooks.RunStartContainer(containerSpec.Hooks, hookState); err != nil {
		return fmt.Errorf("startContainer hooks: %w", err)
	}

	// Run prestart hooks (deprecated but still supported)
	if err := hooks.RunPrestart(containerSpec.Hooks, hookState); err != nil {
		return fmt.Errorf("prestart hooks: %w", err)
	}

	// Apply seccomp filter (must be last before exec)
	if containerSpec.Linux != nil && containerSpec.Linux.Seccomp != nil {
		profile := seccomp.FromSpec(containerSpec.Linux.Seccomp)
		if err := seccomp.LoadFilter(profile); err != nil {
			return fmt.Errorf("load seccomp filter: %w", err)
		}
	}

	// Execute the container process
	return m.execProcess(containerSpec.Process)
}

// createDevicesFromSpec creates device nodes from the OCI spec.
func createDevicesFromSpec(devices []oci.LinuxDevice) error {
	for _, dev := range devices {
		if err := createDevice(dev); err != nil {
			return err
		}
	}
	return nil
}

// createDevice creates a single device node.
func createDevice(dev oci.LinuxDevice) error {
	// Ensure parent directory exists
	dir := filepath.Dir(dev.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create device dir %s: %w", dir, err)
	}

	// Determine device mode
	var mode uint32 = 0666
	if dev.FileMode != nil {
		mode = *dev.FileMode
	}

	// Add device type to mode
	switch dev.Type {
	case "c", "u": // character device
		mode |= 0020000 // S_IFCHR
	case "b": // block device
		mode |= 0060000 // S_IFBLK
	case "p": // FIFO
		mode |= 0010000 // S_IFIFO
	default:
		return fmt.Errorf("unknown device type: %s", dev.Type)
	}

	// Create the device using mknod
	return createDeviceNode(dev.Path, mode, dev.Major, dev.Minor, dev.UID, dev.GID)
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
