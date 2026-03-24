//go:build !linux

package container

import "fmt"

// createDeviceNode is a stub for non-Linux systems.
func createDeviceNode(path string, mode uint32, major, minor int64, uid, gid *uint32) error {
	return fmt.Errorf("device creation not supported on this platform")
}
