//go:build !linux

package container

// applyAppArmorProfile is a stub for non-Linux systems.
func applyAppArmorProfile(profile string) error {
	// AppArmor is Linux-specific, silently ignore on other platforms
	return nil
}

// isAppArmorEnabled always returns false on non-Linux systems.
func isAppArmorEnabled() bool {
	return false
}

// changeAppArmorProfile is a stub for non-Linux systems.
func changeAppArmorProfile(profile string) error {
	return nil
}
