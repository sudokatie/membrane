// Package seccomp provides syscall filtering using seccomp-bpf.
package seccomp

import (
	"github.com/sudokatie/membrane/pkg/oci"
)

// Profile represents a seccomp filter profile.
type Profile struct {
	// DefaultAction is the action for syscalls not in the rules.
	DefaultAction Action
	// Architectures are the allowed architectures.
	Architectures []string
	// Syscalls are the syscall rules.
	Syscalls []SyscallRule
}

// Action represents a seccomp action.
type Action uint32

// Seccomp actions.
const (
	ActionKill  Action = 0x00000000 // SECCOMP_RET_KILL_THREAD
	ActionTrap  Action = 0x00030000 // SECCOMP_RET_TRAP
	ActionErrno Action = 0x00050000 // SECCOMP_RET_ERRNO
	ActionTrace Action = 0x7ff00000 // SECCOMP_RET_TRACE
	ActionAllow Action = 0x7fff0000 // SECCOMP_RET_ALLOW
	ActionLog   Action = 0x7ffc0000 // SECCOMP_RET_LOG
)

// SyscallRule represents a rule for a set of syscalls.
type SyscallRule struct {
	// Names are the syscall names.
	Names []string
	// Action is the action for these syscalls.
	Action Action
	// ErrnoRet is the errno to return (for ActionErrno).
	ErrnoRet uint
	// Args are argument filters.
	Args []ArgFilter
}

// ArgFilter represents a syscall argument filter.
type ArgFilter struct {
	// Index is the argument index (0-5).
	Index uint
	// Value is the value to compare.
	Value uint64
	// ValueTwo is the second value (for masked compare).
	ValueTwo uint64
	// Op is the comparison operator.
	Op Operator
}

// Operator is a comparison operator.
type Operator uint

// Comparison operators.
const (
	OpNotEqual     Operator = 1
	OpLessThan     Operator = 2
	OpLessEqual    Operator = 3
	OpEqualTo      Operator = 4
	OpGreaterEqual Operator = 5
	OpGreaterThan  Operator = 6
	OpMaskedEqual  Operator = 7
)

// FromSpec creates a Profile from an OCI LinuxSeccomp spec.
func FromSpec(spec *oci.LinuxSeccomp) *Profile {
	if spec == nil {
		return nil
	}

	profile := &Profile{
		DefaultAction: actionFromOCI(spec.DefaultAction),
	}

	// Convert architectures
	for _, arch := range spec.Architectures {
		profile.Architectures = append(profile.Architectures, string(arch))
	}

	// Convert syscall rules
	for _, sc := range spec.Syscalls {
		rule := SyscallRule{
			Names:  sc.Names,
			Action: actionFromOCI(sc.Action),
		}
		if sc.ErrnoRet != nil {
			rule.ErrnoRet = *sc.ErrnoRet
		}
		for _, arg := range sc.Args {
			rule.Args = append(rule.Args, ArgFilter{
				Index:    arg.Index,
				Value:    arg.Value,
				ValueTwo: arg.ValueTwo,
				Op:       operatorFromOCI(arg.Op),
			})
		}
		profile.Syscalls = append(profile.Syscalls, rule)
	}

	return profile
}

// actionFromOCI converts an OCI seccomp action to our Action type.
func actionFromOCI(a oci.SeccompAction) Action {
	switch a {
	case oci.ActKill:
		return ActionKill
	case oci.ActTrap:
		return ActionTrap
	case oci.ActErrno:
		return ActionErrno
	case oci.ActTrace:
		return ActionTrace
	case oci.ActAllow:
		return ActionAllow
	case oci.ActLog:
		return ActionLog
	default:
		return ActionKill
	}
}

// operatorFromOCI converts an OCI seccomp operator to our Operator type.
func operatorFromOCI(op oci.SeccompOperator) Operator {
	switch op {
	case oci.OpNotEqual:
		return OpNotEqual
	case oci.OpLessThan:
		return OpLessThan
	case oci.OpLessEqual:
		return OpLessEqual
	case oci.OpEqualTo:
		return OpEqualTo
	case oci.OpGreaterEqual:
		return OpGreaterEqual
	case oci.OpGreaterThan:
		return OpGreaterThan
	case oci.OpMaskedEqual:
		return OpMaskedEqual
	default:
		return OpEqualTo
	}
}
