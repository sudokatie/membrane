package oci

import (
	"encoding/json"
	"testing"
)

func TestSpecUnmarshal(t *testing.T) {
	specJSON := `{
		"ociVersion": "1.0.2",
		"root": {
			"path": "rootfs",
			"readonly": false
		},
		"process": {
			"terminal": false,
			"user": {"uid": 0, "gid": 0},
			"args": ["/bin/sh"],
			"env": ["PATH=/bin"],
			"cwd": "/"
		},
		"hostname": "test"
	}`

	var spec Spec
	if err := json.Unmarshal([]byte(specJSON), &spec); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if spec.Version != "1.0.2" {
		t.Errorf("version = %q, want %q", spec.Version, "1.0.2")
	}
	if spec.Root.Path != "rootfs" {
		t.Errorf("root.path = %q, want %q", spec.Root.Path, "rootfs")
	}
	if spec.Hostname != "test" {
		t.Errorf("hostname = %q, want %q", spec.Hostname, "test")
	}
}

func TestHooksUnmarshal(t *testing.T) {
	specJSON := `{
		"ociVersion": "1.0.2",
		"root": {"path": "rootfs"},
		"process": {"args": ["/bin/sh"], "cwd": "/"},
		"hooks": {
			"prestart": [
				{"path": "/usr/bin/prestart", "args": ["--flag"], "timeout": 10}
			],
			"createRuntime": [
				{"path": "/usr/bin/create-runtime"}
			],
			"createContainer": [
				{"path": "/usr/bin/create-container", "env": ["FOO=bar"]}
			],
			"startContainer": [
				{"path": "/usr/bin/start-container"}
			],
			"poststart": [
				{"path": "/usr/bin/poststart"}
			],
			"poststop": [
				{"path": "/usr/bin/poststop"}
			]
		}
	}`

	var spec Spec
	if err := json.Unmarshal([]byte(specJSON), &spec); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if spec.Hooks == nil {
		t.Fatal("hooks is nil")
	}

	if len(spec.Hooks.Prestart) != 1 {
		t.Errorf("prestart hooks = %d, want 1", len(spec.Hooks.Prestart))
	}
	if spec.Hooks.Prestart[0].Path != "/usr/bin/prestart" {
		t.Errorf("prestart path = %q, want %q", spec.Hooks.Prestart[0].Path, "/usr/bin/prestart")
	}
	if spec.Hooks.Prestart[0].Timeout == nil || *spec.Hooks.Prestart[0].Timeout != 10 {
		t.Error("prestart timeout should be 10")
	}

	if len(spec.Hooks.CreateRuntime) != 1 {
		t.Errorf("createRuntime hooks = %d, want 1", len(spec.Hooks.CreateRuntime))
	}

	if len(spec.Hooks.CreateContainer) != 1 {
		t.Errorf("createContainer hooks = %d, want 1", len(spec.Hooks.CreateContainer))
	}
	if len(spec.Hooks.CreateContainer[0].Env) != 1 {
		t.Errorf("createContainer env = %d, want 1", len(spec.Hooks.CreateContainer[0].Env))
	}

	if len(spec.Hooks.StartContainer) != 1 {
		t.Errorf("startContainer hooks = %d, want 1", len(spec.Hooks.StartContainer))
	}

	if len(spec.Hooks.Poststart) != 1 {
		t.Errorf("poststart hooks = %d, want 1", len(spec.Hooks.Poststart))
	}

	if len(spec.Hooks.Poststop) != 1 {
		t.Errorf("poststop hooks = %d, want 1", len(spec.Hooks.Poststop))
	}
}

func TestLinuxNamespaces(t *testing.T) {
	specJSON := `{
		"ociVersion": "1.0.2",
		"root": {"path": "rootfs"},
		"process": {"args": ["/bin/sh"], "cwd": "/"},
		"linux": {
			"namespaces": [
				{"type": "pid"},
				{"type": "network"},
				{"type": "mount"},
				{"type": "ipc"},
				{"type": "uts"},
				{"type": "user"},
				{"type": "cgroup"}
			]
		}
	}`

	var spec Spec
	if err := json.Unmarshal([]byte(specJSON), &spec); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if spec.Linux == nil {
		t.Fatal("linux is nil")
	}

	if len(spec.Linux.Namespaces) != 7 {
		t.Errorf("namespaces = %d, want 7", len(spec.Linux.Namespaces))
	}

	// Check all namespace types
	types := make(map[NamespaceType]bool)
	for _, ns := range spec.Linux.Namespaces {
		types[ns.Type] = true
	}

	expected := []NamespaceType{
		PIDNamespace, NetworkNamespace, MountNamespace,
		IPCNamespace, UTSNamespace, UserNamespace, CgroupNamespace,
	}
	for _, ns := range expected {
		if !types[ns] {
			t.Errorf("namespace %s not found", ns)
		}
	}
}

