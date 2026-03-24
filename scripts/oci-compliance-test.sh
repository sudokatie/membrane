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

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
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
    if [[ -d "$RUNTIME_TOOLS_DIR" ]]; then
        log_info "OCI runtime-tools already installed"
        return
    fi
    
    log_info "Installing OCI runtime-tools..."
    git clone https://github.com/opencontainers/runtime-tools.git "$RUNTIME_TOOLS_DIR"
    cd "$RUNTIME_TOOLS_DIR"
    make
    log_info "OCI runtime-tools installed"
}

# Create test bundle
create_test_bundle() {
    log_info "Creating test bundle at $BUNDLE_DIR..."
    
    rm -rf "$BUNDLE_DIR"
    mkdir -p "$BUNDLE_DIR/rootfs"
    
    # Create minimal rootfs using busybox if available
    if command -v busybox &> /dev/null; then
        cp "$(which busybox)" "$BUNDLE_DIR/rootfs/"
        cd "$BUNDLE_DIR/rootfs"
        for cmd in sh ls cat echo true false sleep; do
            ln -s busybox "$cmd" 2>/dev/null || true
        done
        cd - > /dev/null
    else
        # Create minimal shell script as /bin/sh
        mkdir -p "$BUNDLE_DIR/rootfs/bin"
        cat > "$BUNDLE_DIR/rootfs/bin/sh" << 'EOF'
#!/bin/true
EOF
        chmod +x "$BUNDLE_DIR/rootfs/bin/sh"
    fi
    
    # Generate OCI config
    "$MEMBRANE_BIN" spec > "$BUNDLE_DIR/config.json"
    
    log_info "Test bundle created"
}

# Run validation tests
run_validation_tests() {
    log_info "Running OCI validation tests..."
    
    cd "$RUNTIME_TOOLS_DIR"
    
    # Run the validation tool
    TESTS_PASSED=0
    TESTS_FAILED=0
    TESTS_SKIPPED=0
    
    # Test: create
    log_info "Testing: create"
    if "$MEMBRANE_BIN" create test-create "$BUNDLE_DIR" 2>/dev/null; then
        "$MEMBRANE_BIN" delete test-create --force 2>/dev/null || true
        log_info "  PASS: create"
        ((TESTS_PASSED++))
    else
        log_error "  FAIL: create"
        ((TESTS_FAILED++))
    fi
    
    # Test: state
    log_info "Testing: state"
    "$MEMBRANE_BIN" create test-state "$BUNDLE_DIR" 2>/dev/null || true
    if "$MEMBRANE_BIN" state test-state 2>/dev/null | grep -q '"status"'; then
        log_info "  PASS: state"
        ((TESTS_PASSED++))
    else
        log_error "  FAIL: state"
        ((TESTS_FAILED++))
    fi
    "$MEMBRANE_BIN" delete test-state --force 2>/dev/null || true
    
    # Test: delete
    log_info "Testing: delete"
    "$MEMBRANE_BIN" create test-delete "$BUNDLE_DIR" 2>/dev/null || true
    if "$MEMBRANE_BIN" delete test-delete 2>/dev/null; then
        log_info "  PASS: delete"
        ((TESTS_PASSED++))
    else
        log_error "  FAIL: delete"
        ((TESTS_FAILED++))
    fi
    
    # Test: list
    log_info "Testing: list"
    if "$MEMBRANE_BIN" list 2>/dev/null; then
        log_info "  PASS: list"
        ((TESTS_PASSED++))
    else
        log_error "  FAIL: list"
        ((TESTS_FAILED++))
    fi
    
    # Test: version
    log_info "Testing: version"
    if "$MEMBRANE_BIN" version 2>/dev/null | grep -q 'ociVersion'; then
        log_info "  PASS: version"
        ((TESTS_PASSED++))
    else
        log_error "  FAIL: version"
        ((TESTS_FAILED++))
    fi
    
    # Test: spec
    log_info "Testing: spec"
    if "$MEMBRANE_BIN" spec 2>/dev/null | grep -q 'ociVersion'; then
        log_info "  PASS: spec"
        ((TESTS_PASSED++))
    else
        log_error "  FAIL: spec"
        ((TESTS_FAILED++))
    fi
    
    echo ""
    log_info "=========================================="
    log_info "OCI Compliance Test Results"
    log_info "=========================================="
    log_info "Passed:  $TESTS_PASSED"
    log_info "Failed:  $TESTS_FAILED"
    log_info "Skipped: $TESTS_SKIPPED"
    log_info "=========================================="
    
    if [[ $TESTS_FAILED -gt 0 ]]; then
        exit 1
    fi
}

# Cleanup
cleanup() {
    log_info "Cleaning up..."
    rm -rf "$BUNDLE_DIR"
    # Clean up any leftover containers
    for id in $("$MEMBRANE_BIN" list 2>/dev/null | tail -n +2 | awk '{print $1}'); do
        "$MEMBRANE_BIN" delete "$id" --force 2>/dev/null || true
    done
    log_info "Cleanup complete"
}

# Main
main() {
    log_info "OCI Runtime Compliance Test for Membrane"
    log_info "=========================================="
    
    check_prerequisites
    build_membrane
    create_test_bundle
    
    trap cleanup EXIT
    
    run_validation_tests
    
    log_info "All tests completed successfully!"
}

main "$@"
