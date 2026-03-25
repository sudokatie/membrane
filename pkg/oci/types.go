// Package oci defines types for the OCI runtime specification.
package oci

// Spec is the OCI runtime configuration.
type Spec struct {
	// Version is the OCI spec version (required).
	Version string `json:"ociVersion"`
	// Root is the root filesystem configuration.
	Root *Root `json:"root,omitempty"`
	// Process is the container process configuration.
	Process *Process `json:"process,omitempty"`
	// Hostname is the container hostname.
	Hostname string `json:"hostname,omitempty"`
	// Mounts are filesystem mounts.
	Mounts []Mount `json:"mounts,omitempty"`
	// Hooks are lifecycle hooks.
	Hooks *Hooks `json:"hooks,omitempty"`
	// Linux contains Linux-specific configuration.
	Linux *Linux `json:"linux,omitempty"`
	// Annotations are arbitrary metadata.
	Annotations map[string]string `json:"annotations,omitempty"`
}

// Hooks contains lifecycle hooks for the container.
type Hooks struct {
	// Prestart is called after the start operation is called but before
	// the user-specified program has been executed.
	// Deprecated: use CreateRuntime, CreateContainer, and StartContainer instead.
	Prestart []Hook `json:"prestart,omitempty"`
	// CreateRuntime is called during the create operation after the runtime
	// environment has been created but before the pivot_root has been executed.
	CreateRuntime []Hook `json:"createRuntime,omitempty"`
	// CreateContainer is called during the create operation after the runtime
	// environment has been created and after the pivot_root has been executed
	// but before the user-specified program has executed.
	CreateContainer []Hook `json:"createContainer,omitempty"`
	// StartContainer is called before executing the user-specified process
	// inside the container.
	StartContainer []Hook `json:"startContainer,omitempty"`
	// Poststart is called after the user-specified process is executed but
	// before the start operation returns.
	Poststart []Hook `json:"poststart,omitempty"`
	// Poststop is called after the container is deleted but before the delete
	// operation returns.
	Poststop []Hook `json:"poststop,omitempty"`
}

// Hook specifies a command to run at a particular point in the lifecycle.
type Hook struct {
	// Path is the absolute path to the hook binary.
	Path string `json:"path"`
	// Args are arguments to pass to the hook binary.
	Args []string `json:"args,omitempty"`
	// Env is environment variables for the hook.
	Env []string `json:"env,omitempty"`
	// Timeout is the number of seconds before the hook times out.
	Timeout *int `json:"timeout,omitempty"`
}

// Root is the root filesystem configuration.
type Root struct {
	// Path is the path to the root filesystem.
	Path string `json:"path"`
	// Readonly sets the root filesystem as read-only.
	Readonly bool `json:"readonly,omitempty"`
}

// Process contains process configuration.
type Process struct {
	// Terminal creates a pseudo-terminal.
	Terminal bool `json:"terminal,omitempty"`
	// User contains user information.
	User User `json:"user"`
	// Args are the command line arguments.
	Args []string `json:"args"`
	// Env is the environment variables.
	Env []string `json:"env,omitempty"`
	// Cwd is the current working directory.
	Cwd string `json:"cwd"`
	// Capabilities are Linux capabilities.
	Capabilities *Capabilities `json:"capabilities,omitempty"`
	// Rlimits are resource limits.
	Rlimits []POSIXRlimit `json:"rlimits,omitempty"`
	// NoNewPrivileges prevents gaining new privileges.
	NoNewPrivileges bool `json:"noNewPrivileges,omitempty"`
	// ApparmorProfile is the apparmor profile.
	ApparmorProfile string `json:"apparmorProfile,omitempty"`
	// SelinuxLabel is the SELinux label.
	SelinuxLabel string `json:"selinuxLabel,omitempty"`
}

// User contains user information.
type User struct {
	// UID is the user id.
	UID uint32 `json:"uid"`
	// GID is the group id.
	GID uint32 `json:"gid"`
	// AdditionalGids are additional group ids.
	AdditionalGids []uint32 `json:"additionalGids,omitempty"`
	// Umask is the umask for the process.
	Umask *uint32 `json:"umask,omitempty"`
}

// Capabilities contains Linux capabilities.
type Capabilities struct {
	// Bounding is the bounding set.
	Bounding []string `json:"bounding,omitempty"`
	// Effective is the effective set.
	Effective []string `json:"effective,omitempty"`
	// Inheritable is the inheritable set.
	Inheritable []string `json:"inheritable,omitempty"`
	// Permitted is the permitted set.
	Permitted []string `json:"permitted,omitempty"`
	// Ambient is the ambient set.
	Ambient []string `json:"ambient,omitempty"`
}

