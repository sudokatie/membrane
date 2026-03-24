//go:build !linux

package container

import "fmt"

// applySysctls is a stub for non-Linux systems.
func applySysctls(sysctls map[string]string) error {
	if len(sysctls) > 0 {
		return fmt.Errorf("sysctls not supported on this platform")
	}
	return nil
}
