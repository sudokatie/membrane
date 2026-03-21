// Package filesystem handles mount operations and rootfs setup.
package filesystem

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sudokatie/membrane/pkg/oci"
)

// Mount represents a filesystem mount.
type Mount struct {
	// Source is the mount source (device or path).
	Source string
	// Target is the mount destination.
	Target string
	// FSType is the filesystem type.
	FSType string
	// Flags are mount flags.
	Flags uintptr
	// Data is mount-specific data.
	Data string
}

// MountConfig holds mount configuration for a container.
type MountConfig struct {
	// RootfsPath is the path to the container rootfs.
	RootfsPath string
	// Mounts is the list of mounts to perform.
	Mounts []Mount
}

// FromSpec creates mount config from an OCI spec.
func FromSpec(spec *oci.Spec, bundlePath string) *MountConfig {
	config := &MountConfig{}

	// Resolve rootfs path
	rootfs := spec.Root.Path
	if !filepath.IsAbs(rootfs) {
		rootfs = filepath.Join(bundlePath, rootfs)
	}
	config.RootfsPath = rootfs

	// Convert OCI mounts
	for _, m := range spec.Mounts {
		flags, data := oci.ParseMountFlags(m.Options)
		config.Mounts = append(config.Mounts, Mount{
			Source: m.Source,
			Target: m.Destination,
			FSType: m.Type,
			Flags:  flags,
			Data:   data,
		})
	}

	return config
}

// PrepareRootfs prepares the container rootfs directory.
func PrepareRootfs(rootfs string) error {
	// Ensure rootfs exists
	if _, err := os.Stat(rootfs); err != nil {
		return fmt.Errorf("rootfs not found: %w", err)
	}

	return nil
}

// CreateMountpoint creates a mount point directory.
func CreateMountpoint(rootfs, path string) error {
	target := filepath.Join(rootfs, path)
	if err := os.MkdirAll(target, 0755); err != nil {
		return fmt.Errorf("create mountpoint %s: %w", path, err)
	}
	return nil
}

// DefaultMounts returns the default container mounts.
func DefaultMounts() []Mount {
	return []Mount{
		{
			Source: "proc",
			Target: "/proc",
			FSType: "proc",
			Flags:  MountFlags["nosuid"] | MountFlags["noexec"] | MountFlags["nodev"],
		},
		{
			Source: "tmpfs",
			Target: "/dev",
			FSType: "tmpfs",
			Flags:  MountFlags["nosuid"],
			Data:   "mode=755,size=65536k",
		},
		{
			Source: "devpts",
			Target: "/dev/pts",
			FSType: "devpts",
			Flags:  MountFlags["nosuid"] | MountFlags["noexec"],
			Data:   "newinstance,ptmxmode=0666,mode=0620",
		},
		{
			Source: "shm",
			Target: "/dev/shm",
			FSType: "tmpfs",
			Flags:  MountFlags["nosuid"] | MountFlags["noexec"] | MountFlags["nodev"],
			Data:   "mode=1777,size=65536k",
		},
		{
			Source: "sysfs",
			Target: "/sys",
			FSType: "sysfs",
			Flags:  MountFlags["nosuid"] | MountFlags["noexec"] | MountFlags["nodev"] | MountFlags["ro"],
		},
	}
}

// MountFlags for convenience - uses the stub/linux-specific values.
var MountFlags = map[string]uintptr{
	"ro":          1,          // MS_RDONLY
	"nosuid":      2,          // MS_NOSUID
	"nodev":       4,          // MS_NODEV
	"noexec":      8,          // MS_NOEXEC
	"remount":     32,         // MS_REMOUNT
	"bind":        4096,       // MS_BIND
	"rbind":       4096 | 16384,
	"rec":         16384,      // MS_REC
	"private":     262144,     // MS_PRIVATE
	"rprivate":    262144 | 16384,
	"slave":       524288,     // MS_SLAVE
	"rslave":      524288 | 16384,
}
