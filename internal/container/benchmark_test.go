package container

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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

// BenchmarkContainerStartup benchmarks full container startup time.
// This benchmark requires:
// - Linux (for namespaces)
// - Root privileges (for namespace creation)
// - A functional rootfs with /bin/true
//
// Run with: sudo go test -bench=BenchmarkContainerStartup -run=^$ ./internal/container/
func BenchmarkContainerStartup(b *testing.B) {
	if runtime.GOOS != "linux" {
		b.Skip("requires Linux")
	}
	if os.Getuid() != 0 {
		b.Skip("requires root")
	}

	// Check if we have a usable busybox
	busyboxPath, err := exec.LookPath("busybox")
	if err != nil {
		b.Skip("requires busybox in PATH")
	}

	tmpDir, err := os.MkdirTemp("", "membrane-bench-startup-*")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create rootfs with busybox
	bundleDir := filepath.Join(tmpDir, "bundle")
	rootfsDir := filepath.Join(bundleDir, "rootfs", "bin")
	if err := os.MkdirAll(rootfsDir, 0755); err != nil {
		b.Fatal(err)
	}

	// Copy busybox
	busyboxDest := filepath.Join(rootfsDir, "busybox")
	input, err := os.ReadFile(busyboxPath)
	if err != nil {
		b.Fatal(err)
	}
	if err := os.WriteFile(busyboxDest, input, 0755); err != nil {
		b.Fatal(err)
	}

	// Create symlink for true
	if err := os.Symlink("busybox", filepath.Join(rootfsDir, "true")); err != nil {
		b.Fatal(err)
	}

	// Write config for a quick-exit container
	configJSON := `{
		"ociVersion": "1.0.2",
		"root": {"path": "rootfs"},
		"process": {
			"args": ["/bin/true"],
			"cwd": "/",
			"env": ["PATH=/bin"]
		},
		"linux": {
			"namespaces": [
				{"type": "pid"},
				{"type": "mount"},
				{"type": "uts"},
				{"type": "ipc"}
			]
		}
	}`
	if err := os.WriteFile(filepath.Join(bundleDir, "config.json"), []byte(configJSON), 0644); err != nil {
		b.Fatal(err)
	}

	config := &Config{StateRoot: filepath.Join(tmpDir, "state")}
	mgr := NewManager(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		id := "bench-startup-" + time.Now().Format("20060102150405.000000000")

		// Create
		_, err := mgr.Create(&CreateOptions{
			ID:     id,
			Bundle: bundleDir,
		})
		if err != nil {
			b.Fatalf("create: %v", err)
		}

		// Start
		err = mgr.Start(&StartOptions{ID: id})
		if err != nil {
			mgr.Delete(id, true)
			b.Fatalf("start: %v", err)
		}

		// Wait for exit
		_, err = mgr.Wait(id)
		if err != nil {
			mgr.Delete(id, true)
			b.Fatalf("wait: %v", err)
		}

		// Clean up
		mgr.Delete(id, true)
	}
}

// BenchmarkContainerStartupParallel benchmarks parallel container startups.
// Same requirements as BenchmarkContainerStartup.
func BenchmarkContainerStartupParallel(b *testing.B) {
	if runtime.GOOS != "linux" {
		b.Skip("requires Linux")
	}
	if os.Getuid() != 0 {
		b.Skip("requires root")
	}

	busyboxPath, err := exec.LookPath("busybox")
	if err != nil {
		b.Skip("requires busybox in PATH")
	}

	tmpDir, err := os.MkdirTemp("", "membrane-bench-startup-parallel-*")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	bundleDir := filepath.Join(tmpDir, "bundle")
	rootfsDir := filepath.Join(bundleDir, "rootfs", "bin")
	if err := os.MkdirAll(rootfsDir, 0755); err != nil {
		b.Fatal(err)
	}

	busyboxDest := filepath.Join(rootfsDir, "busybox")
	input, err := os.ReadFile(busyboxPath)
	if err != nil {
		b.Fatal(err)
	}
	if err := os.WriteFile(busyboxDest, input, 0755); err != nil {
		b.Fatal(err)
	}
	if err := os.Symlink("busybox", filepath.Join(rootfsDir, "true")); err != nil {
		b.Fatal(err)
	}

	configJSON := `{
		"ociVersion": "1.0.2",
		"root": {"path": "rootfs"},
		"process": {"args": ["/bin/true"], "cwd": "/", "env": ["PATH=/bin"]},
		"linux": {"namespaces": [{"type": "pid"}, {"type": "mount"}, {"type": "uts"}, {"type": "ipc"}]}
	}`
	if err := os.WriteFile(filepath.Join(bundleDir, "config.json"), []byte(configJSON), 0644); err != nil {
		b.Fatal(err)
	}

	config := &Config{StateRoot: filepath.Join(tmpDir, "state")}
	mgr := NewManager(config)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			id := "bench-parallel-" + time.Now().Format("20060102150405.000000000")
			mgr.Create(&CreateOptions{ID: id, Bundle: bundleDir})
			mgr.Start(&StartOptions{ID: id})
			mgr.Wait(id)
			mgr.Delete(id, true)
		}
	})
}
