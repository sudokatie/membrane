//go:build linux

package filesystem

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
)

// PivotRoot performs pivot_root to change the root filesystem.
// newRoot becomes the new root, and oldRoot is where the old root is moved.
func PivotRoot(newRoot, oldRoot string) error {
	// Ensure paths are absolute
	var err error
	newRoot, err = filepath.Abs(newRoot)
	if err != nil {
		return fmt.Errorf("resolve new root: %w", err)
	}

	// Create the put_old directory
	putOld := filepath.Join(newRoot, oldRoot)
	if err := os.MkdirAll(putOld, 0700); err != nil {
		return fmt.Errorf("create put_old %s: %w", putOld, err)
	}

	// pivot_root requires the new root to be a mount point
	// Bind mount it to itself first
	if err := unix.Mount(newRoot, newRoot, "", unix.MS_BIND|unix.MS_REC, ""); err != nil {
		return fmt.Errorf("bind mount new root: %w", err)
	}

	// Perform pivot_root
	if err := unix.PivotRoot(newRoot, putOld); err != nil {
		return fmt.Errorf("pivot_root: %w", err)
	}

	// Change to new root
	if err := unix.Chdir("/"); err != nil {
		return fmt.Errorf("chdir /: %w", err)
	}

	// Unmount old root
	oldRootPath := oldRoot
	if err := unix.Unmount(oldRootPath, unix.MNT_DETACH); err != nil {
		return fmt.Errorf("unmount old root: %w", err)
	}

	// Remove the old root directory
	if err := os.RemoveAll(oldRootPath); err != nil {
		// Non-fatal: directory may not be empty
	}

	return nil
}

// SetupRootfs performs complete rootfs setup for a container.
// This includes making mounts private, performing mounts, and pivot_root.
func SetupRootfs(rootfs string, mounts []Mount) error {
	// Make the current root private to prevent mount propagation
	if err := MakePrivate("/"); err != nil {
		return fmt.Errorf("make root private: %w", err)
	}

	// Prepare rootfs
	if err := PrepareRootfs(rootfs); err != nil {
		return err
	}

	// Perform all mounts
	if err := MountAll(rootfs, mounts); err != nil {
		return fmt.Errorf("mount filesystems: %w", err)
	}

	// Create device nodes
	if err := CreateDeviceNodes(rootfs); err != nil {
		return fmt.Errorf("create device nodes: %w", err)
	}

	// Perform pivot_root
	if err := PivotRoot(rootfs, "/.pivot_root"); err != nil {
		return fmt.Errorf("pivot_root: %w", err)
	}

	return nil
}

// MaskPath masks a path by bind-mounting /dev/null over it.
func MaskPath(path string) error {
	// Check if path exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}

	// Bind mount /dev/null over it
	if err := unix.Mount("/dev/null", path, "", unix.MS_BIND, ""); err != nil {
		// Try mounting tmpfs for directories
		if info, statErr := os.Stat(path); statErr == nil && info.IsDir() {
			if err := unix.Mount("tmpfs", path, "tmpfs", unix.MS_RDONLY, "size=0"); err != nil {
				return fmt.Errorf("mask %s: %w", path, err)
			}
			return nil
		}
		return fmt.Errorf("mask %s: %w", path, err)
	}
	return nil
}

// ReadonlyPath makes a path read-only by remounting it.
func ReadonlyPath(path string) error {
	// Check if path exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}

	// Bind mount to self, then remount readonly
	if err := unix.Mount(path, path, "", unix.MS_BIND|unix.MS_REC, ""); err != nil {
		return fmt.Errorf("bind %s: %w", path, err)
	}

	if err := unix.Mount("", path, "", unix.MS_BIND|unix.MS_REMOUNT|unix.MS_RDONLY|unix.MS_REC, ""); err != nil {
		return fmt.Errorf("remount readonly %s: %w", path, err)
	}

	return nil
}

// MaskPaths masks multiple paths.
func MaskPaths(paths []string) error {
	for _, p := range paths {
		if err := MaskPath(p); err != nil {
			return err
		}
	}
	return nil
}

// ReadonlyPaths makes multiple paths read-only.
func ReadonlyPaths(paths []string) error {
	for _, p := range paths {
		if err := ReadonlyPath(p); err != nil {
			return err
		}
	}
	return nil
}
