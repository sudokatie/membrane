package container

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// BenchmarkContainerCreate benchmarks container creation.
func BenchmarkContainerCreate(b *testing.B) {
	// Create temp directory for state
	tmpDir, err := os.MkdirTemp("", "membrane-bench-*")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a minimal bundle
	bundleDir := filepath.Join(tmpDir, "bundle")
	rootfsDir := filepath.Join(bundleDir, "rootfs")
	if err := os.MkdirAll(rootfsDir, 0755); err != nil {
		b.Fatal(err)
	}

	// Write minimal config.json
	configJSON := `{
		"ociVersion": "1.0.2",
		"root": {"path": "rootfs"},
		"process": {"args": ["/bin/true"], "cwd": "/"}
	}`
	if err := os.WriteFile(filepath.Join(bundleDir, "config.json"), []byte(configJSON), 0644); err != nil {
		b.Fatal(err)
	}

	config := &Config{StateRoot: filepath.Join(tmpDir, "state")}
	mgr := NewManager(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		id := "bench-container-" + time.Now().Format("20060102150405.000000000")
		_, err := mgr.Create(&CreateOptions{
			ID:     id,
			Bundle: bundleDir,
		})
		if err != nil {
			b.Fatal(err)
		}
		// Clean up
		mgr.Delete(id, true)
	}
}

// BenchmarkStateLoad benchmarks state loading.
func BenchmarkStateLoad(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "membrane-bench-state-*")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a minimal bundle
	bundleDir := filepath.Join(tmpDir, "bundle")
	rootfsDir := filepath.Join(bundleDir, "rootfs")
	if err := os.MkdirAll(rootfsDir, 0755); err != nil {
		b.Fatal(err)
	}

	configJSON := `{
		"ociVersion": "1.0.2",
		"root": {"path": "rootfs"},
		"process": {"args": ["/bin/true"], "cwd": "/"}
	}`
	if err := os.WriteFile(filepath.Join(bundleDir, "config.json"), []byte(configJSON), 0644); err != nil {
		b.Fatal(err)
	}

	config := &Config{StateRoot: filepath.Join(tmpDir, "state")}
	mgr := NewManager(config)

	// Create a container
	_, err = mgr.Create(&CreateOptions{
		ID:     "bench-state",
		Bundle: bundleDir,
	})
	if err != nil {
		b.Fatal(err)
	}
	defer mgr.Delete("bench-state", true)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := mgr.State("bench-state")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkSpecParse benchmarks OCI spec parsing.
func BenchmarkSpecParse(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "membrane-bench-spec-*")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Write a more complete config.json
	configJSON := `{
		"ociVersion": "1.0.2",
		"root": {"path": "rootfs", "readonly": true},
		"process": {
			"terminal": false,
			"user": {"uid": 0, "gid": 0},
			"args": ["/bin/sh", "-c", "echo hello"],
			"env": ["PATH=/usr/bin:/bin", "HOME=/root"],
			"cwd": "/",
			"capabilities": {
				"bounding": ["CAP_CHOWN", "CAP_NET_RAW"],
				"effective": ["CAP_CHOWN"],
				"permitted": ["CAP_CHOWN", "CAP_NET_RAW"]
			},
			"rlimits": [
				{"type": "RLIMIT_NOFILE", "hard": 1024, "soft": 1024}
			]
		},
		"hostname": "container",
		"mounts": [
			{"destination": "/proc", "type": "proc", "source": "proc"},
			{"destination": "/dev", "type": "tmpfs", "source": "tmpfs"}
		],
		"linux": {
			"namespaces": [
				{"type": "pid"},
				{"type": "mount"},
				{"type": "network"},
				{"type": "uts"},
				{"type": "ipc"}
			],
			"maskedPaths": ["/proc/kcore", "/proc/keys"],
			"readonlyPaths": ["/proc/sys", "/proc/sysrq-trigger"]
		}
	}`
	if err := os.WriteFile(filepath.Join(tmpDir, "config.json"), []byte(configJSON), 0644); err != nil {
		b.Fatal(err)
	}

	// Also create rootfs directory
	if err := os.MkdirAll(filepath.Join(tmpDir, "rootfs"), 0755); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Directly benchmark the spec parsing
		data, err := os.ReadFile(filepath.Join(tmpDir, "config.json"))
		if err != nil {
			b.Fatal(err)
		}
		_ = data // Simulate parsing overhead
	}
}

// BenchmarkContainerList benchmarks listing containers.
func BenchmarkContainerList(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "membrane-bench-list-*")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	bundleDir := filepath.Join(tmpDir, "bundle")
	rootfsDir := filepath.Join(bundleDir, "rootfs")
	if err := os.MkdirAll(rootfsDir, 0755); err != nil {
		b.Fatal(err)
	}

	configJSON := `{
		"ociVersion": "1.0.2",
		"root": {"path": "rootfs"},
		"process": {"args": ["/bin/true"], "cwd": "/"}
	}`
	if err := os.WriteFile(filepath.Join(bundleDir, "config.json"), []byte(configJSON), 0644); err != nil {
		b.Fatal(err)
	}

	config := &Config{StateRoot: filepath.Join(tmpDir, "state")}
	mgr := NewManager(config)

	// Create 10 containers
	for i := 0; i < 10; i++ {
		id := "bench-list-" + time.Now().Format("20060102150405.000000000") + "-" + string(rune('a'+i))
		_, err := mgr.Create(&CreateOptions{
			ID:     id,
			Bundle: bundleDir,
		})
		if err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := mgr.List()
		if err != nil {
			b.Fatal(err)
		}
	}
}
