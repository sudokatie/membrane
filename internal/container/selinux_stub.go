//go:build !linux

package container

// applySELinuxLabel is a stub for non-Linux systems.
func applySELinuxLabel(label string) error {
	// SELinux is Linux-specific, silently ignore on other platforms
	return nil
}

// isSELinuxEnabled always returns false on non-Linux systems.
func isSELinuxEnabled() bool {
	return false
}

// getSELinuxMode returns "disabled" on non-Linux systems.
func getSELinuxMode() string {
	return "disabled"
}

// setCurrentSELinuxContext is a stub for non-Linux systems.
func setCurrentSELinuxContext(context string) error {
	return nil
}

// setKeyLabel is a stub for non-Linux systems.
func setKeyLabel(label string) error {
	return nil
}
