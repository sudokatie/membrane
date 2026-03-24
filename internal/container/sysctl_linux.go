//go:build linux

package container

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// applySysctls applies kernel parameters.
func applySysctls(sysctls map[string]string) error {
	for key, value := range sysctls {
		if err := writeSysctl(key, value); err != nil {
			return fmt.Errorf("sysctl %s: %w", key, err)
		}
	}
	return nil
}

// writeSysctl writes a single sysctl value.
func writeSysctl(key, value string) error {
	// Convert sysctl key to path: kernel.pid_max -> /proc/sys/kernel/pid_max
	path := filepath.Join("/proc/sys", strings.ReplaceAll(key, ".", "/"))

	// Write the value
	if err := os.WriteFile(path, []byte(value), 0644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}

	return nil
}

// validateSysctlKey checks if a sysctl key is safe to set in a container.
// Some sysctls are namespaced and safe, others affect the whole host.
func validateSysctlKey(key string) bool {
	// Safe namespaced sysctls
	safePrefixes := []string{
		"kernel.shm",
		"kernel.msg",
		"kernel.sem",
		"fs.mqueue.",
		"net.",
	}

	for _, prefix := range safePrefixes {
		if strings.HasPrefix(key, prefix) {
			return true
		}
	}

	// Specific safe sysctls
	safeSysctls := map[string]bool{
		"kernel.domainname": true,
		"kernel.hostname":   true,
	}

	return safeSysctls[key]
}
