package seccomp

// DefaultProfile returns a default seccomp profile that blocks dangerous syscalls.
// This is based on Docker's default seccomp profile.
func DefaultProfile() *Profile {
	return &Profile{
		DefaultAction: ActionAllow, // Allow by default, block specific dangerous calls
		Architectures: []string{"SCMP_ARCH_X86_64", "SCMP_ARCH_AARCH64"},
		Syscalls: []SyscallRule{
			// Block syscalls that could escape the container
			{
				Names:  []string{"acct"},
				Action: ActionErrno,
			},
			{
				Names:  []string{"add_key"},
				Action: ActionErrno,
			},
			{
				Names:  []string{"bpf"},
				Action: ActionErrno,
			},
			{
				Names:  []string{"clock_adjtime"},
				Action: ActionErrno,
			},
			{
				Names:  []string{"clock_settime"},
				Action: ActionErrno,
			},
			{
				Names:  []string{"create_module"},
				Action: ActionErrno,
			},
			{
				Names:  []string{"delete_module"},
				Action: ActionErrno,
			},
			{
				Names:  []string{"finit_module"},
				Action: ActionErrno,
			},
			{
				Names:  []string{"get_kernel_syms"},
				Action: ActionErrno,
			},
			{
				Names:  []string{"init_module"},
				Action: ActionErrno,
			},
			{
				Names:  []string{"ioperm"},
				Action: ActionErrno,
			},
			{
				Names:  []string{"iopl"},
				Action: ActionErrno,
			},
			{
				Names:  []string{"kcmp"},
				Action: ActionErrno,
			},
			{
				Names:  []string{"kexec_file_load"},
				Action: ActionErrno,
			},
			{
				Names:  []string{"kexec_load"},
				Action: ActionErrno,
			},
			{
				Names:  []string{"keyctl"},
				Action: ActionErrno,
			},
			{
				Names:  []string{"lookup_dcookie"},
				Action: ActionErrno,
			},
			{
				Names:  []string{"mount"},
				Action: ActionErrno,
			},
			{
				Names:  []string{"move_pages"},
				Action: ActionErrno,
			},
			{
				Names:  []string{"nfsservctl"},
				Action: ActionErrno,
			},
			{
				Names:  []string{"open_by_handle_at"},
				Action: ActionErrno,
			},
			{
				Names:  []string{"perf_event_open"},
				Action: ActionErrno,
			},
			{
				Names:  []string{"personality"},
				Action: ActionErrno,
			},
			{
				Names:  []string{"pivot_root"},
				Action: ActionErrno,
			},
			{
				Names:  []string{"process_vm_readv"},
				Action: ActionErrno,
			},
			{
				Names:  []string{"process_vm_writev"},
				Action: ActionErrno,
			},
			{
				Names:  []string{"ptrace"},
				Action: ActionErrno,
			},
			{
				Names:  []string{"query_module"},
				Action: ActionErrno,
			},
			{
				Names:  []string{"quotactl"},
				Action: ActionErrno,
			},
			{
				Names:  []string{"reboot"},
				Action: ActionErrno,
			},
			{
				Names:  []string{"request_key"},
				Action: ActionErrno,
			},
			{
				Names:  []string{"setns"},
				Action: ActionErrno,
			},
			{
				Names:  []string{"settimeofday"},
				Action: ActionErrno,
			},
			{
				Names:  []string{"stime"},
				Action: ActionErrno,
			},
			{
				Names:  []string{"swapoff"},
				Action: ActionErrno,
			},
			{
				Names:  []string{"swapon"},
				Action: ActionErrno,
			},
			{
				Names:  []string{"sysfs"},
				Action: ActionErrno,
			},
			{
				Names:  []string{"_sysctl"},
				Action: ActionErrno,
			},
			{
				Names:  []string{"umount"},
				Action: ActionErrno,
			},
			{
				Names:  []string{"umount2"},
				Action: ActionErrno,
			},
			{
				Names:  []string{"unshare"},
				Action: ActionErrno,
			},
			{
				Names:  []string{"uselib"},
				Action: ActionErrno,
			},
			{
				Names:  []string{"userfaultfd"},
				Action: ActionErrno,
			},
			{
				Names:  []string{"ustat"},
				Action: ActionErrno,
			},
			{
				Names:  []string{"vm86"},
				Action: ActionErrno,
			},
			{
				Names:  []string{"vm86old"},
				Action: ActionErrno,
			},
		},
	}
}

// RestrictiveProfile returns a more restrictive profile that only allows
// essential syscalls. Use this for high-security containers.
func RestrictiveProfile() *Profile {
	return &Profile{
		DefaultAction: ActionErrno, // Block by default
		Architectures: []string{"SCMP_ARCH_X86_64", "SCMP_ARCH_AARCH64"},
		Syscalls: []SyscallRule{
			// Essential syscalls for basic operation
			{
				Names: []string{
					"read", "write", "open", "close", "stat", "fstat", "lstat",
					"poll", "lseek", "mmap", "mprotect", "munmap", "brk",
					"rt_sigaction", "rt_sigprocmask", "rt_sigreturn", "ioctl",
					"access", "pipe", "select", "sched_yield", "mremap",
					"msync", "mincore", "madvise", "dup", "dup2", "nanosleep",
					"getpid", "socket", "connect", "accept", "sendto", "recvfrom",
					"sendmsg", "recvmsg", "shutdown", "bind", "listen", "getsockname",
					"getpeername", "socketpair", "setsockopt", "getsockopt",
					"clone", "fork", "vfork", "execve", "exit", "wait4", "kill",
					"uname", "fcntl", "flock", "fsync", "fdatasync", "truncate",
					"ftruncate", "getdents", "getcwd", "chdir", "fchdir", "rename",
					"mkdir", "rmdir", "creat", "link", "unlink", "symlink",
					"readlink", "chmod", "fchmod", "chown", "fchown", "lchown",
					"umask", "gettimeofday", "getrlimit", "getrusage", "sysinfo",
					"times", "getuid", "getgid", "setuid", "setgid", "geteuid",
					"getegid", "setpgid", "getppid", "getpgrp", "setsid", "setreuid",
					"setregid", "getgroups", "setgroups", "setresuid", "getresuid",
					"setresgid", "getresgid", "getpgid", "setfsuid", "setfsgid",
					"getsid", "capget", "capset", "rt_sigpending", "rt_sigtimedwait",
					"rt_sigqueueinfo", "rt_sigsuspend", "sigaltstack", "statfs",
					"fstatfs", "prctl", "arch_prctl", "adjtimex", "setrlimit",
					"chroot", "sync", "mount", "umount2", "sethostname",
					"setdomainname", "ioperm", "iopl",
				},
				Action: ActionAllow,
			},
		},
	}
}
