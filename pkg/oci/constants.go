package oci

import "syscall"

// OCI runtime spec version.
const (
	Version = "1.0.2"
)

// Signal names to numbers.
var Signals = map[string]syscall.Signal{
	"SIGHUP":    syscall.SIGHUP,
	"SIGINT":    syscall.SIGINT,
	"SIGQUIT":   syscall.SIGQUIT,
	"SIGILL":    syscall.SIGILL,
	"SIGTRAP":   syscall.SIGTRAP,
	"SIGABRT":   syscall.SIGABRT,
	"SIGBUS":    syscall.SIGBUS,
	"SIGFPE":    syscall.SIGFPE,
	"SIGKILL":   syscall.SIGKILL,
	"SIGUSR1":   syscall.SIGUSR1,
	"SIGSEGV":   syscall.SIGSEGV,
	"SIGUSR2":   syscall.SIGUSR2,
	"SIGPIPE":   syscall.SIGPIPE,
	"SIGALRM":   syscall.SIGALRM,
	"SIGTERM":   syscall.SIGTERM,
	"SIGCHLD":   syscall.SIGCHLD,
	"SIGCONT":   syscall.SIGCONT,
	"SIGSTOP":   syscall.SIGSTOP,
	"SIGTSTP":   syscall.SIGTSTP,
	"SIGTTIN":   syscall.SIGTTIN,
	"SIGTTOU":   syscall.SIGTTOU,
	"SIGURG":    syscall.SIGURG,
	"SIGXCPU":   syscall.SIGXCPU,
	"SIGXFSZ":   syscall.SIGXFSZ,
	"SIGVTALRM": syscall.SIGVTALRM,
	"SIGPROF":   syscall.SIGPROF,
	"SIGWINCH":  syscall.SIGWINCH,
	"SIGIO":     syscall.SIGIO,
	"SIGSYS":    syscall.SIGSYS,
}

// ParseMountFlags parses mount options to flags and data.
func ParseMountFlags(options []string) (flags uintptr, data string) {
	var dataOpts []string
	for _, opt := range options {
		if f, ok := MountFlags[opt]; ok {
			flags |= f
		} else {
			dataOpts = append(dataOpts, opt)
		}
	}
	if len(dataOpts) > 0 {
		for i, opt := range dataOpts {
			if i > 0 {
				data += ","
			}
			data += opt
		}
	}
	return flags, data
}

// DefaultMaskedPaths are paths masked by default.
var DefaultMaskedPaths = []string{
	"/proc/acpi",
	"/proc/asound",
	"/proc/kcore",
	"/proc/keys",
	"/proc/latency_stats",
	"/proc/timer_list",
	"/proc/timer_stats",
	"/proc/sched_debug",
	"/proc/scsi",
	"/sys/firmware",
}

// DefaultReadonlyPaths are paths made read-only by default.
var DefaultReadonlyPaths = []string{
	"/proc/bus",
	"/proc/fs",
	"/proc/irq",
	"/proc/sys",
	"/proc/sysrq-trigger",
}
