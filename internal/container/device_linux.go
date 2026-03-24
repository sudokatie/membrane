//go:build linux

package container

import (
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

// createDeviceNode creates a device node at the specified path.
func createDeviceNode(path string, mode uint32, major, minor int64, uid, gid *uint32) error {
	// Create the device
	dev := unix.Mkdev(uint32(major), uint32(minor))
	if err := unix.Mknod(path, mode, int(dev)); err != nil {
		if !os.IsExist(err) {
			return fmt.Errorf("mknod %s: %w", path, err)
		}
	}

	// Set ownership if specified
	u := -1
	g := -1
	if uid != nil {
		u = int(*uid)
	}
	if gid != nil {
		g = int(*gid)
	}
	if u >= 0 || g >= 0 {
		if err := unix.Chown(path, u, g); err != nil {
			return fmt.Errorf("chown %s: %w", path, err)
		}
	}

	return nil
}
