package volume

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBindMountValidate(t *testing.T) {
	tests := []struct {
		name    string
		mount   BindMount
		wantErr bool
	}{
		{
			name: "valid bind mount",
			mount: BindMount{
				Source: "/host/path",
				Target: "/container/path",
			},
			wantErr: false,
		},
		{
			name: "missing source",
			mount: BindMount{
				Target: "/container/path",
			},
			wantErr: true,
		},
		{
			name: "missing target",
			mount: BindMount{
				Source: "/host/path",
			},
			wantErr: true,
		},
		{
			name: "relative target",
			mount: BindMount{
				Source: "/host/path",
				Target: "relative/path",
			},
			wantErr: true,
		},
		{
			name: "read-only bind mount",
			mount: BindMount{
				Source:   "/host/path",
				Target:   "/container/path",
				ReadOnly: true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.mount.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestVolumeMountValidate(t *testing.T) {
	tests := []struct {
		name    string
		mount   VolumeMount
		wantErr bool
	}{
		{
			name: "valid volume mount",
			mount: VolumeMount{
				Name:   "myvolume",
				Target: "/data",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			mount: VolumeMount{
				Target: "/data",
			},
			wantErr: true,
		},
		{
			name: "missing target",
			mount: VolumeMount{
				Name: "myvolume",
			},
			wantErr: true,
		},
		{
			name: "relative target",
			mount: VolumeMount{
				Name:   "myvolume",
				Target: "data",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.mount.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestManagerCreate(t *testing.T) {
	tmpDir := t.TempDir()
	m, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	vol, err := m.Create("test-volume", nil)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if vol.Name != "test-volume" {
		t.Errorf("vol.Name = %q, want %q", vol.Name, "test-volume")
	}
	if vol.Driver != DriverLocal {
		t.Errorf("vol.Driver = %q, want %q", vol.Driver, DriverLocal)
	}

	// Check mountpoint exists
	if _, err := os.Stat(vol.Mountpoint); err != nil {
		t.Errorf("mountpoint should exist: %v", err)
	}
}

func TestManagerCreateDuplicate(t *testing.T) {
	tmpDir := t.TempDir()
	m, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	_, err = m.Create("dupe", nil)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	_, err = m.Create("dupe", nil)
	if err == nil {
		t.Error("Create() should fail for duplicate volume")
	}
}

func TestManagerGet(t *testing.T) {
	tmpDir := t.TempDir()
	m, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	_, err = m.Create("myvolume", nil)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	vol, err := m.Get("myvolume")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if vol.Name != "myvolume" {
		t.Errorf("vol.Name = %q, want %q", vol.Name, "myvolume")
	}

	_, err = m.Get("nonexistent")
	if err == nil {
		t.Error("Get() should fail for nonexistent volume")
	}
}

func TestManagerGetOrCreate(t *testing.T) {
	tmpDir := t.TempDir()
	m, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	// First call creates
	vol1, err := m.GetOrCreate("vol", nil)
	if err != nil {
		t.Fatalf("GetOrCreate() error = %v", err)
	}

	// Second call returns existing
	vol2, err := m.GetOrCreate("vol", nil)
	if err != nil {
		t.Fatalf("GetOrCreate() error = %v", err)
	}

	if vol1.Mountpoint != vol2.Mountpoint {
		t.Error("GetOrCreate() should return same volume")
	}
}

func TestManagerRemove(t *testing.T) {
	tmpDir := t.TempDir()
	m, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	vol, err := m.Create("removeme", nil)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	mountpoint := vol.Mountpoint

	err = m.Remove("removeme", false)
	if err != nil {
		t.Fatalf("Remove() error = %v", err)
	}

	// Volume should be gone
	_, err = m.Get("removeme")
	if err == nil {
		t.Error("Get() should fail after Remove()")
	}

	// Directory should be gone
	if _, err := os.Stat(mountpoint); !os.IsNotExist(err) {
		t.Error("mountpoint directory should be removed")
	}
}

func TestManagerRemoveNonexistent(t *testing.T) {
	tmpDir := t.TempDir()
	m, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	err = m.Remove("nonexistent", false)
	if err == nil {
		t.Error("Remove() should fail for nonexistent volume")
	}
}

func TestManagerList(t *testing.T) {
	tmpDir := t.TempDir()
	m, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	volumes := m.List()
	if len(volumes) != 0 {
		t.Errorf("List() = %d volumes, want 0", len(volumes))
	}

	m.Create("vol1", nil)
	m.Create("vol2", nil)

	volumes = m.List()
	if len(volumes) != 2 {
		t.Errorf("List() = %d volumes, want 2", len(volumes))
	}
}

func TestManagerPersistence(t *testing.T) {
	tmpDir := t.TempDir()

	// Create manager and volume
	m1, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	_, err = m1.Create("persistent", nil)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Create new manager pointing to same directory
	m2, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	// Volume should be loaded
	vol, err := m2.Get("persistent")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if vol.Name != "persistent" {
		t.Errorf("vol.Name = %q, want %q", vol.Name, "persistent")
	}
}

func TestVolumeWithOptions(t *testing.T) {
	tmpDir := t.TempDir()
	m, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	opts := map[string]string{
		"size": "10G",
		"type": "ext4",
	}

	vol, err := m.Create("with-opts", opts)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if vol.Options["size"] != "10G" {
		t.Errorf("vol.Options[size] = %q, want %q", vol.Options["size"], "10G")
	}
}

func TestMountpointPath(t *testing.T) {
	tmpDir := t.TempDir()
	m, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	vol, err := m.Create("pathtest", nil)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	expected := filepath.Join(tmpDir, "pathtest", "_data")
	if vol.Mountpoint != expected {
		t.Errorf("Mountpoint = %q, want %q", vol.Mountpoint, expected)
	}
}
