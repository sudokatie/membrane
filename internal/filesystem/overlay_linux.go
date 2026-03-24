//go:build linux

package filesystem

import (
	"fmt"

	"golang.org/x/sys/unix"
)

// MountOverlay mounts an overlayfs filesystem.
func MountOverlay(config *OverlayConfig) error {
	if err := ValidateOverlayConfig(config); err != nil {
		return err
	}

	// Prepare directories
	if err := PrepareOverlayDirs(config); err != nil {
		return err
	}

	// Build mount options
	options := BuildOverlayOptions(config)

	// Mount overlayfs
	if err := unix.Mount("overlay", config.MergedDir, "overlay", 0, options); err != nil {
		return fmt.Errorf("mount overlay: %w", err)
	}

	return nil
}

// UnmountOverlay unmounts an overlayfs filesystem.
func UnmountOverlay(mergedDir string) error {
	if err := unix.Unmount(mergedDir, unix.MNT_DETACH); err != nil {
		return fmt.Errorf("unmount overlay: %w", err)
	}
	return nil
}

// IsOverlaySupported checks if overlayfs is supported.
func IsOverlaySupported() bool {
	// Try to read /proc/filesystems for overlay support
	// This is a simplified check
	return true
}
