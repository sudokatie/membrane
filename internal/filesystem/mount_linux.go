//go:build linux

package filesystem

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
)

// MountAll performs all mounts in the config.
func MountAll(rootfs string, mounts []Mount) error {
	for _, m := range mounts {
		target := filepath.Join(rootfs, m.Target)

		// Create mount point
		if err := os.MkdirAll(target, 0755); err != nil {
			return fmt.Errorf("create mountpoint %s: %w", m.Target, err)
		}

		// Perform mount
		if err := unix.Mount(m.Source, target, m.FSType, m.Flags, m.Data); err != nil {
			return fmt.Errorf("mount %s on %s: %w", m.Source, m.Target, err)
		}
	}
	return nil
}

// MountSingle performs a single mount operation.
func MountSingle(source, target, fstype string, flags uintptr, data string) error {
	if err := os.MkdirAll(target, 0755); err != nil {
		return fmt.Errorf("create mountpoint %s: %w", target, err)
	}

	if err := unix.Mount(source, target, fstype, flags, data); err != nil {
		return fmt.Errorf("mount %s on %s: %w", source, target, err)
	}
	return nil
}

// BindMount performs a bind mount.
func BindMount(source, target string, readonly bool) error {
	flags := uintptr(unix.MS_BIND)

	if err := os.MkdirAll(target, 0755); err != nil {
		return fmt.Errorf("create mountpoint %s: %w", target, err)
	}

	// First bind mount
	if err := unix.Mount(source, target, "", flags, ""); err != nil {
		return fmt.Errorf("bind mount %s to %s: %w", source, target, err)
	}

	// If readonly, remount with MS_RDONLY
	if readonly {
		flags := uintptr(unix.MS_BIND | unix.MS_REMOUNT | unix.MS_RDONLY)
		if err := unix.Mount("", target, "", flags, ""); err != nil {
			return fmt.Errorf("remount %s readonly: %w", target, err)
		}
	}

	return nil
}

// Unmount unmounts a filesystem.
func Unmount(target string) error {
	if err := unix.Unmount(target, 0); err != nil {
		return fmt.Errorf("unmount %s: %w", target, err)
	}
	return nil
}

// UnmountLazy performs a lazy unmount.
func UnmountLazy(target string) error {
	if err := unix.Unmount(target, unix.MNT_DETACH); err != nil {
		return fmt.Errorf("lazy unmount %s: %w", target, err)
	}
	return nil
}

// MakePrivate makes a mount point private.
func MakePrivate(target string) error {
	if err := unix.Mount("", target, "", unix.MS_PRIVATE|unix.MS_REC, ""); err != nil {
		return fmt.Errorf("make private %s: %w", target, err)
	}
	return nil
}

// MakeSlave makes a mount point a slave.
func MakeSlave(target string) error {
	if err := unix.Mount("", target, "", unix.MS_SLAVE|unix.MS_REC, ""); err != nil {
		return fmt.Errorf("make slave %s: %w", target, err)
	}
	return nil
}

// CreateDeviceNodes creates essential device nodes in /dev.
func CreateDeviceNodes(rootfs string) error {
	devPath := filepath.Join(rootfs, "dev")
	if err := os.MkdirAll(devPath, 0755); err != nil {
		return err
	}

	devices := []struct {
		path  string
		mode  uint32
		major uint32
		minor uint32
	}{
		{"/dev/null", unix.S_IFCHR | 0666, 1, 3},
		{"/dev/zero", unix.S_IFCHR | 0666, 1, 5},
		{"/dev/full", unix.S_IFCHR | 0666, 1, 7},
		{"/dev/random", unix.S_IFCHR | 0666, 1, 8},
		{"/dev/urandom", unix.S_IFCHR | 0666, 1, 9},
		{"/dev/tty", unix.S_IFCHR | 0666, 5, 0},
	}

	for _, d := range devices {
		path := filepath.Join(rootfs, d.path)
		dev := unix.Mkdev(d.major, d.minor)
		if err := unix.Mknod(path, d.mode, int(dev)); err != nil {
			if !os.IsExist(err) {
				return fmt.Errorf("mknod %s: %w", d.path, err)
			}
		}
	}

	// Create symlinks
	symlinks := [][2]string{
		{"/proc/self/fd", "/dev/fd"},
		{"/proc/self/fd/0", "/dev/stdin"},
		{"/proc/self/fd/1", "/dev/stdout"},
		{"/proc/self/fd/2", "/dev/stderr"},
	}

	for _, s := range symlinks {
		path := filepath.Join(rootfs, s[1])
		if err := os.Symlink(s[0], path); err != nil {
			if !os.IsExist(err) {
				return fmt.Errorf("symlink %s -> %s: %w", s[1], s[0], err)
			}
		}
	}

	return nil
}
