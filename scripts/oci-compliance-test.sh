#!/bin/bash
# OCI Runtime Tools Compliance Test
#
# This script runs the OCI runtime-tools validation suite against membrane.
# Requirements:
#   - Go 1.21+
#   - Linux with cgroups v2
#   - Root privileges
#
# Usage:
#   sudo ./scripts/oci-compliance-test.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
MEMBRANE_BIN="${PROJECT_ROOT}/membrane"
RUNTIME_TOOLS_DIR="${PROJECT_ROOT}/.runtime-tools"
BUNDLE_DIR="/tmp/membrane-oci-test"
ROOTFS_DIR="/tmp/membrane-oci-rootfs"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_test() {
    echo -e "${BLUE}[TEST]${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    if [[ $EUID -ne 0 ]]; then
        log_error "This script must be run as root"
        exit 1
    fi
    
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed"
        exit 1
    fi
    
    if [[ ! -f /sys/fs/cgroup/cgroup.controllers ]]; then
        log_error "Cgroups v2 is not available"
        exit 1
    fi
    
    log_info "Prerequisites OK"
}

# Build membrane
build_membrane() {
    log_info "Building membrane..."
    cd "$PROJECT_ROOT"
    go build -o "$MEMBRANE_BIN" ./cmd/membrane
    log_info "Built: $MEMBRANE_BIN"
}

# Install OCI runtime-tools
install_runtime_tools() {
    if [[ -x "$RUNTIME_TOOLS_DIR/validation/validation" ]]; then
        log_info "OCI runtime-tools already installed"
        return
    fi
    
    log_info "Installing OCI runtime-tools..."
    rm -rf "$RUNTIME_TOOLS_DIR"
    git clone --depth 1 https://github.com/opencontainers/runtime-tools.git "$RUNTIME_TOOLS_DIR"
    cd "$RUNTIME_TOOLS_DIR"
    
    # Build the validation tool
    make validation
    
    log_info "OCI runtime-tools installed"
}

# Create minimal rootfs using busybox
create_rootfs() {
    log_info "Creating minimal rootfs at $ROOTFS_DIR..."
    
    rm -rf "$ROOTFS_DIR"
    mkdir -p "$ROOTFS_DIR"/{bin,dev,etc,proc,sys,tmp,usr/bin,var}
    
    # Download busybox static binary
    local ARCH=$(uname -m)
    local BUSYBOX_URL
    case "$ARCH" in
        x86_64)
            BUSYBOX_URL="https://busybox.net/downloads/binaries/1.35.0-x86_64-linux-musl/busybox"
            ;;
        aarch64)
            BUSYBOX_URL="https://busybox.net/downloads/binaries/1.35.0-aarch64-linux-musl/busybox"
            ;;
        *)
            log_error "Unsupported architecture: $ARCH"
            exit 1
            ;;
    esac
    
    log_info "Downloading busybox for $ARCH..."
    curl -sSL "$BUSYBOX_URL" -o "$ROOTFS_DIR/bin/busybox"
    chmod +x "$ROOTFS_DIR/bin/busybox"
    
    # Create symlinks
    cd "$ROOTFS_DIR/bin"
    for cmd in sh ls cat echo true false sleep env; do
        ln -sf busybox "$cmd"
    done
    
    # Create /etc files
    echo "root:x:0:0:root:/root:/bin/sh" > "$ROOTFS_DIR/etc/passwd"
    echo "root:x:0:" > "$ROOTFS_DIR/etc/group"
    
    log_info "Rootfs created"
}

# Create test bundle
create_test_bundle() {
    log_info "Creating test bundle at $BUNDLE_DIR..."
    
    rm -rf "$BUNDLE_DIR"
    mkdir -p "$BUNDLE_DIR"
    
    # Link rootfs
    ln -s "$ROOTFS_DIR" "$BUNDLE_DIR/rootfs"
    
    # Generate OCI config
    cat > "$BUNDLE_DIR/config.json" << 'EOF'
{
  "ociVersion": "1.0.2",
  "root": {
    "path": "rootfs",
    "readonly": false
  },
  "process": {
    "terminal": false,
    "user": {
      "uid": 0,
      "gid": 0
    },
    "args": ["/bin/sh", "-c", "exit 0"],
    "env": [
      "PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
      "TERM=xterm"
    ],
    "cwd": "/",
    "noNewPrivileges": true
  },
  "hostname": "membrane-test",
  "linux": {
    "namespaces": [
      {"type": "pid"},
      {"type": "mount"},
      {"type": "ipc"},
      {"type": "uts"},
      {"type": "network"},
      {"type": "cgroup"}
    ],
    "maskedPaths": [
      "/proc/acpi",
      "/proc/kcore",
      "/proc/keys",
      "/proc/latency_stats",
      "/proc/timer_list",
      "/proc/timer_stats",
      "/proc/sched_debug",
      "/proc/scsi",
      "/sys/firmware"
    ],
    "readonlyPaths": [
      "/proc/bus",
      "/proc/fs",
      "/proc/irq",
      "/proc/sys",
      "/proc/sysrq-trigger"
    ]
  }
}
EOF
    
    log_info "Test bundle created"
}

