//go:build linux

package namespace

import (
	"fmt"
	"os"
	"runtime"
	"syscall"

	"github.com/sudokatie/membrane/pkg/oci"
	"golang.org/x/sys/unix"
)

// CloneChild forks a new process with the specified namespaces.
// Returns the child PID in the parent, or 0 in the child.
func CloneChild(config *Config) (int, error) {
	// Get clone flags
	flags := config.CloneFlags()

	// Add SIGCHLD so parent gets notified when child exits
	flags |= syscall.SIGCHLD

	// Lock OS thread for namespace operations
	runtime.LockOSThread()

	// Fork with clone
	pid, _, errno := syscall.Syscall6(
		syscall.SYS_CLONE,
		flags,
		0, // child stack (0 = use copy-on-write)
		0, // parent tid
		0, // child tid
		0, // tls
		0,
	)

	if errno != 0 {
		runtime.UnlockOSThread()
		return 0, fmt.Errorf("clone failed: %v", errno)
	}

	if pid == 0 {
		// In child - keep thread locked
		return 0, nil
	}

	// In parent
	runtime.UnlockOSThread()
	return int(pid), nil
}

// Unshare creates new namespaces for the current process.
func Unshare(flags uintptr) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if err := unix.Unshare(int(flags)); err != nil {
		return fmt.Errorf("unshare failed: %w", err)
	}
	return nil
}

// JoinNamespace joins an existing namespace.
func JoinNamespace(path string, nsType oci.NamespaceType) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open namespace %s: %w", path, err)
	}
	defer f.Close()

	// Get the setns flag for this namespace type
	flag, ok := setnsFlags[nsType]
	if !ok {
		return fmt.Errorf("unknown namespace type: %s", nsType)
	}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if err := unix.Setns(int(f.Fd()), flag); err != nil {
		return fmt.Errorf("setns failed: %w", err)
	}

	return nil
}

// setnsFlags maps namespace types to setns flags.
var setnsFlags = map[oci.NamespaceType]int{
	oci.PIDNamespace:     unix.CLONE_NEWPID,
	oci.NetworkNamespace: unix.CLONE_NEWNET,
	oci.MountNamespace:   unix.CLONE_NEWNS,
	oci.IPCNamespace:     unix.CLONE_NEWIPC,
	oci.UTSNamespace:     unix.CLONE_NEWUTS,
	oci.UserNamespace:    unix.CLONE_NEWUSER,
	oci.CgroupNamespace:  unix.CLONE_NEWCGROUP,
}

// SetHostname sets the hostname in a UTS namespace.
func SetHostname(hostname string) error {
	if err := unix.Sethostname([]byte(hostname)); err != nil {
		return fmt.Errorf("sethostname failed: %w", err)
	}
	return nil
}

// SetDomainname sets the domainname in a UTS namespace.
func SetDomainname(domainname string) error {
	if err := unix.Setdomainname([]byte(domainname)); err != nil {
		return fmt.Errorf("setdomainname failed: %w", err)
	}
	return nil
}

// WriteUIDMapping writes UID mappings for a user namespace.
func WriteUIDMapping(pid int, mappings []oci.LinuxIDMapping) error {
	return writeIDMapping(fmt.Sprintf("/proc/%d/uid_map", pid), mappings)
}

// WriteGIDMapping writes GID mappings for a user namespace.
func WriteGIDMapping(pid int, mappings []oci.LinuxIDMapping) error {
	// Disable setgroups first (required before writing gid_map)
	setgroupsPath := fmt.Sprintf("/proc/%d/setgroups", pid)
	if err := os.WriteFile(setgroupsPath, []byte("deny"), 0644); err != nil {
		// Ignore error - file may not exist on older kernels
	}

	return writeIDMapping(fmt.Sprintf("/proc/%d/gid_map", pid), mappings)
}

func writeIDMapping(path string, mappings []oci.LinuxIDMapping) error {
	f, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	for _, m := range mappings {
		line := fmt.Sprintf("%d %d %d\n", m.ContainerID, m.HostID, m.Size)
		if _, err := f.WriteString(line); err != nil {
			return fmt.Errorf("write %s: %w", path, err)
		}
	}

	return nil
}

// GetNamespacePath returns the path to a namespace for a process.
func GetNamespacePath(pid int, nsType oci.NamespaceType) string {
	nsName := nsTypeToName[nsType]
	return fmt.Sprintf("/proc/%d/ns/%s", pid, nsName)
}

var nsTypeToName = map[oci.NamespaceType]string{
	oci.PIDNamespace:     "pid",
	oci.NetworkNamespace: "net",
	oci.MountNamespace:   "mnt",
	oci.IPCNamespace:     "ipc",
	oci.UTSNamespace:     "uts",
	oci.UserNamespace:    "user",
	oci.CgroupNamespace:  "cgroup",
}
