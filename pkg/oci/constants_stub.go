//go:build !linux

package oci

// MountFlags stub for non-Linux systems.
// These values match Linux constants for cross-platform testing.
var MountFlags = map[string]uintptr{
	"ro":          1,          // MS_RDONLY
	"nosuid":      2,          // MS_NOSUID
	"nodev":       4,          // MS_NODEV
	"noexec":      8,          // MS_NOEXEC
	"sync":        16,         // MS_SYNCHRONOUS
	"remount":     32,         // MS_REMOUNT
	"mand":        64,         // MS_MANDLOCK
	"dirsync":     128,        // MS_DIRSYNC
	"noatime":     1024,       // MS_NOATIME
	"nodiratime":  2048,       // MS_NODIRATIME
	"bind":        4096,       // MS_BIND
	"rbind":       4096 | 16384,
	"move":        8192,       // MS_MOVE
	"rec":         16384,      // MS_REC
	"silent":      32768,      // MS_SILENT
	"relatime":    2097152,    // MS_RELATIME
	"strictatime": 16777216,   // MS_STRICTATIME
	"private":     262144,     // MS_PRIVATE
	"rprivate":    262144 | 16384,
	"shared":      1048576,    // MS_SHARED
	"rshared":     1048576 | 16384,
	"slave":       524288,     // MS_SLAVE
	"rslave":      524288 | 16384,
	"unbindable":  131072,     // MS_UNBINDABLE
	"runbindable": 131072 | 16384,
}

// CloneFlags stub for non-Linux systems.
var CloneFlags = map[NamespaceType]uintptr{
	PIDNamespace:     0x20000000, // CLONE_NEWPID
	NetworkNamespace: 0x40000000, // CLONE_NEWNET
	MountNamespace:   0x00020000, // CLONE_NEWNS
	IPCNamespace:     0x08000000, // CLONE_NEWIPC
	UTSNamespace:     0x04000000, // CLONE_NEWUTS
	UserNamespace:    0x10000000, // CLONE_NEWUSER
	CgroupNamespace:  0x02000000, // CLONE_NEWCGROUP
}
