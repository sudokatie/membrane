//go:build !linux

package filesystem

import "errors"

var errNotLinux = errors.New("filesystem operations require Linux")

// MountAll is not supported on non-Linux systems.
func MountAll(rootfs string, mounts []Mount) error {
	return errNotLinux
}

// MountSingle is not supported on non-Linux systems.
func MountSingle(source, target, fstype string, flags uintptr, data string) error {
	return errNotLinux
}

// BindMount is not supported on non-Linux systems.
func BindMount(source, target string, readonly bool) error {
	return errNotLinux
}

// Unmount is not supported on non-Linux systems.
func Unmount(target string) error {
	return errNotLinux
}

// UnmountLazy is not supported on non-Linux systems.
func UnmountLazy(target string) error {
	return errNotLinux
}

// MakePrivate is not supported on non-Linux systems.
func MakePrivate(target string) error {
	return errNotLinux
}

// MakeSlave is not supported on non-Linux systems.
func MakeSlave(target string) error {
	return errNotLinux
}

// CreateDeviceNodes is not supported on non-Linux systems.
func CreateDeviceNodes(rootfs string) error {
	return errNotLinux
}
