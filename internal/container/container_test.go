package container

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/sudokatie/membrane/internal/state"
)

func TestManager_Create(t *testing.T) {
	stateRoot := t.TempDir()
	bundle := createTestBundle(t)

	manager := NewManager(&Config{StateRoot: stateRoot})

	container, err := manager.Create(&CreateOptions{
		ID:     "test-container",
		Bundle: bundle,
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if container.ID != "test-container" {
		t.Errorf("ID = %s, want test-container", container.ID)
	}
	if container.State.Status != state.StatusCreated {
		t.Errorf("Status = %s, want created", container.State.Status)
	}
	if container.Spec == nil {
		t.Error("Spec is nil")
	}
}

func TestManager_CreateInvalidID(t *testing.T) {
	stateRoot := t.TempDir()
	manager := NewManager(&Config{StateRoot: stateRoot})

	_, err := manager.Create(&CreateOptions{
		ID:     "",
		Bundle: "/some/path",
	})
	if err == nil {
		t.Fatal("expected error for empty ID")
	}
}

func TestManager_CreateBundleNotFound(t *testing.T) {
	stateRoot := t.TempDir()
	manager := NewManager(&Config{StateRoot: stateRoot})

	_, err := manager.Create(&CreateOptions{
		ID:     "test-container",
		Bundle: "/nonexistent/bundle",
	})
	if err == nil {
		t.Fatal("expected error for missing bundle")
	}
}

func TestManager_CreateDuplicate(t *testing.T) {
	stateRoot := t.TempDir()
	bundle := createTestBundle(t)
	manager := NewManager(&Config{StateRoot: stateRoot})

	// Create first
	_, err := manager.Create(&CreateOptions{
		ID:     "test-container",
		Bundle: bundle,
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Create again should fail
	_, err = manager.Create(&CreateOptions{
		ID:     "test-container",
		Bundle: bundle,
	})
	if err == nil {
		t.Fatal("expected error for duplicate container")
	}
}

func TestManager_Get(t *testing.T) {
	stateRoot := t.TempDir()
	bundle := createTestBundle(t)
	manager := NewManager(&Config{StateRoot: stateRoot})

	// Create container
	_, err := manager.Create(&CreateOptions{
		ID:     "test-container",
		Bundle: bundle,
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Get container
	container, err := manager.Get("test-container")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if container.ID != "test-container" {
		t.Errorf("ID = %s, want test-container", container.ID)
	}
}

func TestManager_GetNotFound(t *testing.T) {
	stateRoot := t.TempDir()
	manager := NewManager(&Config{StateRoot: stateRoot})

	_, err := manager.Get("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent container")
	}
}

func TestManager_List(t *testing.T) {
	stateRoot := t.TempDir()
	bundle := createTestBundle(t)
	manager := NewManager(&Config{StateRoot: stateRoot})

	// Create multiple containers
	for i := 0; i < 3; i++ {
		_, err := manager.Create(&CreateOptions{
			ID:     fmt.Sprintf("container-%d", i),
			Bundle: bundle,
		})
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	}

	// List
	containers, err := manager.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(containers) != 3 {
		t.Errorf("List returned %d containers, want 3", len(containers))
	}
}

func TestManager_Delete(t *testing.T) {
	stateRoot := t.TempDir()
	bundle := createTestBundle(t)
	manager := NewManager(&Config{StateRoot: stateRoot})

	// Create container
	_, err := manager.Create(&CreateOptions{
		ID:     "test-container",
		Bundle: bundle,
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Delete
	if err := manager.Delete("test-container", false); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deleted
	_, err = manager.Get("test-container")
	if err == nil {
		t.Error("container still exists after delete")
	}
}

func TestManager_DeleteNotFound(t *testing.T) {
	stateRoot := t.TempDir()
	manager := NewManager(&Config{StateRoot: stateRoot})

	err := manager.Delete("nonexistent", false)
	if err == nil {
		t.Fatal("expected error for nonexistent container")
	}
}

func TestManager_CreateWithAnnotations(t *testing.T) {
	stateRoot := t.TempDir()
	bundle := createTestBundle(t)
	manager := NewManager(&Config{StateRoot: stateRoot})

	container, err := manager.Create(&CreateOptions{
		ID:     "test-container",
		Bundle: bundle,
		Annotations: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if container.State.Annotations["key1"] != "value1" {
		t.Error("annotation key1 not set")
	}
	if container.State.Annotations["key2"] != "value2" {
		t.Error("annotation key2 not set")
	}
}

func TestManager_CreateMissingRootfs(t *testing.T) {
	stateRoot := t.TempDir()
	bundle := t.TempDir()

	// Create config.json but no rootfs
	config := `{
		"ociVersion": "1.0.2",
		"root": {"path": "rootfs"},
		"process": {"args": ["/bin/sh"], "cwd": "/"}
	}`
	if err := os.WriteFile(filepath.Join(bundle, "config.json"), []byte(config), 0644); err != nil {
		t.Fatal(err)
	}

	manager := NewManager(&Config{StateRoot: stateRoot})

	_, err := manager.Create(&CreateOptions{
		ID:     "test-container",
		Bundle: bundle,
	})
	if err == nil {
		t.Fatal("expected error for missing rootfs")
	}
}

// createTestBundle creates a minimal test bundle.
func createTestBundle(t *testing.T) string {
	t.Helper()
	bundle := t.TempDir()

	// Create config.json
	config := `{
		"ociVersion": "1.0.2",
		"root": {"path": "rootfs"},
		"process": {"args": ["/bin/sh"], "cwd": "/"}
	}`
	if err := os.WriteFile(filepath.Join(bundle, "config.json"), []byte(config), 0644); err != nil {
		t.Fatal(err)
	}

	// Create rootfs directory
	if err := os.MkdirAll(filepath.Join(bundle, "rootfs"), 0755); err != nil {
		t.Fatal(err)
	}

	return bundle
}
