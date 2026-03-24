//go:build linux

package container

import (
	"fmt"
	"os"
	"strings"
)

// applyAppArmorProfile applies an AppArmor profile to the current process.
func applyAppArmorProfile(profile string) error {
	if profile == "" {
		return nil
	}

	// Check if AppArmor is enabled
	if !isAppArmorEnabled() {
		// AppArmor not available, skip silently
		return nil
	}

	// Write the profile to /proc/self/attr/apparmor/exec or /proc/self/attr/exec
	// Format: "exec <profile_name>"
	execPath := "/proc/self/attr/apparmor/exec"
	if _, err := os.Stat(execPath); os.IsNotExist(err) {
		execPath = "/proc/self/attr/exec"
	}

	// Handle special profiles
	var profileStr string
	switch profile {
	case "unconfined":
		profileStr = "unconfined"
	default:
		profileStr = fmt.Sprintf("exec %s", profile)
	}

	if err := os.WriteFile(execPath, []byte(profileStr), 0); err != nil {
		return fmt.Errorf("apply apparmor profile: %w", err)
	}

	return nil
}

// isAppArmorEnabled checks if AppArmor is enabled on the system.
func isAppArmorEnabled() bool {
	// Check /sys/module/apparmor/parameters/enabled
	data, err := os.ReadFile("/sys/module/apparmor/parameters/enabled")
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(data)) == "Y"
}

// changeAppArmorProfile changes the AppArmor profile of the current process.
// This is used for "changeprofile" transitions.
func changeAppArmorProfile(profile string) error {
	if profile == "" || !isAppArmorEnabled() {
		return nil
	}

	path := "/proc/self/attr/apparmor/current"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		path = "/proc/self/attr/current"
	}

	profileStr := fmt.Sprintf("changeprofile %s", profile)
	if err := os.WriteFile(path, []byte(profileStr), 0); err != nil {
		return fmt.Errorf("change apparmor profile: %w", err)
	}

	return nil
}
