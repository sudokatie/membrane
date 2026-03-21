# Membrane

A minimal container runtime. Because sometimes you want to understand what Docker is actually doing between your `docker run` and your process.

## What This Is

Membrane implements the OCI Runtime Specification using:
- Linux namespaces for isolation (PID, network, mount, UTS, IPC)
- Cgroups v2 for resource limits
- Seccomp for syscall filtering
- Overlayfs for layered filesystems

It's not a replacement for containerd or runc in production. It's for learning, experimenting, and understanding containers at the syscall level.

## Features

- Full OCI runtime spec compliance
- Cgroups v2 (unified hierarchy)
- Seccomp syscall filtering
- Namespace isolation
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
membrane spec                  Generate default config.json
membrane version               Print version info
```

### Global Flags

```
--root <path>    State directory (default: /run/membrane)
```

## Requirements

- Linux kernel 5.8+ (cgroups v2 unified hierarchy)
- Go 1.21+
- Root privileges (for namespace operations)

## What It Does

When you run `membrane run`:

1. Parses OCI config.json from the bundle
2. Creates a cgroup for resource limits
3. Forks with new namespaces (PID, mount, network, etc.)
4. Sets up the root filesystem with pivot_root
5. Applies seccomp filters
6. Executes your process

All the magic happens in about 2000 lines of Go.

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
│   ├── container/      Container lifecycle (create, start, kill)
│   ├── namespace/      Linux namespace setup
│   ├── cgroup/         Cgroups v2 resource limits
│   ├── filesystem/     Mount operations, pivot_root
│   ├── seccomp/        Syscall filtering
│   ├── spec/           OCI spec parsing
│   └── state/          Container state persistence
└── pkg/oci/            OCI type definitions
```

## Testing

```bash
go test ./...           # Run all tests
go test -v ./...        # Verbose output
go vet ./...            # Static analysis
```

## License

MIT

---

*Built by Katie, who wanted to understand what happens between the container and the kernel.*
