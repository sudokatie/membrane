package seccomp

import (
	"testing"

	"github.com/sudokatie/membrane/pkg/oci"
)

func TestFromSpec(t *testing.T) {
	errnoRet := uint(1)
	spec := &oci.LinuxSeccomp{
		DefaultAction: oci.ActAllow,
		Architectures: []oci.Arch{oci.ArchX86_64, oci.ArchAARCH64},
		Syscalls: []oci.LinuxSyscall{
			{
				Names:    []string{"mount", "umount2"},
				Action:   oci.ActErrno,
				ErrnoRet: &errnoRet,
			},
			{
				Names:  []string{"ptrace"},
				Action: oci.ActKill,
			},
		},
	}

	profile := FromSpec(spec)

	if profile.DefaultAction != ActionAllow {
		t.Errorf("DefaultAction = %d, want ActionAllow", profile.DefaultAction)
	}

	if len(profile.Architectures) != 2 {
		t.Errorf("len(Architectures) = %d, want 2", len(profile.Architectures))
	}

	if len(profile.Syscalls) != 2 {
		t.Errorf("len(Syscalls) = %d, want 2", len(profile.Syscalls))
	}

	// Check first rule
	rule := profile.Syscalls[0]
	if len(rule.Names) != 2 {
		t.Errorf("len(rule.Names) = %d, want 2", len(rule.Names))
	}
	if rule.Action != ActionErrno {
		t.Errorf("rule.Action = %d, want ActionErrno", rule.Action)
	}
	if rule.ErrnoRet != 1 {
		t.Errorf("rule.ErrnoRet = %d, want 1", rule.ErrnoRet)
	}
}

func TestFromSpecNil(t *testing.T) {
	profile := FromSpec(nil)
	if profile != nil {
		t.Error("expected nil profile for nil spec")
	}
}

func TestFromSpecWithArgs(t *testing.T) {
	spec := &oci.LinuxSeccomp{
		DefaultAction: oci.ActKill,
		Syscalls: []oci.LinuxSyscall{
			{
				Names:  []string{"socket"},
				Action: oci.ActAllow,
				Args: []oci.LinuxSeccompArg{
					{
						Index: 0,
						Value: 2, // AF_INET
						Op:    oci.OpEqualTo,
					},
				},
			},
		},
	}

	profile := FromSpec(spec)

	if len(profile.Syscalls) != 1 {
		t.Fatalf("len(Syscalls) = %d, want 1", len(profile.Syscalls))
	}

	rule := profile.Syscalls[0]
	if len(rule.Args) != 1 {
		t.Fatalf("len(rule.Args) = %d, want 1", len(rule.Args))
	}

	arg := rule.Args[0]
	if arg.Index != 0 {
		t.Errorf("arg.Index = %d, want 0", arg.Index)
	}
	if arg.Value != 2 {
		t.Errorf("arg.Value = %d, want 2", arg.Value)
	}
	if arg.Op != OpEqualTo {
		t.Errorf("arg.Op = %d, want OpEqualTo", arg.Op)
	}
}

func TestActionFromOCI(t *testing.T) {
	tests := []struct {
		oci    oci.SeccompAction
		expect Action
	}{
		{oci.ActKill, ActionKill},
		{oci.ActTrap, ActionTrap},
		{oci.ActErrno, ActionErrno},
		{oci.ActTrace, ActionTrace},
		{oci.ActAllow, ActionAllow},
		{oci.ActLog, ActionLog},
	}

	for _, tt := range tests {
		t.Run(string(tt.oci), func(t *testing.T) {
			result := actionFromOCI(tt.oci)
			if result != tt.expect {
				t.Errorf("actionFromOCI(%s) = %d, want %d", tt.oci, result, tt.expect)
			}
		})
	}
}

func TestOperatorFromOCI(t *testing.T) {
	tests := []struct {
		oci    oci.SeccompOperator
		expect Operator
	}{
		{oci.OpNotEqual, OpNotEqual},
		{oci.OpLessThan, OpLessThan},
		{oci.OpLessEqual, OpLessEqual},
		{oci.OpEqualTo, OpEqualTo},
		{oci.OpGreaterEqual, OpGreaterEqual},
		{oci.OpGreaterThan, OpGreaterThan},
		{oci.OpMaskedEqual, OpMaskedEqual},
	}

	for _, tt := range tests {
		t.Run(string(tt.oci), func(t *testing.T) {
			result := operatorFromOCI(tt.oci)
			if result != tt.expect {
				t.Errorf("operatorFromOCI(%s) = %d, want %d", tt.oci, result, tt.expect)
			}
		})
	}
}

func TestDefaultProfile(t *testing.T) {
	profile := DefaultProfile()

	if profile == nil {
		t.Fatal("DefaultProfile() returned nil")
	}

	if profile.DefaultAction != ActionAllow {
		t.Errorf("DefaultAction = %d, want ActionAllow", profile.DefaultAction)
	}

	if len(profile.Architectures) == 0 {
		t.Error("expected architectures in default profile")
	}

	if len(profile.Syscalls) == 0 {
		t.Error("expected syscall rules in default profile")
	}

	// Check that dangerous syscalls are blocked
	dangerous := []string{"ptrace", "mount", "reboot", "kexec_load"}
	for _, name := range dangerous {
		found := false
		for _, rule := range profile.Syscalls {
			for _, n := range rule.Names {
				if n == name && rule.Action == ActionErrno {
					found = true
					break
				}
			}
		}
		if !found {
			t.Errorf("expected %s to be blocked in default profile", name)
		}
	}
}

func TestRestrictiveProfile(t *testing.T) {
	profile := RestrictiveProfile()

	if profile == nil {
		t.Fatal("RestrictiveProfile() returned nil")
	}

	if profile.DefaultAction != ActionErrno {
		t.Errorf("DefaultAction = %d, want ActionErrno (block by default)", profile.DefaultAction)
	}

	if len(profile.Syscalls) == 0 {
		t.Error("expected allowed syscalls in restrictive profile")
	}

	// Check that basic syscalls are allowed
	rule := profile.Syscalls[0]
	if rule.Action != ActionAllow {
		t.Error("expected allow action for basic syscalls")
	}

	// Check for essential syscalls
	essentials := []string{"read", "write", "open", "close", "exit"}
	for _, name := range essentials {
		found := false
		for _, n := range rule.Names {
			if n == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected %s in allowed syscalls", name)
		}
	}
}

func TestProfileStruct(t *testing.T) {
	profile := &Profile{
		DefaultAction: ActionAllow,
		Architectures: []string{"SCMP_ARCH_X86_64"},
		Syscalls: []SyscallRule{
			{
				Names:    []string{"mount"},
				Action:   ActionErrno,
				ErrnoRet: 1,
			},
		},
	}

	if profile.DefaultAction != ActionAllow {
		t.Error("DefaultAction not set correctly")
	}
	if len(profile.Syscalls) != 1 {
		t.Error("Syscalls not set correctly")
	}
}
