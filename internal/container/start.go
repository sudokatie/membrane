package container

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sudokatie/membrane/internal/capabilities"
	"github.com/sudokatie/membrane/internal/cgroup"
	"github.com/sudokatie/membrane/internal/filesystem"
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

	// Set up terminal if requested
	var terminal *Terminal
	if containerSpec.Process != nil && containerSpec.Process.Terminal {
		var err error
		terminal, err = NewTerminal()
		if err != nil {
			return fmt.Errorf("create terminal: %w", err)
		}
		defer func() {
			if terminal != nil {
				terminal.Close()
			}
		}()
	}

	// Fork child process
	pid, err := namespace.CloneChild(nsConfig)
	if err != nil {
		return fmt.Errorf("clone: %w", err)
	}

	if pid == 0 {
		// In child process
		if err := m.initChild(containerSpec, st.Bundle, terminal); err != nil {
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
func (m *Manager) initChild(containerSpec *oci.Spec, bundle string, terminal *Terminal) error {
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

	// Apply seccomp filter (must be last before exec)
	if containerSpec.Linux != nil && containerSpec.Linux.Seccomp != nil {
		profile := seccomp.FromSpec(containerSpec.Linux.Seccomp)
		if err := seccomp.LoadFilter(profile); err != nil {
			return fmt.Errorf("load seccomp filter: %w", err)
		}
	}

	// Change to working directory
	if containerSpec.Process != nil && containerSpec.Process.Cwd != "" {
		if err := os.Chdir(containerSpec.Process.Cwd); err != nil {
			return fmt.Errorf("chdir: %w", err)
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
