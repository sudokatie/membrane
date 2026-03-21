package namespace

import (
	"testing"

	"github.com/sudokatie/membrane/pkg/oci"
)

func TestFromSpec(t *testing.T) {
	linux := &oci.Linux{
		Namespaces: []oci.LinuxNamespace{
			{Type: oci.PIDNamespace},
			{Type: oci.MountNamespace},
			{Type: oci.UTSNamespace, Path: "/proc/1/ns/uts"},
		},
	}

	config := FromSpec(linux)

	if len(config.Namespaces) != 3 {
		t.Errorf("expected 3 namespaces, got %d", len(config.Namespaces))
	}

	if config.Namespaces[0].Type != oci.PIDNamespace {
		t.Errorf("expected PID namespace first")
	}

	if config.Namespaces[2].Path != "/proc/1/ns/uts" {
		t.Errorf("expected UTS namespace path")
	}
}

func TestFromSpecNil(t *testing.T) {
	config := FromSpec(nil)
	if config == nil {
		t.Fatal("expected non-nil config")
	}
	if len(config.Namespaces) != 0 {
		t.Errorf("expected 0 namespaces, got %d", len(config.Namespaces))
	}
}

func TestCloneFlags(t *testing.T) {
	config := &Config{
		Namespaces: []Namespace{
			{Type: oci.PIDNamespace},
			{Type: oci.MountNamespace},
			{Type: oci.UTSNamespace},
		},
	}

	flags := config.CloneFlags()

	// Check that flags are set
	if flags == 0 {
		t.Error("expected non-zero flags")
	}

	// Each namespace should contribute its flag
	expectedFlags := oci.CloneFlags[oci.PIDNamespace] |
		oci.CloneFlags[oci.MountNamespace] |
		oci.CloneFlags[oci.UTSNamespace]

	if flags != expectedFlags {
		t.Errorf("flags = %x, want %x", flags, expectedFlags)
	}
}

func TestCloneFlagsWithPath(t *testing.T) {
	config := &Config{
		Namespaces: []Namespace{
			{Type: oci.PIDNamespace},
			{Type: oci.UTSNamespace, Path: "/proc/1/ns/uts"}, // Join existing
		},
	}

	flags := config.CloneFlags()

	// Only PID namespace should contribute (UTS has path)
	expectedFlags := oci.CloneFlags[oci.PIDNamespace]

	if flags != expectedFlags {
		t.Errorf("flags = %x, want %x", flags, expectedFlags)
	}
}

func TestHasNamespace(t *testing.T) {
	config := &Config{
		Namespaces: []Namespace{
			{Type: oci.PIDNamespace},
			{Type: oci.MountNamespace},
		},
	}

	if !config.HasNamespace(oci.PIDNamespace) {
		t.Error("expected to have PID namespace")
	}
	if !config.HasNamespace(oci.MountNamespace) {
		t.Error("expected to have mount namespace")
	}
	if config.HasNamespace(oci.UTSNamespace) {
		t.Error("expected not to have UTS namespace")
	}
}

func TestGetNamespace(t *testing.T) {
	config := &Config{
		Namespaces: []Namespace{
			{Type: oci.PIDNamespace},
			{Type: oci.UTSNamespace, Path: "/path/to/ns"},
		},
	}

	ns := config.GetNamespace(oci.UTSNamespace)
	if ns == nil {
		t.Fatal("expected to find UTS namespace")
	}
	if ns.Path != "/path/to/ns" {
		t.Errorf("path = %s, want /path/to/ns", ns.Path)
	}

	ns = config.GetNamespace(oci.NetworkNamespace)
	if ns != nil {
		t.Error("expected nil for missing namespace")
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid",
			config: &Config{
				Namespaces: []Namespace{
					{Type: oci.PIDNamespace},
					{Type: oci.MountNamespace},
				},
			},
			wantErr: false,
		},
		{
			name: "duplicate namespace",
			config: &Config{
				Namespaces: []Namespace{
					{Type: oci.PIDNamespace},
					{Type: oci.PIDNamespace},
				},
			},
			wantErr: true,
		},
		{
			name: "user namespace first",
			config: &Config{
				Namespaces: []Namespace{
					{Type: oci.UserNamespace},
					{Type: oci.PIDNamespace},
				},
			},
			wantErr: false,
		},
		{
			name: "user namespace not first",
			config: &Config{
				Namespaces: []Namespace{
					{Type: oci.PIDNamespace},
					{Type: oci.UserNamespace},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSortForClone(t *testing.T) {
	config := &Config{
		Namespaces: []Namespace{
			{Type: oci.PIDNamespace},
			{Type: oci.UserNamespace},
			{Type: oci.MountNamespace},
		},
	}

	config.SortForClone()

	if config.Namespaces[0].Type != oci.UserNamespace {
		t.Errorf("expected user namespace first after sort, got %s", config.Namespaces[0].Type)
	}
}

func TestSortForCloneNoUser(t *testing.T) {
	config := &Config{
		Namespaces: []Namespace{
			{Type: oci.PIDNamespace},
			{Type: oci.MountNamespace},
		},
	}

	original := config.Namespaces[0].Type
	config.SortForClone()

	if config.Namespaces[0].Type != original {
		t.Error("sort should not change order when no user namespace")
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if len(config.Namespaces) == 0 {
		t.Error("default config should have namespaces")
	}

	if !config.HasNamespace(oci.PIDNamespace) {
		t.Error("default config should have PID namespace")
	}
	if !config.HasNamespace(oci.MountNamespace) {
		t.Error("default config should have mount namespace")
	}
	if !config.HasNamespace(oci.IPCNamespace) {
		t.Error("default config should have IPC namespace")
	}
	if !config.HasNamespace(oci.UTSNamespace) {
		t.Error("default config should have UTS namespace")
	}
	if !config.HasNamespace(oci.NetworkNamespace) {
		t.Error("default config should have network namespace")
	}
}

func TestGetNamespacePath(t *testing.T) {
	path := GetNamespacePath(1234, oci.PIDNamespace)
	// Just check it returns something reasonable
	if path == "" {
		t.Error("expected non-empty path")
	}
}