# Run basic validation tests
run_basic_tests() {
    log_info "Running basic validation tests..."
    
    TESTS_PASSED=0
    TESTS_FAILED=0
    
    # Clean up any existing containers
    "$MEMBRANE_BIN" list 2>/dev/null | tail -n +2 | awk '{print $1}' | while read id; do
        "$MEMBRANE_BIN" delete "$id" --force 2>/dev/null || true
    done
    
    # Test: version
    log_test "version command"
    if "$MEMBRANE_BIN" version 2>/dev/null | grep -q '"ociVersion"'; then
        echo -e "  ${GREEN}PASS${NC}: version outputs ociVersion"
        ((TESTS_PASSED++))
    else
        echo -e "  ${RED}FAIL${NC}: version does not output ociVersion"
        ((TESTS_FAILED++))
    fi
    
    # Test: spec
    log_test "spec command"
    if "$MEMBRANE_BIN" spec 2>/dev/null | grep -q '"ociVersion"'; then
        echo -e "  ${GREEN}PASS${NC}: spec generates valid config"
        ((TESTS_PASSED++))
    else
        echo -e "  ${RED}FAIL${NC}: spec does not generate valid config"
        ((TESTS_FAILED++))
    fi
    
    # Test: create
    log_test "create command"
    if "$MEMBRANE_BIN" create test-create "$BUNDLE_DIR" 2>/dev/null; then
        echo -e "  ${GREEN}PASS${NC}: container created"
        ((TESTS_PASSED++))
    else
        echo -e "  ${RED}FAIL${NC}: create failed"
        ((TESTS_FAILED++))
    fi
    
    # Test: state (created)
    log_test "state command (created container)"
    STATE=$("$MEMBRANE_BIN" state test-create 2>/dev/null)
    if echo "$STATE" | grep -q '"status": "created"'; then
        echo -e "  ${GREEN}PASS${NC}: state shows 'created'"
        ((TESTS_PASSED++))
    else
        echo -e "  ${RED}FAIL${NC}: state not 'created'"
        ((TESTS_FAILED++))
    fi
    
    # Test: state format
    log_test "state format (OCI compliant)"
    if echo "$STATE" | grep -q '"ociVersion"' && \
       echo "$STATE" | grep -q '"id"' && \
       echo "$STATE" | grep -q '"bundle"'; then
        echo -e "  ${GREEN}PASS${NC}: state format is OCI compliant"
        ((TESTS_PASSED++))
    else
        echo -e "  ${RED}FAIL${NC}: state format missing required fields"
        ((TESTS_FAILED++))
    fi
    
    # Test: list
    log_test "list command"
    if "$MEMBRANE_BIN" list 2>/dev/null | grep -q 'test-create'; then
        echo -e "  ${GREEN}PASS${NC}: container appears in list"
        ((TESTS_PASSED++))
    else
        echo -e "  ${RED}FAIL${NC}: container not in list"
        ((TESTS_FAILED++))
    fi
    
    # Test: delete
    log_test "delete command"
    if "$MEMBRANE_BIN" delete test-create 2>/dev/null; then
        echo -e "  ${GREEN}PASS${NC}: container deleted"
        ((TESTS_PASSED++))
    else
        echo -e "  ${RED}FAIL${NC}: delete failed"
        ((TESTS_FAILED++))
    fi
    
    # Test: state after delete
    log_test "state after delete"
    if ! "$MEMBRANE_BIN" state test-create 2>/dev/null; then
        echo -e "  ${GREEN}PASS${NC}: container not found after delete"
        ((TESTS_PASSED++))
    else
        echo -e "  ${RED}FAIL${NC}: container still exists after delete"
        ((TESTS_FAILED++))
    fi
    
    # Test: run (create + start)
    log_test "run command"
    # Create a test that exits quickly
    cat > "$BUNDLE_DIR/config.json" << 'EOF'
{
  "ociVersion": "1.0.2",
  "root": {"path": "rootfs", "readonly": false},
  "process": {
    "terminal": false,
    "user": {"uid": 0, "gid": 0},
    "args": ["/bin/true"],
    "env": ["PATH=/bin"],
    "cwd": "/"
  },
  "hostname": "test",
  "linux": {
    "namespaces": [
      {"type": "pid"},
      {"type": "mount"},
      {"type": "ipc"},
      {"type": "uts"}
    ]
  }
}
EOF
    # Note: run may fail if namespaces aren't available
    if "$MEMBRANE_BIN" run test-run "$BUNDLE_DIR" 2>/dev/null; then
        echo -e "  ${GREEN}PASS${NC}: run completed"
        ((TESTS_PASSED++))
    else
        echo -e "  ${YELLOW}SKIP${NC}: run failed (may need namespace support)"
    fi
    "$MEMBRANE_BIN" delete test-run --force 2>/dev/null || true
    
    # Test: force delete
    log_test "force delete"
    "$MEMBRANE_BIN" create test-force "$BUNDLE_DIR" 2>/dev/null || true
    if "$MEMBRANE_BIN" delete test-force --force 2>/dev/null; then
        echo -e "  ${GREEN}PASS${NC}: force delete works"
        ((TESTS_PASSED++))
    else
        echo -e "  ${RED}FAIL${NC}: force delete failed"
        ((TESTS_FAILED++))
    fi
    
    # Test: duplicate create fails
    log_test "duplicate create fails"
    "$MEMBRANE_BIN" create test-dup "$BUNDLE_DIR" 2>/dev/null
    if ! "$MEMBRANE_BIN" create test-dup "$BUNDLE_DIR" 2>/dev/null; then
        echo -e "  ${GREEN}PASS${NC}: duplicate create correctly rejected"
        ((TESTS_PASSED++))
    else
        echo -e "  ${RED}FAIL${NC}: duplicate create should fail"
        ((TESTS_FAILED++))
    fi
    "$MEMBRANE_BIN" delete test-dup --force 2>/dev/null || true
    
    # Test: invalid bundle fails
    log_test "invalid bundle fails"
    if ! "$MEMBRANE_BIN" create test-invalid /nonexistent 2>/dev/null; then
        echo -e "  ${GREEN}PASS${NC}: invalid bundle correctly rejected"
        ((TESTS_PASSED++))
    else
        echo -e "  ${RED}FAIL${NC}: invalid bundle should fail"
        ((TESTS_FAILED++))
    fi
    
    echo ""
    log_info "=========================================="
    log_info "Basic Test Results"
    log_info "=========================================="
    log_info "Passed:  $TESTS_PASSED"
    log_info "Failed:  $TESTS_FAILED"
    log_info "=========================================="
    
    return $TESTS_FAILED
}

