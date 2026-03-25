# Membrane

A minimal container runtime. Because sometimes you want to understand what Docker is actually doing between your `docker run` and your process.

## What This Is

Membrane implements the OCI Runtime Specification using:
- Linux namespaces for isolation (PID, network, mount, UTS, IPC, user, cgroup)
- Cgroups v2 for resource limits
- Seccomp for syscall filtering
- Overlayfs for layered filesystems
- OCI lifecycle hooks

It's not a replacement for containerd or runc in production. It's for learning, experimenting, and understanding containers at the syscall level.

## Features

- Full OCI runtime spec compliance
- All namespace types including user and cgroup namespaces
- UID/GID mapping for user namespaces
- Cgroups v2 (unified hierarchy)
- CPU, memory, PIDs, and IO limits
- Seccomp syscall filtering
- Capability management
- OCI lifecycle hooks (prestart, createRuntime, createContainer, startContainer, poststart, poststop)
- AppArmor and SELinux integration
- PTY/terminal support
- Structured logging
- No external dependencies at runtime

## Quick Start

```bash
# Build
go build ./cmd/membrane

# Create an OCI bundle (you'll need a rootfs)
mkdir -p mycontainer/rootfs
# Copy a minimal rootfs (busybox, alpine, etc.)

# Create config.json (OCI spec format)
cat > mycontainer/config.json << 'EOF'
{
  "ociVersion": "1.0.2",
  "process": {
    "args": ["/bin/sh"],
    "cwd": "/"
  },
  "root": {
    "path": "rootfs",
    "readonly": false
  }
}
EOF

# Run it
sudo ./membrane run mycontainer ./mycontainer
```

## Commands

```
membrane create <id> <bundle>  Create a container
membrane start <id>            Start a created container
membrane run <id> <bundle>     Create and start (convenience)
membrane state <id>            Query container state (JSON)
membrane kill <id> [signal]    Send signal to container
membrane delete <id>           Clean up container
membrane list                  List all containers
membrane exec <id> <cmd>       Execute command in running container
membrane wait <id>             Wait for container to exit
membrane spec                  Generate default config.json
membrane version               Print version info
```

### Global Flags

```
--root <path>       State directory (default: /run/membrane)
--log-level <lvl>   Log level: error, warn, info, debug (default: info)
```

### Exec Flags

```
--cwd <path>    Working directory (default: /)
--env <var>     Environment variable (can be repeated)
--user <uid>    User ID
--group <gid>   Group ID
```

### Delete Flags

```
-f, --force     Force delete even if running
```

## Exit Codes

Per OCI runtime spec:
- `0`: Success
- `1`: General error
- `125`: Container failed to run
- `127`: Command not found in container

## OCI Lifecycle Hooks

Membrane supports all OCI lifecycle hooks:

```json
{
  "hooks": {
    "prestart": [{"path": "/usr/bin/hook", "timeout": 10}],
    "createRuntime": [{"path": "/usr/bin/hook"}],
    "createContainer": [{"path": "/usr/bin/hook"}],
    "startContainer": [{"path": "/usr/bin/hook"}],
    "poststart": [{"path": "/usr/bin/hook"}],
    "poststop": [{"path": "/usr/bin/hook"}]
  }
}
```

Hooks receive container state via stdin (JSON) and can set environment variables and arguments.

## User Namespace Support

Enable user namespaces with UID/GID mappings:

```json
{
  "linux": {
    "namespaces": [{"type": "user"}],
    "uidMappings": [{"containerID": 0, "hostID": 1000, "size": 1}],
    "gidMappings": [{"containerID": 0, "hostID": 1000, "size": 1}]
  }
}
```

## Requirements

- Linux kernel 5.8+ (cgroups v2 unified hierarchy)
- Go 1.21+
- Root privileges (for namespace operations)

## What It Does

When you run `membrane run`:

1. Parses OCI config.json from the bundle
2. Runs createRuntime hooks
3. Creates a cgroup for resource limits
4. Forks with new namespaces (PID, mount, network, user, cgroup, etc.)
5. Writes UID/GID mappings for user namespace
6. Sets up hostname
7. Applies sysctl settings
8. Sets up root filesystem with pivot_root
9. Runs createContainer hooks
10. Masks sensitive paths, applies readonly paths
11. Creates device nodes
12. Applies rlimits
13. Sets no_new_privs flag
14. Applies AppArmor/SELinux profiles
15. Drops capabilities
16. Runs startContainer and prestart hooks
17. Applies seccomp filters
18. Executes your process
19. Runs poststart hooks
20. On delete, runs poststop hooks

All the magic happens in about 3000 lines of Go.

## What It Doesn't Do

- Pull images (use `skopeo` or similar)
- Network setup (use CNI plugins)
- Run on Windows/macOS
- Checkpoint/restore
- Rootless mode (requires root for v0.1.0)

## Architecture

```
membrane/
├── cmd/membrane/       CLI entry point
├── internal/
│   ├── capabilities/   Linux capability management
│   ├── cgroup/         Cgroups v2 resource limits
│   ├── container/      Container lifecycle (create, start, kill, exec)
│   ├── filesystem/     Mount operations, pivot_root, overlayfs
│   ├── hooks/          OCI lifecycle hooks
│   ├── log/            Structured logging
│   ├── namespace/      Linux namespace setup
│   ├── seccomp/        Syscall filtering
│   ├── spec/           OCI spec parsing
│   └── state/          Container state persistence
└── pkg/oci/            OCI type definitions
```

## Testing

```bash
# Run all tests
go test ./...

# Verbose output
go test -v ./...

# Run OCI compliance tests (requires root)
sudo ./scripts/oci-compliance-test.sh

# Full runtime-tools validation
sudo FULL_VALIDATION=1 ./scripts/oci-compliance-test.sh
```

## Debug Mode

Enable debug logging for troubleshooting:

```bash
sudo ./membrane --log-level=debug run mycontainer ./bundle
```

## License

MIT

---

*Built by Katie, who wanted to understand what happens between the container and the kernel.*