func TestLinuxUIDGIDMappings(t *testing.T) {
	specJSON := `{
		"ociVersion": "1.0.2",
		"root": {"path": "rootfs"},
		"process": {"args": ["/bin/sh"], "cwd": "/"},
		"linux": {
			"uidMappings": [
				{"containerID": 0, "hostID": 1000, "size": 1}
			],
			"gidMappings": [
				{"containerID": 0, "hostID": 1000, "size": 1}
			]
		}
	}`

	var spec Spec
	if err := json.Unmarshal([]byte(specJSON), &spec); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(spec.Linux.UIDMappings) != 1 {
		t.Fatalf("uidMappings = %d, want 1", len(spec.Linux.UIDMappings))
	}
	if spec.Linux.UIDMappings[0].ContainerID != 0 {
		t.Errorf("uid containerID = %d, want 0", spec.Linux.UIDMappings[0].ContainerID)
	}
	if spec.Linux.UIDMappings[0].HostID != 1000 {
		t.Errorf("uid hostID = %d, want 1000", spec.Linux.UIDMappings[0].HostID)
	}

	if len(spec.Linux.GIDMappings) != 1 {
		t.Fatalf("gidMappings = %d, want 1", len(spec.Linux.GIDMappings))
	}
}

func TestSeccompConfig(t *testing.T) {
	specJSON := `{
		"ociVersion": "1.0.2",
		"root": {"path": "rootfs"},
		"process": {"args": ["/bin/sh"], "cwd": "/"},
		"linux": {
			"seccomp": {
				"defaultAction": "SCMP_ACT_ALLOW",
				"architectures": ["SCMP_ARCH_X86_64"],
				"syscalls": [
					{
						"names": ["reboot", "kexec_load"],
						"action": "SCMP_ACT_ERRNO",
						"errnoRet": 1
					}
				]
			}
		}
	}`

	var spec Spec
	if err := json.Unmarshal([]byte(specJSON), &spec); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if spec.Linux.Seccomp == nil {
		t.Fatal("seccomp is nil")
	}

	if spec.Linux.Seccomp.DefaultAction != ActAllow {
		t.Errorf("defaultAction = %q, want %q", spec.Linux.Seccomp.DefaultAction, ActAllow)
	}

	if len(spec.Linux.Seccomp.Syscalls) != 1 {
		t.Fatalf("syscalls = %d, want 1", len(spec.Linux.Seccomp.Syscalls))
	}

	sc := spec.Linux.Seccomp.Syscalls[0]
	if len(sc.Names) != 2 {
		t.Errorf("syscall names = %d, want 2", len(sc.Names))
	}
	if sc.Action != ActErrno {
		t.Errorf("action = %q, want %q", sc.Action, ActErrno)
	}
	if sc.ErrnoRet == nil || *sc.ErrnoRet != 1 {
		t.Error("errnoRet should be 1")
	}
}

func TestResourceLimits(t *testing.T) {
	specJSON := `{
		"ociVersion": "1.0.2",
		"root": {"path": "rootfs"},
		"process": {"args": ["/bin/sh"], "cwd": "/"},
		"linux": {
			"resources": {
				"memory": {"limit": 536870912},
				"cpu": {"quota": 50000, "period": 100000},
				"pids": {"limit": 100}
			}
		}
	}`

	var spec Spec
	if err := json.Unmarshal([]byte(specJSON), &spec); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if spec.Linux.Resources == nil {
		t.Fatal("resources is nil")
	}

	if spec.Linux.Resources.Memory == nil || *spec.Linux.Resources.Memory.Limit != 536870912 {
		t.Error("memory limit should be 536870912")
	}

	if spec.Linux.Resources.CPU == nil {
		t.Fatal("cpu is nil")
	}
	if *spec.Linux.Resources.CPU.Quota != 50000 {
		t.Errorf("cpu quota = %d, want 50000", *spec.Linux.Resources.CPU.Quota)
	}

	if spec.Linux.Resources.Pids == nil || spec.Linux.Resources.Pids.Limit != 100 {
		t.Error("pids limit should be 100")
	}
}
