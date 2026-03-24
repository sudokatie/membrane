// Package capabilities handles Linux capability management.
package capabilities

import (
	"fmt"
	"strings"

	"github.com/sudokatie/membrane/pkg/oci"
)

// Config holds capability configuration.
type Config struct {
	// Bounding is the bounding set.
	Bounding []string
	// Effective is the effective set.
	Effective []string
	// Inheritable is the inheritable set.
	Inheritable []string
	// Permitted is the permitted set.
	Permitted []string
	// Ambient is the ambient set.
	Ambient []string
}

// FromSpec creates a Config from an OCI Capabilities spec.
func FromSpec(caps *oci.Capabilities) *Config {
	if caps == nil {
		return nil
	}
	return &Config{
		Bounding:    caps.Bounding,
		Effective:   caps.Effective,
		Inheritable: caps.Inheritable,
		Permitted:   caps.Permitted,
		Ambient:     caps.Ambient,
	}
}

// DefaultConfig returns default capabilities for a container.
func DefaultConfig() *Config {
	// Minimal set of capabilities for a basic container
	defaultCaps := []string{
		"CAP_CHOWN",
		"CAP_DAC_OVERRIDE",
		"CAP_FSETID",
		"CAP_FOWNER",
		"CAP_MKNOD",
		"CAP_NET_RAW",
		"CAP_SETGID",
		"CAP_SETUID",
		"CAP_SETFCAP",
		"CAP_SETPCAP",
		"CAP_NET_BIND_SERVICE",
		"CAP_SYS_CHROOT",
		"CAP_KILL",
		"CAP_AUDIT_WRITE",
	}
	return &Config{
		Bounding:    defaultCaps,
		Effective:   defaultCaps,
		Inheritable: defaultCaps,
		Permitted:   defaultCaps,
		Ambient:     defaultCaps,
	}
}

// capabilityMap maps capability names to their bit values.
var capabilityMap = map[string]uint{
	"CAP_CHOWN":              0,
	"CAP_DAC_OVERRIDE":       1,
	"CAP_DAC_READ_SEARCH":    2,
	"CAP_FOWNER":             3,
	"CAP_FSETID":             4,
	"CAP_KILL":               5,
	"CAP_SETGID":             6,
	"CAP_SETUID":             7,
	"CAP_SETPCAP":            8,
	"CAP_LINUX_IMMUTABLE":    9,
	"CAP_NET_BIND_SERVICE":   10,
	"CAP_NET_BROADCAST":      11,
	"CAP_NET_ADMIN":          12,
	"CAP_NET_RAW":            13,
	"CAP_IPC_LOCK":           14,
	"CAP_IPC_OWNER":          15,
	"CAP_SYS_MODULE":         16,
	"CAP_SYS_RAWIO":          17,
	"CAP_SYS_CHROOT":         18,
	"CAP_SYS_PTRACE":         19,
	"CAP_SYS_PACCT":          20,
	"CAP_SYS_ADMIN":          21,
	"CAP_SYS_BOOT":           22,
	"CAP_SYS_NICE":           23,
	"CAP_SYS_RESOURCE":       24,
	"CAP_SYS_TIME":           25,
	"CAP_SYS_TTY_CONFIG":     26,
	"CAP_MKNOD":              27,
	"CAP_LEASE":              28,
	"CAP_AUDIT_WRITE":        29,
	"CAP_AUDIT_CONTROL":      30,
	"CAP_SETFCAP":            31,
	"CAP_MAC_OVERRIDE":       32,
	"CAP_MAC_ADMIN":          33,
	"CAP_SYSLOG":             34,
	"CAP_WAKE_ALARM":         35,
	"CAP_BLOCK_SUSPEND":      36,
	"CAP_AUDIT_READ":         37,
	"CAP_PERFMON":            38,
	"CAP_BPF":                39,
	"CAP_CHECKPOINT_RESTORE": 40,
}

// LastCap is the highest capability number.
const LastCap = 40

// ParseCapability parses a capability name to its number.
func ParseCapability(name string) (uint, error) {
	// Normalize name
	name = strings.ToUpper(name)
	if !strings.HasPrefix(name, "CAP_") {
		name = "CAP_" + name
	}

	if cap, ok := capabilityMap[name]; ok {
		return cap, nil
	}
	return 0, fmt.Errorf("unknown capability: %s", name)
}

// ToBitset converts a list of capability names to a bitset.
func ToBitset(caps []string) (uint64, error) {
	var bits uint64
	for _, c := range caps {
		cap, err := ParseCapability(c)
		if err != nil {
			return 0, err
		}
		bits |= (1 << cap)
	}
	return bits, nil
}
