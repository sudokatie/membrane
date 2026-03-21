package filesystem

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sudokatie/membrane/pkg/oci"
)

func TestFromSpec(t *testing.T) {
	bundle := t.TempDir()
	rootfs := filepath.Join(bundle, "rootfs")
	os.MkdirAll(rootfs, 0755)

	spec := &oci.Spec{
		Root: &oci.Root{
			Path: "rootfs",
		},
		Mounts: []oci.Mount{
			{
				Destination: "/proc",
				Type:        "proc",
				Source:      "proc",
				Options:     []string{"nosuid", "noexec", "nodev"},
			},
			{
				Destination: "/tmp",
				Type:        "tmpfs",
				Source:      "tmpfs",
				Options:     []string{"nosuid", "mode=1777"},
			},
		},
	}

	config := FromSpec(spec, bundle)

	if config.RootfsPath != rootfs {
		t.Errorf("RootfsPath = %s, want %s", config.RootfsPath, rootfs)
	}

	if len(config.Mounts) != 2 {
		t.Errorf("expected 2 mounts, got %d", len(config.Mounts))
	}

	// Check first mount
	if config.Mounts[0].Target != "/proc" {
		t.Errorf("mount[0].Target = %s, want /proc", config.Mounts[0].Target)
	}
	if config.Mounts[0].FSType != "proc" {
		t.Errorf("mount[0].FSType = %s, want proc", config.Mounts[0].FSType)
	}
}

func TestFromSpecAbsolutePath(t *testing.T) {
	bundle := t.TempDir()
	rootfs := filepath.Join(bundle, "rootfs")
	os.MkdirAll(rootfs, 0755)

	spec := &oci.Spec{
		Root: &oci.Root{
			Path: rootfs, // Absolute path
		},
	}

	config := FromSpec(spec, bundle)

	if config.RootfsPath != rootfs {
		t.Errorf("RootfsPath = %s, want %s", config.RootfsPath, rootfs)
	}
}

func TestPrepareRootfs(t *testing.T) {
	rootfs := t.TempDir()

	if err := PrepareRootfs(rootfs); err != nil {
		t.Errorf("PrepareRootfs failed: %v", err)
	}
}

func TestPrepareRootfsNotFound(t *testing.T) {
	if err := PrepareRootfs("/nonexistent/path"); err == nil {
		t.Error("expected error for nonexistent rootfs")
	}
}

func TestCreateMountpoint(t *testing.T) {
	rootfs := t.TempDir()

	if err := CreateMountpoint(rootfs, "/proc"); err != nil {
		t.Errorf("CreateMountpoint failed: %v", err)
	}

	path := filepath.Join(rootfs, "proc")
	info, err := os.Stat(path)
	if err != nil {
		t.Errorf("mountpoint not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("mountpoint is not a directory")
	}
}

func TestCreateMountpointNested(t *testing.T) {
	rootfs := t.TempDir()

	if err := CreateMountpoint(rootfs, "/a/b/c"); err != nil {
		t.Errorf("CreateMountpoint failed: %v", err)
	}

	path := filepath.Join(rootfs, "a", "b", "c")
	if _, err := os.Stat(path); err != nil {
		t.Errorf("nested mountpoint not created: %v", err)
	}
}

func TestDefaultMounts(t *testing.T) {
	mounts := DefaultMounts()

	if len(mounts) == 0 {
		t.Error("expected default mounts")
	}

	// Check for expected mounts
	targets := make(map[string]bool)
	for _, m := range mounts {
		targets[m.Target] = true
	}

	expected := []string{"/proc", "/dev", "/dev/pts", "/dev/shm", "/sys"}
	for _, e := range expected {
		if !targets[e] {
			t.Errorf("expected mount for %s", e)
		}
	}
}

func TestMountFlagsMap(t *testing.T) {
	if MountFlags["ro"] == 0 {
		t.Error("expected non-zero ro flag")
	}
	if MountFlags["nosuid"] == 0 {
		t.Error("expected non-zero nosuid flag")
	}
	if MountFlags["bind"] == 0 {
		t.Error("expected non-zero bind flag")
	}
}

func TestMountStruct(t *testing.T) {
	m := Mount{
		Source: "proc",
		Target: "/proc",
		FSType: "proc",
		Flags:  MountFlags["nosuid"] | MountFlags["noexec"],
		Data:   "",
	}

	if m.Source != "proc" {
		t.Errorf("Source = %s, want proc", m.Source)
	}
	if m.Flags == 0 {
		t.Error("expected non-zero flags")
	}
}
