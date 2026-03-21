//go:build !linux

package namespace

import (
	"errors"

	"github.com/sudokatie/membrane/pkg/oci"
)

var errNotLinux = errors.New("namespace operations require Linux")

// CloneChild is not supported on non-Linux systems.
func CloneChild(config *Config) (int, error) {
	return 0, errNotLinux
}

// Unshare is not supported on non-Linux systems.
func Unshare(flags uintptr) error {
	return errNotLinux
}

// JoinNamespace is not supported on non-Linux systems.
func JoinNamespace(path string, nsType oci.NamespaceType) error {
	return errNotLinux
}

// SetHostname is not supported on non-Linux systems.
func SetHostname(hostname string) error {
	return errNotLinux
}

// SetDomainname is not supported on non-Linux systems.
func SetDomainname(domainname string) error {
	return errNotLinux
}

// WriteUIDMapping is not supported on non-Linux systems.
func WriteUIDMapping(pid int, mappings []oci.LinuxIDMapping) error {
	return errNotLinux
}

// WriteGIDMapping is not supported on non-Linux systems.
func WriteGIDMapping(pid int, mappings []oci.LinuxIDMapping) error {
	return errNotLinux
}

// GetNamespacePath returns the path to a namespace for a process.
func GetNamespacePath(pid int, nsType oci.NamespaceType) string {
	nsName := map[oci.NamespaceType]string{
		oci.PIDNamespace:     "pid",
		oci.NetworkNamespace: "net",
		oci.MountNamespace:   "mnt",
		oci.IPCNamespace:     "ipc",
		oci.UTSNamespace:     "uts",
		oci.UserNamespace:    "user",
		oci.CgroupNamespace:  "cgroup",
	}[nsType]
	return "/proc/" + string(rune(pid)) + "/ns/" + nsName
}
