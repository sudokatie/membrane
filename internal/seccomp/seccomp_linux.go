//go:build linux

package seccomp

import (
	"encoding/binary"
	"fmt"
	"unsafe"

	"golang.org/x/sys/unix"
)

// BPF instruction structure
type bpfInstruction struct {
	Code uint16
	Jt   uint8
	Jf   uint8
	K    uint32
}

// BPF instruction codes
const (
	bpfLD   = 0x00
	bpfJMP  = 0x05
	bpfRET  = 0x06
	bpfABS  = 0x20
	bpfJEQ  = 0x10
	bpfK    = 0x00
	bpfW    = 0x00
)

// seccomp data offsets
const (
	offsetNr   = 0  // syscall number
	offsetArch = 4  // architecture
)

// Architecture constants
const (
	auditArchX86_64  = 0xc000003e
	auditArchAarch64 = 0xc00000b7
)

// LoadFilter loads and applies a seccomp filter.
func LoadFilter(profile *Profile) error {
	if profile == nil {
		return nil
	}

	// Build BPF program
	program, err := buildBPF(profile)
	if err != nil {
		return fmt.Errorf("build bpf: %w", err)
	}

	// Apply the filter
	prog := unix.SockFprog{
		Len:    uint16(len(program)),
		Filter: (*unix.SockFilter)(unsafe.Pointer(&program[0])),
	}

	// Set no_new_privs to allow seccomp without CAP_SYS_ADMIN
	if err := unix.Prctl(unix.PR_SET_NO_NEW_PRIVS, 1, 0, 0, 0); err != nil {
		return fmt.Errorf("prctl no_new_privs: %w", err)
	}

	// Apply seccomp filter
	if err := unix.Prctl(unix.PR_SET_SECCOMP, unix.SECCOMP_MODE_FILTER,
		uintptr(unsafe.Pointer(&prog)), 0, 0); err != nil {
		return fmt.Errorf("prctl seccomp: %w", err)
	}

	return nil
}

// buildBPF builds a BPF program from a profile.
func buildBPF(profile *Profile) ([]bpfInstruction, error) {
	var program []bpfInstruction

	// Load architecture
	program = append(program, bpfInstruction{
		Code: bpfLD | bpfW | bpfABS,
		K:    offsetArch,
	})

	// Check architecture (x86_64)
	program = append(program, bpfInstruction{
		Code: bpfJMP | bpfJEQ | bpfK,
		K:    auditArchX86_64,
		Jt:   1, // continue
		Jf:   0, // check next arch
	})

	// If no architecture matches, use default action
	// (simplified: we just continue for now)

	// Load syscall number
	program = append(program, bpfInstruction{
		Code: bpfLD | bpfW | bpfABS,
		K:    offsetNr,
	})

	// Add rules for each syscall
	for _, rule := range profile.Syscalls {
		syscallNrs := getSyscallNumbers(rule.Names)
		action := uint32(rule.Action)
		if rule.Action == ActionErrno && rule.ErrnoRet > 0 {
			action |= uint32(rule.ErrnoRet) & 0xffff
		}

		for _, nr := range syscallNrs {
			// Jump if equal to this syscall
			program = append(program, bpfInstruction{
				Code: bpfJMP | bpfJEQ | bpfK,
				K:    uint32(nr),
				Jt:   0, // return action
				Jf:   1, // continue checking
			})
			// Return action for this syscall
			program = append(program, bpfInstruction{
				Code: bpfRET | bpfK,
				K:    action,
			})
		}
	}

	// Default action
	program = append(program, bpfInstruction{
		Code: bpfRET | bpfK,
		K:    uint32(profile.DefaultAction),
	})

	return program, nil
}

// getSyscallNumbers returns syscall numbers for the given names.
// This is a simplified version - a full implementation would use
// a syscall name to number mapping.
func getSyscallNumbers(names []string) []int {
	var numbers []int
	for _, name := range names {
		if nr, ok := syscallMap[name]; ok {
			numbers = append(numbers, nr)
		}
	}
	return numbers
}

// syscallMap maps syscall names to numbers (x86_64).
// This is a partial list of common syscalls.
var syscallMap = map[string]int{
	"read":               0,
	"write":              1,
	"open":               2,
	"close":              3,
	"stat":               4,
	"fstat":              5,
	"lstat":              6,
	"poll":               7,
	"lseek":              8,
	"mmap":               9,
	"mprotect":           10,
	"munmap":             11,
	"brk":                12,
	"rt_sigaction":       13,
	"rt_sigprocmask":     14,
	"rt_sigreturn":       15,
	"ioctl":              16,
	"access":             21,
	"pipe":               22,
	"select":             23,
	"sched_yield":        24,
	"mremap":             25,
	"msync":              26,
	"mincore":            27,
	"madvise":            28,
	"dup":                32,
	"dup2":               33,
	"nanosleep":          35,
	"getpid":             39,
	"socket":             41,
	"connect":            42,
	"accept":             43,
	"sendto":             44,
	"recvfrom":           45,
	"sendmsg":            46,
	"recvmsg":            47,
	"shutdown":           48,
	"bind":               49,
	"listen":             50,
	"getsockname":        51,
	"getpeername":        52,
	"clone":              56,
	"fork":               57,
	"vfork":              58,
	"execve":             59,
	"exit":               60,
	"wait4":              61,
	"kill":               62,
	"uname":              63,
	"fcntl":              72,
	"flock":              73,
	"fsync":              74,
	"fdatasync":          75,
	"truncate":           76,
	"ftruncate":          77,
	"getdents":           78,
	"getcwd":             79,
	"chdir":              80,
	"fchdir":             81,
	"rename":             82,
	"mkdir":              83,
	"rmdir":              84,
	"creat":              85,
	"link":               86,
	"unlink":             87,
	"symlink":            88,
	"readlink":           89,
	"chmod":              90,
	"fchmod":             91,
	"chown":              92,
	"fchown":             93,
	"lchown":             94,
	"umask":              95,
	"gettimeofday":       96,
	"getuid":             102,
	"getgid":             104,
	"setuid":             105,
	"setgid":             106,
	"geteuid":            107,
	"getegid":            108,
	"setpgid":            109,
	"getppid":            110,
	"getpgrp":            111,
	"setsid":             112,
	"setgroups":          116,
	"prctl":              157,
	"arch_prctl":         158,
	"mount":              165,
	"umount2":            166,
	"pivot_root":         155,
	"reboot":             169,
	"sethostname":        170,
	"setdomainname":      171,
	"init_module":        175,
	"delete_module":      176,
	"quotactl":           179,
	"acct":               163,
	"settimeofday":       164,
	"swapon":             167,
	"swapoff":            168,
	"ptrace":             101,
	"bpf":                321,
	"unshare":            272,
	"setns":              308,
	"kexec_load":         246,
	"kexec_file_load":    320,
	"finit_module":       313,
	"open_by_handle_at":  304,
	"perf_event_open":    298,
	"userfaultfd":        323,
}

// Ensure binary is imported for potential future use
var _ = binary.LittleEndian