// POSIXRlimit is a resource limit.
type POSIXRlimit struct {
	// Type is the resource type (e.g., RLIMIT_NOFILE).
	Type string `json:"type"`
	// Hard is the hard limit.
	Hard uint64 `json:"hard"`
	// Soft is the soft limit.
	Soft uint64 `json:"soft"`
}

// Mount is a filesystem mount.
type Mount struct {
	// Destination is the mount point.
	Destination string `json:"destination"`
	// Type is the filesystem type.
	Type string `json:"type,omitempty"`
	// Source is the mount source.
	Source string `json:"source,omitempty"`
	// Options are mount options.
	Options []string `json:"options,omitempty"`
}

// Linux contains Linux-specific configuration.
type Linux struct {
	// Namespaces are the namespaces to create/join.
	Namespaces []LinuxNamespace `json:"namespaces,omitempty"`
	// UIDMappings are user id mappings for user namespaces.
	UIDMappings []LinuxIDMapping `json:"uidMappings,omitempty"`
	// GIDMappings are group id mappings for user namespaces.
	GIDMappings []LinuxIDMapping `json:"gidMappings,omitempty"`
	// Devices are devices to create in the container.
	Devices []LinuxDevice `json:"devices,omitempty"`
	// CgroupsPath is the path to the cgroup.
	CgroupsPath string `json:"cgroupsPath,omitempty"`
	// Resources are cgroup resource limits.
	Resources *LinuxResources `json:"resources,omitempty"`
	// Seccomp is the seccomp filter configuration.
	Seccomp *LinuxSeccomp `json:"seccomp,omitempty"`
	// Sysctl are kernel parameters.
	Sysctl map[string]string `json:"sysctl,omitempty"`
	// MaskedPaths are paths to mask in the container.
	MaskedPaths []string `json:"maskedPaths,omitempty"`
	// ReadonlyPaths are paths to make read-only.
	ReadonlyPaths []string `json:"readonlyPaths,omitempty"`
	// MountLabel is the SELinux mount label.
	MountLabel string `json:"mountLabel,omitempty"`
	// RootfsPropagation is the rootfs mount propagation.
	RootfsPropagation string `json:"rootfsPropagation,omitempty"`
}

// LinuxNamespace is a namespace configuration.
type LinuxNamespace struct {
	// Type is the namespace type.
	Type NamespaceType `json:"type"`
	// Path is the path to an existing namespace.
	Path string `json:"path,omitempty"`
}

// NamespaceType is the type of namespace.
type NamespaceType string

// Namespace types.
const (
	PIDNamespace     NamespaceType = "pid"
	NetworkNamespace NamespaceType = "network"
	MountNamespace   NamespaceType = "mount"
	IPCNamespace     NamespaceType = "ipc"
	UTSNamespace     NamespaceType = "uts"
	UserNamespace    NamespaceType = "user"
	CgroupNamespace  NamespaceType = "cgroup"
)

// LinuxIDMapping is a user/group id mapping.
type LinuxIDMapping struct {
	// ContainerID is the starting id in the container.
	ContainerID uint32 `json:"containerID"`
	// HostID is the starting id on the host.
	HostID uint32 `json:"hostID"`
	// Size is the number of ids to map.
	Size uint32 `json:"size"`
}

// LinuxDevice is a device configuration.
type LinuxDevice struct {
	// Path is the device path in the container.
	Path string `json:"path"`
	// Type is the device type (c, b, u, p).
	Type string `json:"type"`
	// Major is the major device number.
	Major int64 `json:"major"`
	// Minor is the minor device number.
	Minor int64 `json:"minor"`
	// FileMode is the device file mode.
	FileMode *uint32 `json:"fileMode,omitempty"`
	// UID is the device owner.
	UID *uint32 `json:"uid,omitempty"`
	// GID is the device group.
	GID *uint32 `json:"gid,omitempty"`
}