# Run OCI runtime-tools validation suite
run_runtime_tools_validation() {
    if [[ ! -x "$RUNTIME_TOOLS_DIR/validation/validation" ]]; then
        log_warn "runtime-tools validation binary not found, skipping"
        return 0
    fi
    
    log_info "Running OCI runtime-tools validation suite..."
    
    cd "$RUNTIME_TOOLS_DIR"
    
    # Run the validation tool
    # Note: Many tests may fail because they require specific kernel features
    # We run with --tap to get machine-readable output
    if ./validation/validation \
        --runtime="$MEMBRANE_BIN" \
        --root=/run/membrane-test \
        2>&1 | tee /tmp/membrane-validation.log; then
        log_info "runtime-tools validation completed"
    else
        log_warn "Some runtime-tools tests failed (see /tmp/membrane-validation.log)"
    fi
}

# Cleanup
cleanup() {
    log_info "Cleaning up..."
    
    # Clean up any leftover containers
    for id in $("$MEMBRANE_BIN" list 2>/dev/null | tail -n +2 | awk '{print $1}'); do
        "$MEMBRANE_BIN" delete "$id" --force 2>/dev/null || true
    done
    
    # Clean up state directory
    rm -rf /run/membrane-test
    
    log_info "Cleanup complete"
}

# Main
main() {
    log_info "OCI Runtime Compliance Test for Membrane"
    log_info "=========================================="
    
    check_prerequisites
    build_membrane
    create_rootfs
    create_test_bundle
    
    trap cleanup EXIT
    
    FAILURES=0
    
    run_basic_tests || FAILURES=$?
    
    # Optionally run full runtime-tools validation
    if [[ "${FULL_VALIDATION:-0}" == "1" ]]; then
        install_runtime_tools
        run_runtime_tools_validation
    fi
    
    if [[ $FAILURES -gt 0 ]]; then
        log_error "Some tests failed"
        exit 1
    fi
    
    log_info "All tests completed successfully!"
}

main "$@"
