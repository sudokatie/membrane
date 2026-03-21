//go:build !linux

package seccomp

import "errors"

var errNotLinux = errors.New("seccomp requires Linux")

// LoadFilter is not supported on non-Linux systems.
func LoadFilter(profile *Profile) error {
	return errNotLinux
}
