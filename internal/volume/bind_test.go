package volume

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParsePropagation(t *testing.T) {
	tests := []struct {
		input   string
		want    PropagationMode
		wantErr bool
	}{
		{"", PropagationPrivate, false},
		{"private", PropagationPrivate, false},
		{"rprivate", PropagationRPrivate, false},
		{"slave", PropagationSlave, false},
		{"rslave", PropagationRSlave, false},
		{"shared", PropagationShared, false},
		{"rshared", PropagationRShared, false},
		{"invalid", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParsePropagation(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePropagation(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParsePropagation(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestBindMountOptionsValidate(t *testing.T) {
	tests := []struct {
		name    string
		opts    BindMountOptions
		wantErr bool
	}{
		{
			name: "valid",
			opts: BindMountOptions{
				Source: "/host/path",
				Target: "/container/path",
			},
			wantErr: false,
		},
		{
			name: "missing source",
			opts: BindMountOptions{
				Target: "/container/path",
			},
			wantErr: true,
		},
		{
			name: "missing target",
			opts: BindMountOptions{
				Source: "/host/path",
			},
			wantErr: true,
		},
		{
			name: "relative target",
			opts: BindMountOptions{
				Source: "/host/path",
				Target: "relative",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.opts.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBindMountOptionsFlags(t *testing.T) {
	tests := []struct {
		name     string
		opts     BindMountOptions
		wantFlag uintptr
	}{
		{
			name:     "basic bind",
			opts:     BindMountOptions{Source: "/a", Target: "/b"},
			wantFlag: 4096, // MS_BIND
		},
		{
			name:     "read-only bind",
			opts:     BindMountOptions{Source: "/a", Target: "/b", ReadOnly: true},
			wantFlag: 4096 | 1, // MS_BIND | MS_RDONLY
		},
		{
			name:     "private propagation",
			opts:     BindMountOptions{Source: "/a", Target: "/b", Propagation: PropagationPrivate},
			wantFlag: 4096 | 262144, // MS_BIND | MS_PRIVATE
		},
		{
			name:     "rprivate propagation",
			opts:     BindMountOptions{Source: "/a", Target: "/b", Propagation: PropagationRPrivate},
			wantFlag: 4096 | 262144 | 16384, // MS_BIND | MS_PRIVATE | MS_REC
		},
		{
			name:     "slave propagation",
			opts:     BindMountOptions{Source: "/a", Target: "/b", Propagation: PropagationSlave},
			wantFlag: 4096 | 524288, // MS_BIND | MS_SLAVE
		},
		{
			name:     "shared propagation",
			opts:     BindMountOptions{Source: "/a", Target: "/b", Propagation: PropagationShared},
			wantFlag: 4096 | 1048576, // MS_BIND | MS_SHARED
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.opts.MountFlags()
			if got != tt.wantFlag {
				t.Errorf("MountFlags() = %d, want %d", got, tt.wantFlag)
			}
		})
	}
}

func TestPrepareBindMount(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a source directory
	srcDir := filepath.Join(tmpDir, "source")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		opts    BindMountOptions
		wantErr bool
	}{
		{
			name: "existing source",
			opts: BindMountOptions{
				Source: srcDir,
				Target: "/container/data",
			},
			wantErr: false,
		},
		{
			name: "nonexistent source without create",
			opts: BindMountOptions{
				Source:       filepath.Join(tmpDir, "nonexistent"),
				Target:       "/container/data",
				CreateSource: false,
			},
			wantErr: true,
		},
		{
			name: "nonexistent source with create",
			opts: BindMountOptions{
				Source:       filepath.Join(tmpDir, "newdir"),
				Target:       "/container/data",
				CreateSource: true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := PrepareBindMount(&tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("PrepareBindMount() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPrepareBindTarget(t *testing.T) {
	tmpDir := t.TempDir()
	rootfs := filepath.Join(tmpDir, "rootfs")
	if err := os.MkdirAll(rootfs, 0755); err != nil {
		t.Fatal(err)
	}

	// Test directory target
	err := PrepareBindTarget(rootfs, "/data/subdir", true)
	if err != nil {
		t.Fatalf("PrepareBindTarget() for dir error = %v", err)
	}
	
	targetDir := filepath.Join(rootfs, "data", "subdir")
	if info, err := os.Stat(targetDir); err != nil || !info.IsDir() {
		t.Error("directory target should be created")
	}

	// Test file target
	err = PrepareBindTarget(rootfs, "/etc/config.txt", false)
	if err != nil {
		t.Fatalf("PrepareBindTarget() for file error = %v", err)
	}

	targetFile := filepath.Join(rootfs, "etc", "config.txt")
	if info, err := os.Stat(targetFile); err != nil || info.IsDir() {
		t.Error("file target should be created")
	}
}

func TestIsDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a directory
	dir := filepath.Join(tmpDir, "testdir")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a file
	file := filepath.Join(tmpDir, "testfile")
	if err := os.WriteFile(file, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		path      string
		wantIsDir bool
		wantErr   bool
	}{
		{dir, true, false},
		{file, false, false},
		{filepath.Join(tmpDir, "nonexistent"), false, true},
	}

	for _, tt := range tests {
		t.Run(filepath.Base(tt.path), func(t *testing.T) {
			isDir, err := IsDirectory(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsDirectory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && isDir != tt.wantIsDir {
				t.Errorf("IsDirectory() = %v, want %v", isDir, tt.wantIsDir)
			}
		})
	}
}
