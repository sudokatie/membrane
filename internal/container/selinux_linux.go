//go:build linux

package container

import (
	"fmt"
	"os"
	"strings"
)

// applySELinuxLabel applies an SELinux label to the current process.
func applySELinuxLabel(label string) error {
	if label == "" {
		return nil
	}

	// Check if SELinux is enabled
	if !isSELinuxEnabled() {
		// SELinux not available, skip silently
		return nil
	}

	// Write the label to /proc/self/attr/exec for the next exec
	execPath := "/proc/self/attr/exec"
	if err := os.WriteFile(execPath, []byte(label), 0); err != nil {
		return fmt.Errorf("apply selinux label: %w", err)
	}

	return nil
}

// isSELinuxEnabled checks if SELinux is enabled on the system.
func isSELinuxEnabled() bool {
	// Check /sys/fs/selinux/enforce
	data, err := os.ReadFile("/sys/fs/selinux/enforce")
	if err != nil {
		// Also check getenforce via /selinux/enforce (older path)
		data, err = os.ReadFile("/selinux/enforce")
		if err != nil {
			return false
		}
	}

	val := strings.TrimSpace(string(data))
	return val == "1" || val == "0" // 1 = enforcing, 0 = permissive, both mean SELinux is on
}

// getSELinuxMode returns the current SELinux mode.
func getSELinuxMode() string {
	data, err := os.ReadFile("/sys/fs/selinux/enforce")
	if err != nil {
		data, err = os.ReadFile("/selinux/enforce")
		if err != nil {
			return "disabled"
		}
	}

	switch strings.TrimSpace(string(data)) {
	case "1":
		return "enforcing"
	case "0":
		return "permissive"
	default:
		return "unknown"
	}
}

// setCurrentSELinuxContext sets the SELinux context of the current process.
func setCurrentSELinuxContext(context string) error {
	if context == "" || !isSELinuxEnabled() {
		return nil
	}

	// Write to /proc/self/attr/current
	if err := os.WriteFile("/proc/self/attr/current", []byte(context), 0); err != nil {
		return fmt.Errorf("set selinux context: %w", err)
	}

	return nil
}

// setKeyLabel sets the SELinux label for keys created by this process.
func setKeyLabel(label string) error {
	if label == "" || !isSELinuxEnabled() {
		return nil
	}

	if err := os.WriteFile("/proc/self/attr/keycreate", []byte(label), 0); err != nil {
		// Non-fatal - keycreate may not be supported
		return nil
	}

	return nil
}
