//go:build linux

package capabilities

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/unix"
)

// Apply applies the capability configuration to the current process.
func Apply(config *Config) error {
	if config == nil {
		return nil
	}

	// Get current capabilities
	var hdr unix.CapUserHeader
	var data [2]unix.CapUserData

	hdr.Version = unix.LINUX_CAPABILITY_VERSION_3
	hdr.Pid = 0 // current process

	if err := unix.Capget(&hdr, &data[0]); err != nil {
		return fmt.Errorf("capget: %w", err)
	}

	// Calculate the capability sets
	boundingSet, err := ToBitset(config.Bounding)
	if err != nil {
		return fmt.Errorf("parse bounding set: %w", err)
	}
	effectiveSet, err := ToBitset(config.Effective)
	if err != nil {
		return fmt.Errorf("parse effective set: %w", err)
	}
	inheritableSet, err := ToBitset(config.Inheritable)
	if err != nil {
		return fmt.Errorf("parse inheritable set: %w", err)
	}
	permittedSet, err := ToBitset(config.Permitted)
	if err != nil {
		return fmt.Errorf("parse permitted set: %w", err)
	}
	ambientSet, err := ToBitset(config.Ambient)
	if err != nil {
		return fmt.Errorf("parse ambient set: %w", err)
	}

	// Drop capabilities not in bounding set
	for i := uint(0); i <= LastCap; i++ {
		if (boundingSet & (1 << i)) == 0 {
			if err := unix.Prctl(unix.PR_CAPBSET_DROP, uintptr(i), 0, 0, 0); err != nil {
				// Ignore EINVAL for capabilities not supported by kernel
				if err != unix.EINVAL {
					return fmt.Errorf("drop bounding cap %d: %w", i, err)
				}
			}
		}
	}

	// Set permitted, effective, and inheritable
	data[0].Effective = uint32(effectiveSet & 0xffffffff)
	data[1].Effective = uint32(effectiveSet >> 32)
	data[0].Permitted = uint32(permittedSet & 0xffffffff)
	data[1].Permitted = uint32(permittedSet >> 32)
	data[0].Inheritable = uint32(inheritableSet & 0xffffffff)
	data[1].Inheritable = uint32(inheritableSet >> 32)

	if err := unix.Capset(&hdr, &data[0]); err != nil {
		return fmt.Errorf("capset: %w", err)
	}

	// Set ambient capabilities
	for i := uint(0); i <= LastCap; i++ {
		if (ambientSet & (1 << i)) != 0 {
			if err := unix.Prctl(unix.PR_CAP_AMBIENT, unix.PR_CAP_AMBIENT_RAISE, uintptr(i), 0, 0); err != nil {
				// Ignore errors - ambient caps may not be supported
			}
		}
	}

	return nil
}

// SetNoNewPrivs sets the no_new_privs flag.
func SetNoNewPrivs() error {
	if err := unix.Prctl(unix.PR_SET_NO_NEW_PRIVS, 1, 0, 0, 0); err != nil {
		return fmt.Errorf("set no_new_privs: %w", err)
	}
	return nil
}

// DropAllCapabilities drops all capabilities from the current process.
func DropAllCapabilities() error {
	// Drop all bounding set capabilities
	for i := uint(0); i <= LastCap; i++ {
		if err := unix.Prctl(unix.PR_CAPBSET_DROP, uintptr(i), 0, 0, 0); err != nil {
			if err != unix.EINVAL {
				return fmt.Errorf("drop cap %d: %w", i, err)
			}
		}
	}

	// Clear all capability sets
	var hdr unix.CapUserHeader
	var data [2]unix.CapUserData

	hdr.Version = unix.LINUX_CAPABILITY_VERSION_3
	hdr.Pid = 0

	// All zeros = no capabilities
	if err := unix.Capset(&hdr, &data[0]); err != nil {
		return fmt.Errorf("capset: %w", err)
	}

	return nil
}

// Ensure unsafe is used
var _ = unsafe.Sizeof(0)
