//go:build !linux

package filesystem

import "fmt"

// MountOverlay is a stub for non-Linux systems.
func MountOverlay(config *OverlayConfig) error {
	return fmt.Errorf("overlayfs not supported on this platform")
}

// UnmountOverlay is a stub for non-Linux systems.
func UnmountOverlay(mergedDir string) error {
	return fmt.Errorf("overlayfs not supported on this platform")
}

// IsOverlaySupported returns false on non-Linux systems.
func IsOverlaySupported() bool {
	return false
}