// LinuxResources is cgroup resource configuration.
type LinuxResources struct {
	// Memory is memory cgroup settings.
	Memory *LinuxMemory `json:"memory,omitempty"`
	// CPU is CPU cgroup settings.
	CPU *LinuxCPU `json:"cpu,omitempty"`
	// Pids is pids cgroup settings.
	Pids *LinuxPids `json:"pids,omitempty"`
	// BlockIO is block IO cgroup settings.
	BlockIO *LinuxBlockIO `json:"blockIO,omitempty"`
	// HugepageLimits are hugepage cgroup settings.
	HugepageLimits []LinuxHugepageLimit `json:"hugepageLimits,omitempty"`
	// Network is network cgroup settings.
	Network *LinuxNetwork `json:"network,omitempty"`
}

// LinuxMemory is memory cgroup configuration.
type LinuxMemory struct {
	// Limit is the memory limit in bytes.
	Limit *int64 `json:"limit,omitempty"`
	// Reservation is the soft memory limit.
	Reservation *int64 `json:"reservation,omitempty"`
	// Swap is the swap limit.
	Swap *int64 `json:"swap,omitempty"`
	// Kernel is the kernel memory limit.
	Kernel *int64 `json:"kernel,omitempty"`
	// KernelTCP is the kernel TCP buffer limit.
	KernelTCP *int64 `json:"kernelTCP,omitempty"`
	// Swappiness is the swappiness (0-100).
	Swappiness *uint64 `json:"swappiness,omitempty"`
	// DisableOOMKiller disables the OOM killer.
	DisableOOMKiller *bool `json:"disableOOMKiller,omitempty"`
}

// LinuxCPU is CPU cgroup configuration.
type LinuxCPU struct {
	// Shares is the CPU shares.
	Shares *uint64 `json:"shares,omitempty"`
	// Quota is the CPU CFS quota in microseconds.
	Quota *int64 `json:"quota,omitempty"`
	// Period is the CPU CFS period in microseconds.
	Period *uint64 `json:"period,omitempty"`
	// RealtimeRuntime is the realtime runtime in microseconds.
	RealtimeRuntime *int64 `json:"realtimeRuntime,omitempty"`
	// RealtimePeriod is the realtime period in microseconds.
	RealtimePeriod *uint64 `json:"realtimePeriod,omitempty"`
	// Cpus is the CPU set (e.g., "0-2,6").
	Cpus string `json:"cpus,omitempty"`
	// Mems is the memory nodes (e.g., "0-1").
	Mems string `json:"mems,omitempty"`
}

// LinuxPids is pids cgroup configuration.
type LinuxPids struct {
	// Limit is the maximum number of PIDs.
	Limit int64 `json:"limit"`
}

// LinuxBlockIO is block IO cgroup configuration.
type LinuxBlockIO struct {
	// Weight is the block IO weight (10-1000).
	Weight *uint16 `json:"weight,omitempty"`
	// LeafWeight is the leaf weight.
	LeafWeight *uint16 `json:"leafWeight,omitempty"`
	// WeightDevice is per-device weight.
	WeightDevice []LinuxWeightDevice `json:"weightDevice,omitempty"`
	// ThrottleReadBpsDevice is read bytes per second limit.
	ThrottleReadBpsDevice []LinuxThrottleDevice `json:"throttleReadBpsDevice,omitempty"`
	// ThrottleWriteBpsDevice is write bytes per second limit.
	ThrottleWriteBpsDevice []LinuxThrottleDevice `json:"throttleWriteBpsDevice,omitempty"`
	// ThrottleReadIOPSDevice is read IOPS limit.
	ThrottleReadIOPSDevice []LinuxThrottleDevice `json:"throttleReadIOPSDevice,omitempty"`
	// ThrottleWriteIOPSDevice is write IOPS limit.
	ThrottleWriteIOPSDevice []LinuxThrottleDevice `json:"throttleWriteIOPSDevice,omitempty"`
}

// LinuxWeightDevice is per-device weight.
type LinuxWeightDevice struct {
	// Major is the major device number.
	Major int64 `json:"major"`
	// Minor is the minor device number.
	Minor int64 `json:"minor"`
	// Weight is the device weight.
	Weight *uint16 `json:"weight,omitempty"`
	// LeafWeight is the leaf weight.
	LeafWeight *uint16 `json:"leafWeight,omitempty"`
}

// LinuxThrottleDevice is per-device throttle limit.
type LinuxThrottleDevice struct {
	// Major is the major device number.
	Major int64 `json:"major"`
	// Minor is the minor device number.
	Minor int64 `json:"minor"`
	// Rate is the throttle rate.
	Rate uint64 `json:"rate"`
}

