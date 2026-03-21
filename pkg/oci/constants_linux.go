//go:build linux

package oci

import "syscall"

// Mount flags mapped from string options.
var MountFlags = map[string]uintptr{
	"ro":          syscall.MS_RDONLY,
	"nosuid":      syscall.MS_NOSUID,
	"nodev":       syscall.MS_NODEV,
	"noexec":      syscall.MS_NOEXEC,
	"sync":        syscall.MS_SYNCHRONOUS,
	"remount":     syscall.MS_REMOUNT,
	"mand":        syscall.MS_MANDLOCK,
	"dirsync":     syscall.MS_DIRSYNC,
	"noatime":     syscall.MS_NOATIME,
	"nodiratime":  syscall.MS_NODIRATIME,
	"bind":        syscall.MS_BIND,
	"rbind":       syscall.MS_BIND | syscall.MS_REC,
	"move":        syscall.MS_MOVE,
	"rec":         syscall.MS_REC,
	"silent":      syscall.MS_SILENT,
	"relatime":    syscall.MS_RELATIME,
	"strictatime": syscall.MS_STRICTATIME,
	"private":     syscall.MS_PRIVATE,
	"rprivate":    syscall.MS_PRIVATE | syscall.MS_REC,
	"shared":      syscall.MS_SHARED,
	"rshared":     syscall.MS_SHARED | syscall.MS_REC,
	"slave":       syscall.MS_SLAVE,
	"rslave":      syscall.MS_SLAVE | syscall.MS_REC,
	"unbindable":  syscall.MS_UNBINDABLE,
	"runbindable": syscall.MS_UNBINDABLE | syscall.MS_REC,
}

// Clone flags for namespaces.
var CloneFlags = map[NamespaceType]uintptr{
	PIDNamespace:     syscall.CLONE_NEWPID,
	NetworkNamespace: syscall.CLONE_NEWNET,
	MountNamespace:   syscall.CLONE_NEWNS,
	IPCNamespace:     syscall.CLONE_NEWIPC,
	UTSNamespace:     syscall.CLONE_NEWUTS,
	UserNamespace:    syscall.CLONE_NEWUSER,
	CgroupNamespace:  0x02000000, // CLONE_NEWCGROUP
}
