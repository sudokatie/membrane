//go:build !linux

package filesystem

import "errors"

var errPivotNotLinux = errors.New("pivot_root requires Linux")

// PivotRoot is not supported on non-Linux systems.
func PivotRoot(newRoot, oldRoot string) error {
	return errPivotNotLinux
}

// SetupRootfs is not supported on non-Linux systems.
func SetupRootfs(rootfs string, mounts []Mount) error {
	return errPivotNotLinux
}

// MaskPath is not supported on non-Linux systems.
func MaskPath(path string) error {
	return errPivotNotLinux
}

// ReadonlyPath is not supported on non-Linux systems.
func ReadonlyPath(path string) error {
	return errPivotNotLinux
}

// MaskPaths is not supported on non-Linux systems.
func MaskPaths(paths []string) error {
	return errPivotNotLinux
}

// ReadonlyPaths is not supported on non-Linux systems.
func ReadonlyPaths(paths []string) error {
	return errPivotNotLinux
}