// LinuxHugepageLimit is a hugepage limit.
type LinuxHugepageLimit struct {
	// Pagesize is the hugepage size (e.g., "2MB").
	Pagesize string `json:"pageSize"`
	// Limit is the limit in bytes.
	Limit uint64 `json:"limit"`
}

// LinuxNetwork is network cgroup configuration.
type LinuxNetwork struct {
	// ClassID is the network class id.
	ClassID *uint32 `json:"classID,omitempty"`
	// Priorities are interface priorities.
	Priorities []LinuxInterfacePriority `json:"priorities,omitempty"`
}

// LinuxInterfacePriority is an interface priority.
type LinuxInterfacePriority struct {
	// Name is the interface name.
	Name string `json:"name"`
	// Priority is the priority.
	Priority uint32 `json:"priority"`
}

// LinuxSeccomp is seccomp configuration.
type LinuxSeccomp struct {
	// DefaultAction is the default action for syscalls.
	DefaultAction SeccompAction `json:"defaultAction"`
	// Architectures are the allowed architectures.
	Architectures []Arch `json:"architectures,omitempty"`
	// Syscalls are the syscall rules.
	Syscalls []LinuxSyscall `json:"syscalls,omitempty"`
}

// SeccompAction is a seccomp action.
type SeccompAction string

// Seccomp actions.
const (
	ActKill  SeccompAction = "SCMP_ACT_KILL"
	ActTrap  SeccompAction = "SCMP_ACT_TRAP"
	ActErrno SeccompAction = "SCMP_ACT_ERRNO"
	ActTrace SeccompAction = "SCMP_ACT_TRACE"
	ActAllow SeccompAction = "SCMP_ACT_ALLOW"
	ActLog   SeccompAction = "SCMP_ACT_LOG"
)

// Arch is a seccomp architecture.
type Arch string

// Architecture values.
const (
	ArchX86         Arch = "SCMP_ARCH_X86"
	ArchX86_64      Arch = "SCMP_ARCH_X86_64"
	ArchARM         Arch = "SCMP_ARCH_ARM"
	ArchAARCH64     Arch = "SCMP_ARCH_AARCH64"
	ArchMIPS        Arch = "SCMP_ARCH_MIPS"
	ArchMIPS64      Arch = "SCMP_ARCH_MIPS64"
	ArchMIPSEL      Arch = "SCMP_ARCH_MIPSEL"
	ArchMIPSEL64    Arch = "SCMP_ARCH_MIPSEL64"
	ArchPPC         Arch = "SCMP_ARCH_PPC"
	ArchPPC64       Arch = "SCMP_ARCH_PPC64"
	ArchPPC64LE     Arch = "SCMP_ARCH_PPC64LE"
	ArchS390        Arch = "SCMP_ARCH_S390"
	ArchS390X       Arch = "SCMP_ARCH_S390X"
	ArchRISCV64     Arch = "SCMP_ARCH_RISCV64"
)

// LinuxSyscall is a syscall rule.
type LinuxSyscall struct {
	// Names are the syscall names.
	Names []string `json:"names"`
	// Action is the action for these syscalls.
	Action SeccompAction `json:"action"`
	// ErrnoRet is the errno to return (for SCMP_ACT_ERRNO).
	ErrnoRet *uint `json:"errnoRet,omitempty"`
	// Args are argument filters.
	Args []LinuxSeccompArg `json:"args,omitempty"`
}

// LinuxSeccompArg is a syscall argument filter.
type LinuxSeccompArg struct {
	// Index is the argument index (0-5).
	Index uint `json:"index"`
	// Value is the value to compare.
	Value uint64 `json:"value"`
	// ValueTwo is the second value (for SCMP_CMP_MASKED_EQ).
	ValueTwo uint64 `json:"valueTwo,omitempty"`
	// Op is the comparison operator.
	Op SeccompOperator `json:"op"`
}

// SeccompOperator is a seccomp comparison operator.
type SeccompOperator string

// Seccomp operators.
const (
	OpNotEqual     SeccompOperator = "SCMP_CMP_NE"
	OpLessThan     SeccompOperator = "SCMP_CMP_LT"
	OpLessEqual    SeccompOperator = "SCMP_CMP_LE"
	OpEqualTo      SeccompOperator = "SCMP_CMP_EQ"
	OpGreaterEqual SeccompOperator = "SCMP_CMP_GE"
	OpGreaterThan  SeccompOperator = "SCMP_CMP_GT"
	OpMaskedEqual  SeccompOperator = "SCMP_CMP_MASKED_EQ"
)
