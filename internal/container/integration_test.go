package container

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sudokatie/membrane/internal/state"
)

// TestFullLifecycle tests the complete container lifecycle:
// create -> state -> delete
// Note: start requires Linux namespaces
func TestFullLifecycle(t *testing.T) {
	stateRoot := t.TempDir()
	bundle := createTestBundle(t)

	mgr := NewManager(&Config{StateRoot: stateRoot})

	// Create
	container, err := mgr.Create(&CreateOptions{
		ID:     "test-lifecycle",
		Bundle: bundle,
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if container.State.Status != state.StatusCreated {
		t.Errorf("expected status created, got %s", container.State.Status)
	}

	// State
	st, err := mgr.State("test-lifecycle")
	if err != nil {
		t.Fatalf("State failed: %v", err)
	}
	if st.ID != "test-lifecycle" {
		t.Errorf("expected id test-lifecycle, got %s", st.ID)
	}
	if st.Bundle == "" {
		t.Error("expected bundle path in state")
	}

	// Get
	c, err := mgr.Get("test-lifecycle")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if c.ID != "test-lifecycle" {
		t.Errorf("expected id test-lifecycle, got %s", c.ID)
	}

	// List
	containers, err := mgr.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(containers) != 1 {
		t.Errorf("expected 1 container, got %d", len(containers))
	}

	// Delete
	if err := mgr.Delete("test-lifecycle", false); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deleted
	_, err = mgr.Get("test-lifecycle")
	if err == nil {
		t.Error("expected error after delete")
	}
}

// TestMultipleContainers tests managing multiple containers
func TestMultipleContainers(t *testing.T) {
	stateRoot := t.TempDir()
	bundle := createTestBundle(t)

	mgr := NewManager(&Config{StateRoot: stateRoot})

	// Create multiple containers
	ids := []string{"container-1", "container-2", "container-3"}
	for _, id := range ids {
		_, err := mgr.Create(&CreateOptions{
			ID:     id,
			Bundle: bundle,
		})
		if err != nil {
			t.Fatalf("Create %s failed: %v", id, err)
		}
	}

	// List all
	containers, err := mgr.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(containers) != 3 {
		t.Errorf("expected 3 containers, got %d", len(containers))
	}

	// Delete one
	if err := mgr.Delete("container-2", false); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// List again
	containers, err = mgr.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(containers) != 2 {
		t.Errorf("expected 2 containers, got %d", len(containers))
	}

	// Cleanup
	for _, c := range containers {
		mgr.Delete(c.ID, false)
	}
}

// TestContainerAnnotations tests annotation handling
func TestContainerAnnotations(t *testing.T) {
	stateRoot := t.TempDir()
	bundle := createTestBundle(t)

	mgr := NewManager(&Config{StateRoot: stateRoot})

	container, err := mgr.Create(&CreateOptions{
		ID:     "test-annotations",
		Bundle: bundle,
		Annotations: map[string]string{
			"com.example.key1": "value1",
			"com.example.key2": "value2",
		},
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if container.State.Annotations["com.example.key1"] != "value1" {
		t.Error("annotation key1 not set")
	}
	if container.State.Annotations["com.example.key2"] != "value2" {
		t.Error("annotation key2 not set")
	}

	// Verify annotations persist
	st, err := mgr.State("test-annotations")
	if err != nil {
		t.Fatalf("State failed: %v", err)
	}
	if st.Annotations["com.example.key1"] != "value1" {
		t.Error("annotation not persisted")
	}

	mgr.Delete("test-annotations", false)
}

// TestContainerStateTransitions tests state transition validation
func TestContainerStateTransitions(t *testing.T) {
	stateRoot := t.TempDir()
	bundle := createTestBundle(t)

	mgr := NewManager(&Config{StateRoot: stateRoot})

	// Create
	_, err := mgr.Create(&CreateOptions{
		ID:     "test-transitions",
		Bundle: bundle,
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Can't create again
	_, err = mgr.Create(&CreateOptions{
		ID:     "test-transitions",
		Bundle: bundle,
	})
	if err == nil {
		t.Error("expected error creating duplicate container")
	}

	// Can delete in created state
	if err := mgr.Delete("test-transitions", false); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

// TestBundleValidation tests bundle path validation
func TestBundleValidation(t *testing.T) {
	stateRoot := t.TempDir()
	mgr := NewManager(&Config{StateRoot: stateRoot})

	// Missing bundle
	_, err := mgr.Create(&CreateOptions{
		ID:     "test-missing-bundle",
		Bundle: "/nonexistent/bundle",
	})
	if err == nil {
		t.Error("expected error for missing bundle")
	}

	// Bundle is a file, not directory
	tmpFile := filepath.Join(t.TempDir(), "not-a-dir")
	os.WriteFile(tmpFile, []byte("test"), 0644)
	_, err = mgr.Create(&CreateOptions{
		ID:     "test-file-bundle",
		Bundle: tmpFile,
	})
	if err == nil {
		t.Error("expected error for file bundle")
	}

	// Missing config.json
	emptyBundle := t.TempDir()
	_, err = mgr.Create(&CreateOptions{
		ID:     "test-no-config",
		Bundle: emptyBundle,
	})
	if err == nil {
		t.Error("expected error for missing config.json")
	}

	// Invalid config.json
	badBundle := t.TempDir()
	os.WriteFile(filepath.Join(badBundle, "config.json"), []byte("not json"), 0644)
	_, err = mgr.Create(&CreateOptions{
		ID:     "test-bad-config",
		Bundle: badBundle,
	})
	if err == nil {
		t.Error("expected error for invalid config.json")
	}
}

// TestSpecValidation tests OCI spec validation
func TestSpecValidation(t *testing.T) {
	stateRoot := t.TempDir()
	mgr := NewManager(&Config{StateRoot: stateRoot})

	tests := []struct {
		name   string
		config string
		errMsg string
	}{
		{
			name:   "missing ociVersion",
			config: `{"root": {"path": "rootfs"}, "process": {"args": ["/bin/sh"], "cwd": "/"}}`,
			errMsg: "ociVersion",
		},
		{
			name:   "missing root",
			config: `{"ociVersion": "1.0.2", "process": {"args": ["/bin/sh"], "cwd": "/"}}`,
			errMsg: "root",
		},
		{
			name:   "missing process",
			config: `{"ociVersion": "1.0.2", "root": {"path": "rootfs"}}`,
			errMsg: "process",
		},
		{
			name:   "missing args",
			config: `{"ociVersion": "1.0.2", "root": {"path": "rootfs"}, "process": {"cwd": "/"}}`,
			errMsg: "args",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bundle := t.TempDir()
			os.WriteFile(filepath.Join(bundle, "config.json"), []byte(tt.config), 0644)
			os.MkdirAll(filepath.Join(bundle, "rootfs"), 0755)

			_, err := mgr.Create(&CreateOptions{
				ID:     "test-" + tt.name,
				Bundle: bundle,
			})
			if err == nil {
				t.Errorf("expected error containing %q", tt.errMsg)
			}
		})
	}
}

// TestForceDelete tests force deletion
func TestForceDelete(t *testing.T) {
	stateRoot := t.TempDir()
	bundle := createTestBundle(t)

	mgr := NewManager(&Config{StateRoot: stateRoot})

	// Create container
	_, err := mgr.Create(&CreateOptions{
		ID:     "test-force",
		Bundle: bundle,
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Force delete should work in any state
	if err := mgr.Delete("test-force", true); err != nil {
		t.Fatalf("Force delete failed: %v", err)
	}
}
