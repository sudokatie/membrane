// Package volume handles bind mount operations.
package volume

import (
	"fmt"
	"os"
	"path/filepath"
)

// PropagationMode represents mount propagation settings.
type PropagationMode string

const (
	// PropagationPrivate is MS_PRIVATE mount propagation.
	PropagationPrivate PropagationMode = "private"
	// PropagationRPrivate is MS_REC | MS_PRIVATE mount propagation.
	PropagationRPrivate PropagationMode = "rprivate"
	// PropagationSlave is MS_SLAVE mount propagation.
	PropagationSlave PropagationMode = "slave"
	// PropagationRSlave is MS_REC | MS_SLAVE mount propagation.
	PropagationRSlave PropagationMode = "rslave"
	// PropagationShared is MS_SHARED mount propagation.
	PropagationShared PropagationMode = "shared"
	// PropagationRShared is MS_REC | MS_SHARED mount propagation.
	PropagationRShared PropagationMode = "rshared"
)

// ParsePropagation converts a string to a PropagationMode.
func ParsePropagation(s string) (PropagationMode, error) {
	switch s {
	case "", "private":
		return PropagationPrivate, nil
	case "rprivate":
		return PropagationRPrivate, nil
	case "slave":
		return PropagationSlave, nil
	case "rslave":
		return PropagationRSlave, nil
	case "shared":
		return PropagationShared, nil
	case "rshared":
		return PropagationRShared, nil
	default:
		return "", fmt.Errorf("unknown propagation mode: %s", s)
	}
}

// BindMountOptions holds options for a bind mount.
type BindMountOptions struct {
	// Source path on the host.
	Source string
	// Target path in the container.
	Target string
	// ReadOnly makes the mount read-only.
	ReadOnly bool
	// Propagation mode.
	Propagation PropagationMode
	// CreateSource creates the source if it doesn't exist.
	CreateSource bool
	// CreateTarget creates the target mountpoint.
	CreateTarget bool
}

// MountFlags returns the mount flags for this bind mount.
func (o *BindMountOptions) MountFlags() uintptr {
	var flags uintptr
	
	// MS_BIND is required for bind mounts
	flags = 4096 // MS_BIND

	if o.ReadOnly {
		flags |= 1 // MS_RDONLY
	}

	// Add propagation flags
	switch o.Propagation {
	case PropagationPrivate:
		flags |= 262144 // MS_PRIVATE
	case PropagationRPrivate:
		flags |= 262144 | 16384 // MS_PRIVATE | MS_REC
	case PropagationSlave:
		flags |= 524288 // MS_SLAVE
	case PropagationRSlave:
		flags |= 524288 | 16384 // MS_SLAVE | MS_REC
	case PropagationShared:
		flags |= 1048576 // MS_SHARED
	case PropagationRShared:
		flags |= 1048576 | 16384 // MS_SHARED | MS_REC
	}

	return flags
}

// Validate checks if the bind mount options are valid.
func (o *BindMountOptions) Validate() error {
	if o.Source == "" {
		return fmt.Errorf("bind mount source is required")
	}
	if o.Target == "" {
		return fmt.Errorf("bind mount target is required")
	}
	if !filepath.IsAbs(o.Target) {
		return fmt.Errorf("bind mount target must be absolute: %s", o.Target)
	}
	return nil
}

// PrepareBindMount prepares directories for a bind mount.
func PrepareBindMount(opts *BindMountOptions) error {
	if err := opts.Validate(); err != nil {
		return err
	}

	// Create source if requested and it doesn't exist
	if opts.CreateSource {
		info, err := os.Stat(opts.Source)
		if os.IsNotExist(err) {
			if err := os.MkdirAll(opts.Source, 0755); err != nil {
				return fmt.Errorf("create bind source %s: %w", opts.Source, err)
			}
		} else if err != nil {
			return fmt.Errorf("stat bind source %s: %w", opts.Source, err)
		} else if !info.IsDir() {
			// Source exists and is a file - that's fine for file mounts
		}
	} else {
		// Source must exist
		if _, err := os.Stat(opts.Source); err != nil {
			return fmt.Errorf("bind source does not exist: %s", opts.Source)
		}
	}

	return nil
}

// PrepareBindTarget creates the target mountpoint within a rootfs.
func PrepareBindTarget(rootfs string, target string, sourceIsDir bool) error {
	fullTarget := filepath.Join(rootfs, target)
	
	if sourceIsDir {
		if err := os.MkdirAll(fullTarget, 0755); err != nil {
			return fmt.Errorf("create bind target directory %s: %w", target, err)
		}
	} else {
		// Create parent directory and touch file
		parent := filepath.Dir(fullTarget)
		if err := os.MkdirAll(parent, 0755); err != nil {
			return fmt.Errorf("create bind target parent %s: %w", parent, err)
		}
		f, err := os.OpenFile(fullTarget, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("create bind target file %s: %w", target, err)
		}
		f.Close()
	}

	return nil
}

// IsDirectory checks if a path is a directory.
func IsDirectory(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}
