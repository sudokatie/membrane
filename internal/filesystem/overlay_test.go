package filesystem

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateOverlayConfig(t *testing.T) {
	// Create temp dirs for testing
	tmpDir, err := os.MkdirTemp("", "overlay-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	lowerDir := filepath.Join(tmpDir, "lower")
	if err := os.MkdirAll(lowerDir, 0755); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		config  *OverlayConfig
		wantErr bool
	}{
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
		},
		{
			name: "empty lower dirs",
			config: &OverlayConfig{
				LowerDirs: []string{},
				UpperDir:  "/tmp/upper",
				WorkDir:   "/tmp/work",
				MergedDir: "/tmp/merged",
			},
			wantErr: true,
		},
		{
			name: "missing upper dir",
			config: &OverlayConfig{
				LowerDirs: []string{lowerDir},
				UpperDir:  "",
				WorkDir:   "/tmp/work",
				MergedDir: "/tmp/merged",
			},
			wantErr: true,
		},
		{
			name: "valid config",
			config: &OverlayConfig{
				LowerDirs: []string{lowerDir},
				UpperDir:  filepath.Join(tmpDir, "upper"),
				WorkDir:   filepath.Join(tmpDir, "work"),
				MergedDir: filepath.Join(tmpDir, "merged"),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateOverlayConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateOverlayConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBuildOverlayOptions(t *testing.T) {
	config := &OverlayConfig{
		LowerDirs: []string{"/layer1", "/layer2"},
		UpperDir:  "/upper",
		WorkDir:   "/work",
		MergedDir: "/merged",
	}

	options := BuildOverlayOptions(config)
	expected := "lowerdir=/layer1:/layer2,upperdir=/upper,workdir=/work"
	if options != expected {
		t.Errorf("BuildOverlayOptions() = %v, want %v", options, expected)
	}
}

func TestOverlayFromLayers(t *testing.T) {
	layers := []string{"/base", "/layer1", "/layer2"}
	containerDir := "/var/lib/membrane/containers/test123"

	config := OverlayFromLayers(layers, containerDir)

	if len(config.LowerDirs) != 3 {
		t.Errorf("LowerDirs count = %d, want 3", len(config.LowerDirs))
	}
	if config.UpperDir != "/var/lib/membrane/containers/test123/diff" {
		t.Errorf("UpperDir = %s, want /var/lib/membrane/containers/test123/diff", config.UpperDir)
	}
	if config.WorkDir != "/var/lib/membrane/containers/test123/work" {
		t.Errorf("WorkDir = %s, want /var/lib/membrane/containers/test123/work", config.WorkDir)
	}
	if config.MergedDir != "/var/lib/membrane/containers/test123/merged" {
		t.Errorf("MergedDir = %s, want /var/lib/membrane/containers/test123/merged", config.MergedDir)
	}
}

func TestPrepareOverlayDirs(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "overlay-prepare-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	config := &OverlayConfig{
		LowerDirs: []string{tmpDir}, // Just needs to exist for validation
		UpperDir:  filepath.Join(tmpDir, "upper"),
		WorkDir:   filepath.Join(tmpDir, "work"),
		MergedDir: filepath.Join(tmpDir, "merged"),
	}

	if err := PrepareOverlayDirs(config); err != nil {
		t.Fatalf("PrepareOverlayDirs() error = %v", err)
	}

	// Check directories were created
	for _, dir := range []string{config.UpperDir, config.WorkDir, config.MergedDir} {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("Directory %s was not created", dir)
		}
	}
}
