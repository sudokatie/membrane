package oci

import (
	"syscall"
	"testing"
)

func TestParseMountFlags(t *testing.T) {
	tests := []struct {
		name      string
		options   []string
		wantFlags uintptr
		wantData  string
	}{
		{
			name:      "empty",
			options:   nil,
			wantFlags: 0,
			wantData:  "",
		},
		{
			name:      "single flag",
			options:   []string{"ro"},
			wantFlags: MountFlags["ro"],
			wantData:  "",
		},
		{
			name:      "multiple flags",
			options:   []string{"ro", "nosuid", "nodev"},
			wantFlags: MountFlags["ro"] | MountFlags["nosuid"] | MountFlags["nodev"],
			wantData:  "",
		},
		{
			name:      "data option",
			options:   []string{"mode=755"},
			wantFlags: 0,
			wantData:  "mode=755",
		},
		{
			name:      "mixed flags and data",
			options:   []string{"ro", "nosuid", "mode=755", "size=100m"},
			wantFlags: MountFlags["ro"] | MountFlags["nosuid"],
			wantData:  "mode=755,size=100m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags, data := ParseMountFlags(tt.options)
			if flags != tt.wantFlags {
				t.Errorf("flags = %v, want %v", flags, tt.wantFlags)
			}
			if data != tt.wantData {
				t.Errorf("data = %q, want %q", data, tt.wantData)
			}
		})
	}
}

func TestSignals(t *testing.T) {
	if sig, ok := Signals["SIGTERM"]; !ok {
		t.Error("SIGTERM not in Signals map")
	} else if sig != syscall.SIGTERM {
		t.Errorf("SIGTERM = %v, want %v", sig, syscall.SIGTERM)
	}

	if sig, ok := Signals["SIGKILL"]; !ok {
		t.Error("SIGKILL not in Signals map")
	} else if sig != syscall.SIGKILL {
		t.Errorf("SIGKILL = %v, want %v", sig, syscall.SIGKILL)
	}
}

func TestCloneFlags(t *testing.T) {
	if _, ok := CloneFlags[PIDNamespace]; !ok {
		t.Error("PIDNamespace not in CloneFlags map")
	}
	if _, ok := CloneFlags[MountNamespace]; !ok {
		t.Error("MountNamespace not in CloneFlags map")
	}
	if _, ok := CloneFlags[NetworkNamespace]; !ok {
		t.Error("NetworkNamespace not in CloneFlags map")
	}
}

func TestDefaultPaths(t *testing.T) {
	if len(DefaultMaskedPaths) == 0 {
		t.Error("DefaultMaskedPaths is empty")
	}
	if len(DefaultReadonlyPaths) == 0 {
		t.Error("DefaultReadonlyPaths is empty")
	}

	// Check expected entries
	found := false
	for _, p := range DefaultMaskedPaths {
		if p == "/proc/kcore" {
			found = true
			break
		}
	}
	if !found {
		t.Error("/proc/kcore not in DefaultMaskedPaths")
	}
}
